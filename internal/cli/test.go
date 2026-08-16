package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"leetcli/internal/runner"
)

var (
	passStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	failStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	warnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

type metaFile struct {
	Slug string `json:"slug"`
	Lang string `json:"lang"`
}

func readMeta(path string) (metaFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return metaFile{}, errors.New("leetcli.json not found in this directory — run `leetcli init <slug>` first")
		}
		return metaFile{}, fmt.Errorf("cannot read %s: %w", path, err)
	}
	var m metaFile
	if err := json.Unmarshal(data, &m); err != nil {
		return metaFile{}, fmt.Errorf("cannot parse %s: %w", path, err)
	}
	if m.Slug == "" || m.Lang == "" {
		return metaFile{}, fmt.Errorf("%s is missing slug or lang", path)
	}
	return m, nil
}

func runTest(args []string) int {
	keep := false
	for _, a := range args {
		switch a {
		case "--keep":
			keep = true
		case "-h", "--help":
			fmt.Print(usage)
			return 0
		default:
			fmt.Fprintf(os.Stderr, "unknown argument %q\n", a)
			return 2
		}
	}
	meta, err := readMeta("leetcli.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	r, ok := runner.For(meta.Lang)
	if !ok {
		fmt.Fprintf(os.Stderr, "leetcli.json: unsupported language %q\n", meta.Lang)
		return 1
	}
	solutionPath := "solution." + r.Extension()
	data, err := os.ReadFile(solutionPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s not found — run `leetcli init %s` first\n", solutionPath, meta.Slug)
		return 1
	}
	if !r.Available() {
		fmt.Fprintf(os.Stderr, "%s toolchain not found on PATH\n", meta.Lang)
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	detail, err := fetchProblem(ctx, meta.Slug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	task, err := runner.NewTaskForLang(detail, meta.Lang)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	task.Solution = string(data)
	workDir := ""
	if keep {
		dir, err := os.MkdirTemp("", "leetcli-"+meta.Slug+"-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		workDir = dir
	}
	results, err := r.Run(ctx, task, workDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, formatRunError(err))
		return 1
	}
	fmt.Printf("%s [%s] sample case\n", meta.Slug, meta.Lang)
	passed := 0
	failed := 0
	for _, res := range results {
		n := res.Case + 1
		switch res.Status {
		case runner.StatusPass:
			passed++
			fmt.Printf("case %d  %s  got=%s\n", n, passStyle.Render("PASS"), oneLine(res.Got))
		case runner.StatusNoExpect:
			passed++
			fmt.Printf("case %d  %s  no expected output to compare\n", n, warnStyle.Render("OK"))
		case runner.StatusFail:
			failed++
			fmt.Printf("case %d  %s\n", n, failStyle.Render("FAIL"))
			fmt.Printf("  want: %s\n", oneLine(res.Want))
			fmt.Printf("  got:  %s\n", oneLine(res.Got))
		case runner.StatusRuntime:
			failed++
			fmt.Printf("case %d  %s\n", n, warnStyle.Render("ERROR"))
			for _, line := range strings.Split(strings.TrimRight(res.Message, "\n"), "\n") {
				fmt.Printf("  %s\n", line)
			}
		}
	}
	if failed > 0 {
		fmt.Println(failStyle.Render(fmt.Sprintf("failed %d/%d", failed, len(results))))
	} else {
		fmt.Println(passStyle.Render(fmt.Sprintf("passed %d/%d", passed, len(results))))
	}
	if workDir != "" {
		fmt.Printf("generated sources: %s\n", workDir)
	}
	if failed > 0 {
		return 1
	}
	return 0
}

func formatRunError(err error) string {
	var be *runner.BuildError
	if errors.As(err, &be) {
		return be.Error()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timed out"
	}
	return err.Error()
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	return truncate(s, 300)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + "..."
}
