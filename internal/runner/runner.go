package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"leetcli/internal/domain"
)

type Case struct {
	Inputs   []any
	Expected string
}

type Task struct {
	Slug     string
	Lang     string
	Solution string
	Sig      domain.Signature
	Cases    []Case
}

type Status string

const (
	StatusPass     Status = "pass"
	StatusFail     Status = "fail"
	StatusRuntime  Status = "runtime"
	StatusNoExpect Status = "ran"
)

type Result struct {
	Case    int
	Status  Status
	Got     string
	Want    string
	Message string
}

type BuildError struct {
	Output string
}

func (e *BuildError) Error() string {
	return "build failed:\n" + e.Output
}

type Runner interface {
	Name() string
	Extension() string
	Available() bool
	Scaffold(task Task, dir string) (string, error)
	Run(ctx context.Context, task Task, workDir string) ([]Result, error)
}

var registry = map[string]Runner{}

func Register(r Runner) {
	registry[r.Name()] = r
}

func For(lang string) (Runner, bool) {
	r, ok := registry[lang]
	return r, ok
}

func Languages() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

var leetSlugs = map[string]string{
	"golang":     "go",
	"python":     "python",
	"python3":    "python",
	"pypy3":      "python",
	"javascript": "javascript",
	"cpp":        "cpp",
}

func FromLeetCodeSlug(slug string) (string, error) {
	if lang, ok := leetSlugs[slug]; ok {
		return lang, nil
	}
	return "", fmt.Errorf("local tests for %q are not supported yet (supported: %s)", slug, strings.Join(Languages(), ", "))
}

func NewTask(detail domain.ProblemDetail, leetSlug string) (Task, error) {
	lang, err := FromLeetCodeSlug(leetSlug)
	if err != nil {
		return Task{}, err
	}
	snip, ok := detail.Snippet(leetSlug)
	if !ok {
		return Task{}, fmt.Errorf("no code snippet for %q", leetSlug)
	}
	return taskFrom(detail, lang, snip.Code)
}

func NewTaskForLang(detail domain.ProblemDetail, lang string) (Task, error) {
	if _, ok := For(lang); !ok {
		return Task{}, fmt.Errorf("unsupported language %q (supported: %s)", lang, strings.Join(Languages(), ", "))
	}
	return taskFrom(detail, lang, "")
}

func taskFrom(detail domain.ProblemDetail, lang, solution string) (Task, error) {
	if detail.Signature.Name == "" {
		return Task{}, errors.New("problem signature is missing")
	}
	inputs, err := ParseArgs(detail.SampleInput, len(detail.Signature.Params))
	if err != nil {
		return Task{}, err
	}
	if len(inputs) == 0 {
		return Task{}, errors.New("problem has no sample test case")
	}
	expected := ""
	for _, ex := range detail.Examples {
		if strings.TrimSpace(ex.Output) != "" {
			expected = strings.TrimSpace(ex.Output)
			break
		}
	}
	for _, p := range detail.Signature.Params {
		if !supportedType(p.Type) {
			return Task{}, fmt.Errorf("parameter %q has unsupported type %q", p.Name, p.Type)
		}
	}
	if ret := detail.Signature.Return; ret != "" && ret != "void" && !supportedType(ret) {
		return Task{}, fmt.Errorf("unsupported return type %q", ret)
	}
	return Task{
		Slug:     detail.Slug,
		Lang:     lang,
		Solution: solution,
		Sig:      detail.Signature,
		Cases:    []Case{{Inputs: inputs, Expected: expected}},
	}, nil
}

var langAliases = map[string]string{
	"go":         "golang",
	"golang":     "golang",
	"py":         "python3",
	"python":     "python3",
	"python3":    "python3",
	"pypy3":      "python3",
	"js":         "javascript",
	"javascript": "javascript",
	"cpp":        "cpp",
	"c++":        "cpp",
}

func ResolveLang(arg string) (string, string, error) {
	leet, ok := langAliases[strings.ToLower(strings.TrimSpace(arg))]
	if !ok {
		return "", "", fmt.Errorf("unsupported language %q (use: golang, python3, javascript, cpp)", arg)
	}
	lang, err := FromLeetCodeSlug(leet)
	if err != nil {
		return "", "", err
	}
	return lang, leet, nil
}

func ParseArgs(s string, n int) ([]any, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var out []any
	for {
		var v any
		if err := dec.Decode(&v); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("cannot parse sample test case: %w", err)
		}
		out = append(out, v)
	}
	if n > 0 && len(out) != n {
		return nil, fmt.Errorf("sample test case has %d values, expected %d", len(out), n)
	}
	return out, nil
}

type program struct {
	Files map[string]string
	Build []string
	Exec  []string
}

func runTask(ctx context.Context, r Runner, task Task, workDir string) ([]Result, error) {
	prog, err := generate(r, task)
	if err != nil {
		return nil, err
	}
	dir := workDir
	if dir == "" {
		dir, err = os.MkdirTemp("", "leetcli-run-")
		if err != nil {
			return nil, err
		}
		defer func() { _ = os.RemoveAll(dir) }()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	for name, src := range prog.Files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			return nil, err
		}
	}
	ctx, cancel := runContext(ctx)
	defer cancel()
	if len(prog.Build) > 0 {
		out, berr := commandCombined(ctx, dir, prog.Build)
		if berr != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return nil, context.DeadlineExceeded
			}
			if strings.TrimSpace(out) == "" {
				out = berr.Error()
			}
			return nil, &BuildError{Output: strings.TrimSpace(out)}
		}
	}
	stdout, stderr, runErr := commandPair(ctx, dir, prog.Exec)
	if runErr != nil && ctx.Err() == context.DeadlineExceeded {
		runErr = context.DeadlineExceeded
	}
	return assemble(task, stdout, stderr, runErr), nil
}

func runContext(parent context.Context) (context.Context, context.CancelFunc) {
	if _, ok := parent.Deadline(); ok {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, 30*time.Second)
}

func commandCombined(ctx context.Context, dir string, argv []string) (string, error) {
	if len(argv) == 0 {
		return "", nil
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

func commandPair(ctx context.Context, dir string, argv []string) (string, string, error) {
	if len(argv) == 0 {
		return "", "", errors.New("no command to run")
	}
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return out.String(), errBuf.String(), err
}

type caseOut struct {
	status  string
	payload string
}

func assemble(task Task, stdout, stderr string, runErr error) []Result {
	parsed := parseCases(stdout)
	void := strings.EqualFold(task.Sig.Return, "void")
	results := make([]Result, 0, len(task.Cases))
	for i, c := range task.Cases {
		want := strings.TrimSpace(c.Expected)
		co, ok := parsed[i]
		switch {
		case ok && co.status == "OK":
			res := Result{Case: i, Got: co.payload, Want: want}
			switch {
			case void:
				res.Status = StatusPass
			case want == "":
				res.Status = StatusNoExpect
			case outputsMatch(co.payload, want):
				res.Status = StatusPass
			default:
				res.Status = StatusFail
			}
			results = append(results, res)
		case ok:
			results = append(results, Result{
				Case:    i,
				Status:  StatusRuntime,
				Want:    want,
				Message: co.payload,
			})
		default:
			var msg strings.Builder
			switch {
			case errors.Is(runErr, context.DeadlineExceeded):
				msg.WriteString("timeout")
			case runErr != nil:
				msg.WriteString("program terminated: " + runErr.Error())
			default:
				msg.WriteString("program produced no result")
			}
			if t := strings.TrimSpace(stderr); t != "" {
				msg.WriteString("\n" + tail(t, 2000))
			}
			results = append(results, Result{
				Case:    i,
				Status:  StatusRuntime,
				Want:    want,
				Message: msg.String(),
			})
		}
	}
	return results
}

func parseCases(stdout string) map[int]caseOut {
	out := map[int]caseOut{}
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if !strings.HasPrefix(line, "CASE ") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(line, "CASE "), "\t", 3)
		if len(parts) < 2 {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}
		payload := ""
		if len(parts) == 3 {
			payload = parts[2]
		}
		out[idx] = caseOut{status: parts[1], payload: payload}
	}
	return out
}

func outputsMatch(got, want string) bool {
	g, w := normalize(got), normalize(want)
	if g == w {
		return true
	}
	if isIntLiteral(g) && isIntLiteral(w) {
		return false
	}
	gf, gerr := strconv.ParseFloat(g, 64)
	wf, werr := strconv.ParseFloat(w, 64)
	if gerr != nil || werr != nil {
		return false
	}
	diff := math.Abs(gf - wf)
	return diff <= 1e-6*math.Max(1, math.Max(math.Abs(gf), math.Abs(wf)))
}

func isIntLiteral(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	if s[0] == '-' {
		i = 1
	}
	if i == len(s) {
		return false
	}
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func normalize(s string) string {
	var b strings.Builder
	inQuote := false
	escaped := false
	for _, r := range s {
		switch {
		case escaped:
			b.WriteRune(r)
			escaped = false
		case r == '\\' && inQuote:
			b.WriteRune(r)
			escaped = true
		case r == '"':
			inQuote = !inQuote
			b.WriteRune(r)
		case !inQuote && unicode.IsSpace(r):
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	s = s[len(s)-n:]
	for len(s) > 0 && !utf8.RuneStart(s[0]) {
		s = s[1:]
	}
	return "..." + s
}

func generate(r Runner, task Task) (program, error) {
	switch r.Name() {
	case "go":
		return goProgram(task)
	case "python":
		return pythonProgram(task)
	case "javascript":
		return jsProgram(task)
	case "cpp":
		return cppProgram(task)
	}
	return program{}, fmt.Errorf("no generator for language %q", r.Name())
}

func scaffoldContent(lang string, task Task) (string, error) {
	switch lang {
	case "go":
		return "package main\n\n" + stripPackage(task.Solution), nil
	case "python", "javascript":
		return ensureNewline(task.Solution), nil
	case "cpp":
		return cppHeader() + ensureNewline(task.Solution), nil
	}
	return "", fmt.Errorf("no scaffold for language %q", lang)
}

func writeScaffold(r Runner, task Task, dir string) (string, error) {
	content, err := scaffoldContent(r.Name(), task)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "solution."+r.Extension())
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	meta := fmt.Sprintf("{\"slug\": %q, \"lang\": %q}\n", task.Slug, r.Name())
	if err := os.WriteFile(filepath.Join(dir, "leetcli.json"), []byte(meta), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
