package runner

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"leetcli/internal/domain"
)

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func mustParse(t *testing.T, s string, n int) []any {
	t.Helper()
	args, err := ParseArgs(s, n)
	if err != nil {
		t.Fatalf("ParseArgs(%q): %v", s, err)
	}
	return args
}

func TestParseArgs(t *testing.T) {
	args := mustParse(t, "[2,7,11,15]\n9", 2)
	if len(args) != 2 {
		t.Fatalf("want 2 values, got %d", len(args))
	}
	if _, ok := args[0].([]any); !ok {
		t.Fatalf("first value should be array, got %T", args[0])
	}
	if n, ok := args[1].(json.Number); !ok || n.String() != "9" {
		t.Fatalf("second value should be 9, got %v", args[1])
	}
}

func TestParseArgsStringAndFloat(t *testing.T) {
	args := mustParse(t, "\"abc\"\n2.5\ntrue", 3)
	if args[0] != "abc" {
		t.Fatalf("want abc, got %v", args[0])
	}
	if n, ok := args[1].(json.Number); !ok || n.String() != "2.5" {
		t.Fatalf("want 2.5, got %v", args[1])
	}
	if args[2] != true {
		t.Fatalf("want true, got %v", args[2])
	}
}

func TestParseArgsCountMismatch(t *testing.T) {
	if _, err := ParseArgs("[1,2]", 3); err == nil {
		t.Fatal("expected count mismatch error")
	}
}

func TestParseArgsEmpty(t *testing.T) {
	args, err := ParseArgs("  \n ", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args != nil {
		t.Fatalf("want nil, got %v", args)
	}
}

func TestParseArgsInvalid(t *testing.T) {
	if _, err := ParseArgs("[1,2", 2); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLitPrimitives(t *testing.T) {
	cases := []struct {
		lang string
		typ  string
		val  any
		want string
	}{
		{"go", "integer", json.Number("9"), "9"},
		{"python", "integer", json.Number("9"), "9"},
		{"cpp", "integer", json.Number("9"), "9"},
		{"go", "double", json.Number("2.5"), "2.5"},
		{"python", "boolean", true, "True"},
		{"javascript", "boolean", false, "false"},
		{"go", "string", "ab", `"ab"`},
		{"python", "string", "a\"b", `"a\"b"`},
		{"cpp", "string", "a\nb", `"a\nb"`},
		{"go", "character", "x", `'x'`},
		{"cpp", "character", "'", `'\''`},
	}
	for _, c := range cases {
		got, err := lit(c.lang, c.typ, c.val)
		if err != nil {
			t.Fatalf("lit(%s, %s): %v", c.lang, c.typ, err)
		}
		if got != c.want {
			t.Errorf("lit(%s, %s, %v) = %q, want %q", c.lang, c.typ, c.val, got, c.want)
		}
	}
}

func TestLitArrays(t *testing.T) {
	nums := []any{json.Number("1"), json.Number("2")}
	nested := []any{[]any{json.Number("1"), json.Number("2")}, []any{json.Number("3")}}
	cases := []struct {
		lang string
		typ  string
		val  any
		want string
	}{
		{"go", "integer[]", nums, "[]int{1, 2}"},
		{"python", "integer[]", nums, "[1, 2]"},
		{"javascript", "integer[]", nums, "[1, 2]"},
		{"cpp", "integer[]", nums, "std::vector<int>{1, 2}"},
		{"go", "integer[][]", nested, "[][]int{[]int{1, 2}, []int{3}}"},
		{"python", "string[]", []any{"a", "b"}, `["a", "b"]`},
		{"go", "integer[]", nil, "[]int{}"},
	}
	for _, c := range cases {
		got, err := lit(c.lang, c.typ, c.val)
		if err != nil {
			t.Fatalf("lit(%s, %s): %v", c.lang, c.typ, err)
		}
		if got != c.want {
			t.Errorf("lit(%s, %s) = %q, want %q", c.lang, c.typ, got, c.want)
		}
	}
}

func TestLitListsAndTrees(t *testing.T) {
	list := []any{json.Number("1"), json.Number("2"), json.Number("3")}
	tree := []any{json.Number("1"), nil, json.Number("2")}
	cases := []struct {
		lang string
		typ  string
		val  any
		want string
	}{
		{"go", "ListNode", list, "lcMakeList([]int{1, 2, 3})"},
		{"python", "ListNode", list, "lc_make_list([1, 2, 3])"},
		{"javascript", "ListNode", list, "lcMakeList([1, 2, 3])"},
		{"cpp", "ListNode", list, "lcMakeList(std::vector<int>{1, 2, 3})"},
		{"go", "TreeNode", tree, "lcMakeTree([]any{1, nil, 2})"},
		{"python", "TreeNode", tree, "lc_make_tree([1, None, 2])"},
		{"javascript", "TreeNode", tree, "lcMakeTree([1, null, 2])"},
		{"cpp", "TreeNode", tree, "lcMakeTree(std::vector<std::optional<long long>>{1, std::nullopt, 2})"},
	}
	for _, c := range cases {
		got, err := lit(c.lang, c.typ, c.val)
		if err != nil {
			t.Fatalf("lit(%s, %s): %v", c.lang, c.typ, err)
		}
		if got != c.want {
			t.Errorf("lit(%s, %s) = %q, want %q", c.lang, c.typ, got, c.want)
		}
	}
}

func TestSupportedType(t *testing.T) {
	for _, typ := range []string{"integer", "integer[]", "integer[][]", "string", "double", "boolean", "character", "ListNode", "TreeNode", "string[]"} {
		if !supportedType(typ) {
			t.Errorf("%q should be supported", typ)
		}
	}
	for _, typ := range []string{"MountainArray", "integer[][][]ListNode", ""} {
		if supportedType(typ) {
			t.Errorf("%q should not be supported", typ)
		}
	}
}

func TestCallExpr(t *testing.T) {
	tasks := []struct {
		lang     string
		solution string
		args     []string
		want     string
	}{
		{"go", "func twoSum() {}", []string{"[]int{1}", "2"}, "twoSum([]int{1}, 2)"},
		{"python", "class Solution:\n    def twoSum(self): pass", []string{"[1]", "2"}, "Solution().twoSum([1], 2)"},
		{"python", "def twoSum(nums): pass", []string{"[1]", "2"}, "twoSum([1], 2)"},
		{"javascript", "class Solution {\n  twoSum() {}\n}", []string{"[1]", "2"}, "new Solution().twoSum([1], 2)"},
		{"javascript", "var twoSum = function(nums) {}", []string{"[1]", "2"}, "twoSum([1], 2)"},
		{"cpp", "class Solution {\npublic:\n  vector<int> twoSum() {}\n};", []string{"p0", "p1"}, "Solution().twoSum(p0, p1)"},
	}
	for _, c := range tasks {
		task := Task{
			Lang:     c.lang,
			Solution: c.solution,
			Sig:      domain.Signature{Name: "twoSum"},
		}
		got, err := callExpr(task, c.args)
		if err != nil {
			t.Fatalf("callExpr(%s): %v", c.lang, err)
		}
		if got != c.want {
			t.Errorf("callExpr(%s) = %q, want %q", c.lang, got, c.want)
		}
	}
}

func TestCallExprInvalidName(t *testing.T) {
	task := Task{Lang: "go", Sig: domain.Signature{Name: "1bad"}}
	if _, err := callExpr(task, nil); err == nil {
		t.Fatal("expected error for invalid function name")
	}
}

func TestOutputsMatch(t *testing.T) {
	cases := []struct {
		got  string
		want string
		ok   bool
	}{
		{"[0,1]", "[0,1]", true},
		{"[0, 1]", "[0,1]", true},
		{"  4 ", "4", true},
		{"4", "4.0", true},
		{"2.0000001", "2", true},
		{"[1,0]", "[0,1]", false},
		{"abc", "abcd", false},
		{"0.333333333333", "0.3333333333333333", true},
		{"1234567", "1234568", false},
		{"1000000", "1000001", false},
		{"-42", "-42", true},
		{"21", "21.0", true},
		{"9007199254740993", "9007199254740994", false},
	}
	for _, c := range cases {
		if got := outputsMatch(c.got, c.want); got != c.ok {
			t.Errorf("outputsMatch(%q, %q) = %v, want %v", c.got, c.want, got, c.ok)
		}
	}
}

func TestTailUTF8(t *testing.T) {
	s := strings.Repeat("олень", 500)
	got := tail(s, 10)
	if !utf8.ValidString(got) {
		t.Fatalf("tail result is not valid UTF-8: %q", got)
	}
	if !strings.HasPrefix(got, "...") {
		t.Fatalf("tail should start with ellipsis: %q", got)
	}
	if !strings.HasSuffix(s, strings.TrimPrefix(got, "...")) {
		t.Fatalf("tail should end with the original text")
	}
	if long := tail("short", 100); long != "short" {
		t.Fatalf("short string should pass through: %q", long)
	}
}

func TestAssembleStatuses(t *testing.T) {
	task := Task{
		Sig:   domain.Signature{Name: "f", Return: "integer[]"},
		Cases: []Case{{Expected: "[0,1]"}, {Expected: "[0,1]"}, {Expected: ""}},
	}
	stdout := "debug line\nCASE 0\tOK\t[0,1]\nCASE 1\tOK\t[1,0]\nCASE 2\tOK\t42\n"
	results := assemble(task, stdout, "", nil)
	if len(results) != 3 {
		t.Fatalf("want 3 results, got %d", len(results))
	}
	if results[0].Status != StatusPass {
		t.Errorf("case 0 = %s, want pass", results[0].Status)
	}
	if results[1].Status != StatusFail {
		t.Errorf("case 1 = %s, want fail", results[1].Status)
	}
	if results[2].Status != StatusNoExpect {
		t.Errorf("case 2 = %s, want ran", results[2].Status)
	}
}

func TestAssembleRuntimeAndTimeout(t *testing.T) {
	task := Task{
		Sig:   domain.Signature{Name: "f", Return: "integer"},
		Cases: []Case{{Expected: "5"}, {Expected: "5"}},
	}
	stdout := "CASE 0\tERR\tValueError: boom\n"
	results := assemble(task, stdout, "", nil)
	if results[0].Status != StatusRuntime || !strings.Contains(results[0].Message, "ValueError") {
		t.Errorf("case 0 = %s / %q", results[0].Status, results[0].Message)
	}
	if results[1].Status != StatusRuntime {
		t.Errorf("case 1 = %s, want runtime", results[1].Status)
	}

	results = assemble(task, "", "", context.DeadlineExceeded)
	if results[0].Status != StatusRuntime || results[0].Message != "timeout" {
		t.Errorf("timeout result = %s / %q", results[0].Status, results[0].Message)
	}
}

func TestAssembleStderrIncluded(t *testing.T) {
	task := Task{
		Sig:   domain.Signature{Name: "f", Return: "integer"},
		Cases: []Case{{Expected: "1"}},
	}
	results := assemble(task, "", "some fatal error", errors.New("exit status 1"))
	if results[0].Status != StatusRuntime {
		t.Fatalf("status = %s", results[0].Status)
	}
	if !strings.Contains(results[0].Message, "some fatal error") {
		t.Errorf("stderr not included: %q", results[0].Message)
	}
}

func TestFromLeetCodeSlug(t *testing.T) {
	lang, err := FromLeetCodeSlug("golang")
	if err != nil || lang != "go" {
		t.Fatalf("golang -> %q, %v", lang, err)
	}
	lang, err = FromLeetCodeSlug("python3")
	if err != nil || lang != "python" {
		t.Fatalf("python3 -> %q, %v", lang, err)
	}
	if _, err := FromLeetCodeSlug("java"); err == nil {
		t.Fatal("java should be unsupported")
	}
}

func TestRegistry(t *testing.T) {
	got := Languages()
	want := []string{"cpp", "go", "javascript", "python"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("Languages() = %v, want %v", got, want)
	}
	for _, lang := range want {
		if _, ok := For(lang); !ok {
			t.Errorf("runner for %q not registered", lang)
		}
	}
	if _, ok := For("rust"); ok {
		t.Error("rust should not be registered")
	}
}

func TestNewTask(t *testing.T) {
	detail := domain.ProblemDetail{
		Slug: "two-sum",
		Signature: domain.Signature{
			Name:   "twoSum",
			Params: []domain.Param{{Name: "nums", Type: "integer[]"}, {Name: "target", Type: "integer"}},
			Return: "integer[]",
		},
		Snippets: []domain.CodeSnippet{
			{LangSlug: "golang", Code: "func twoSum() {}"},
			{LangSlug: "python3", Code: "class Solution: pass"},
		},
		SampleInput: "[2,7,11,15]\n9",
		Examples:    []domain.Example{{Number: 1, Output: "[0,1]"}},
	}
	task, err := NewTask(detail, "golang")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if task.Lang != "go" {
		t.Errorf("Lang = %q, want go", task.Lang)
	}
	if task.Solution != "func twoSum() {}" {
		t.Errorf("Solution = %q", task.Solution)
	}
	if len(task.Cases) != 1 || task.Cases[0].Expected != "[0,1]" {
		t.Errorf("cases = %+v", task.Cases)
	}
	if len(task.Cases[0].Inputs) != 2 {
		t.Errorf("inputs = %v", task.Cases[0].Inputs)
	}

	if _, err := NewTask(detail, "java"); err == nil {
		t.Error("java should fail")
	}
	noSnippet := detail
	noSnippet.Snippets = nil
	if _, err := NewTask(noSnippet, "golang"); err == nil {
		t.Error("missing snippet should fail")
	}
	badType := detail
	badType.Signature.Params = []domain.Param{{Name: "arr", Type: "MountainArray"}}
	if _, err := NewTask(badType, "golang"); err == nil {
		t.Error("unsupported param type should fail")
	}
}

func TestNewTaskForLang(t *testing.T) {
	detail := domain.ProblemDetail{
		Slug: "two-sum",
		Signature: domain.Signature{
			Name:   "twoSum",
			Params: []domain.Param{{Name: "nums", Type: "integer[]"}, {Name: "target", Type: "integer"}},
			Return: "integer[]",
		},
		SampleInput: "[2,7,11,15]\n9",
		Examples:    []domain.Example{{Number: 1, Output: "[0,1]"}},
	}
	task, err := NewTaskForLang(detail, "go")
	if err != nil {
		t.Fatalf("NewTaskForLang: %v", err)
	}
	if task.Lang != "go" {
		t.Errorf("Lang = %q, want go", task.Lang)
	}
	if task.Solution != "" {
		t.Errorf("Solution = %q, want empty", task.Solution)
	}
	if len(task.Cases) != 1 || task.Cases[0].Expected != "[0,1]" {
		t.Errorf("cases = %+v", task.Cases)
	}
	if _, err := NewTaskForLang(detail, "rust"); err == nil {
		t.Error("rust should fail")
	}
	noSig := detail
	noSig.Signature = domain.Signature{}
	if _, err := NewTaskForLang(noSig, "go"); err == nil {
		t.Error("missing signature should fail")
	}
}

func TestResolveLang(t *testing.T) {
	cases := []struct {
		arg     string
		lang    string
		leet    string
		wantErr bool
	}{
		{"golang", "go", "golang", false},
		{"go", "go", "golang", false},
		{"GOLANG", "go", "golang", false},
		{"python3", "python", "python3", false},
		{"python", "python", "python3", false},
		{"py", "python", "python3", false},
		{"javascript", "javascript", "javascript", false},
		{"js", "javascript", "javascript", false},
		{"cpp", "cpp", "cpp", false},
		{"c++", "cpp", "cpp", false},
		{"java", "", "", true},
		{"typescript", "", "", true},
		{"", "", "", true},
	}
	for _, c := range cases {
		lang, leet, err := ResolveLang(c.arg)
		if c.wantErr {
			if err == nil {
				t.Errorf("ResolveLang(%q) should fail", c.arg)
			}
			continue
		}
		if err != nil {
			t.Errorf("ResolveLang(%q): %v", c.arg, err)
			continue
		}
		if lang != c.lang || leet != c.leet {
			t.Errorf("ResolveLang(%q) = (%q, %q), want (%q, %q)", c.arg, lang, leet, c.lang, c.leet)
		}
	}
}

func TestScaffoldFiles(t *testing.T) {
	task := Task{
		Slug:     "two-sum",
		Lang:     "go",
		Solution: "func twoSum() {}",
		Sig:      domain.Signature{Name: "twoSum"},
	}
	for _, lang := range Languages() {
		r, _ := For(lang)
		dir := t.TempDir()
		path, err := r.Scaffold(task, dir)
		if err != nil {
			t.Fatalf("scaffold %s: %v", lang, err)
		}
		if !strings.HasSuffix(path, "."+r.Extension()) {
			t.Errorf("path = %q, want extension .%s", path, r.Extension())
		}
		data, err := readFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		switch lang {
		case "go":
			if !strings.HasPrefix(data, "package main") {
				t.Errorf("go scaffold should start with package main: %q", data)
			}
		case "cpp":
			if !strings.Contains(data, "#include") {
				t.Errorf("cpp scaffold should contain includes: %q", data)
			}
		default:
			if data != "func twoSum() {}\n" {
				t.Errorf("%s scaffold = %q", lang, data)
			}
		}
		meta, err := readFile(filepath.Join(dir, "leetcli.json"))
		if err != nil {
			t.Fatalf("meta for %s: %v", lang, err)
		}
		if !strings.Contains(meta, `"slug": "two-sum"`) {
			t.Errorf("meta = %q", meta)
		}
	}
}

func TestGenerateAllLanguages(t *testing.T) {
	inputs := mustParse(t, "[2,7,11,15]\n9", 2)
	task := Task{
		Slug:     "two-sum",
		Solution: "class Solution {}",
		Sig: domain.Signature{
			Name:   "twoSum",
			Params: []domain.Param{{Name: "nums", Type: "integer[]"}, {Name: "target", Type: "integer"}},
			Return: "integer[]",
		},
		Cases: []Case{{Inputs: inputs, Expected: "[0,1]"}},
	}
	for _, lang := range Languages() {
		r, _ := For(lang)
		task.Lang = lang
		prog, err := generate(r, task)
		if err != nil {
			t.Fatalf("generate %s: %v", lang, err)
		}
		if len(prog.Files) == 0 || len(prog.Exec) == 0 {
			t.Errorf("generate %s: empty program %+v", lang, prog)
		}
		for name, src := range prog.Files {
			if !strings.Contains(src, "CASE ") {
				t.Errorf("generate %s: %s missing case output", lang, name)
			}
		}
	}
}

func TestGenerateListNodeConditional(t *testing.T) {
	inputs := mustParse(t, "[1,2,3]", 1)
	task := Task{
		Lang:     "go",
		Solution: "type ListNode struct{}\nfunc f() {}",
		Sig: domain.Signature{
			Name:   "f",
			Params: []domain.Param{{Name: "head", Type: "ListNode"}},
			Return: "integer[]",
		},
		Cases: []Case{{Inputs: inputs}},
	}
	r, _ := For("go")
	prog, err := generate(r, task)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src := prog.Files["run.go"]
	if !strings.Contains(src, "*ListNode") || !strings.Contains(src, "lcMakeList") {
		t.Error("ListNode helpers should be emitted when used")
	}

	task.Sig.Params = []domain.Param{{Name: "n", Type: "integer"}}
	task.Cases = []Case{{Inputs: []any{json.Number("1")}}}
	prog, err = generate(r, task)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	src = prog.Files["run.go"]
	if strings.Contains(src, "lcMakeList") || strings.Contains(src, "*ListNode") {
		t.Error("ListNode helpers should not be emitted when unused")
	}
}

func TestStripPackage(t *testing.T) {
	got := stripPackage("package main\n\nfunc f() {}\n")
	if got != "\nfunc f() {}\n" {
		t.Errorf("stripPackage = %q", got)
	}
	raw := "func f() {}\n"
	if stripPackage(raw) != raw {
		t.Errorf("raw snippet should stay unchanged")
	}
}

func TestArgSpecsCountMismatch(t *testing.T) {
	task := Task{
		Lang: "go",
		Sig:  domain.Signature{Name: "f", Params: []domain.Param{{Name: "a", Type: "integer"}}},
	}
	if _, _, err := argSpecs(task, nil); err == nil {
		t.Fatal("expected count mismatch error")
	}
}
