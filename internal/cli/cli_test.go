package cli

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"leetcli/internal/domain"
	"leetcli/internal/runner"
)

func twoSumDetail() domain.ProblemDetail {
	return domain.ProblemDetail{
		ID:         1,
		Slug:       "two-sum",
		Title:      "Two Sum",
		Difficulty: "Easy",
		Signature: domain.Signature{
			Name:   "twoSum",
			Params: []domain.Param{{Name: "nums", Type: "integer[]"}, {Name: "target", Type: "integer"}},
			Return: "integer[]",
		},
		Snippets: []domain.CodeSnippet{
			{LangSlug: "python3", Code: "class Solution:\n    def twoSum(self, nums, target):\n        pass\n"},
			{LangSlug: "golang", Code: "func twoSum(nums []int, target int) []int {\n}\n"},
		},
		SampleInput: "[2,7,11,15]\n9",
		Examples:    []domain.Example{{Number: 1, Output: "[0,1]"}},
	}
}

func stubFetch(t *testing.T) {
	t.Helper()
	old := fetchProblem
	fetchProblem = func(ctx context.Context, slug string) (domain.ProblemDetail, error) {
		if slug != "two-sum" {
			return domain.ProblemDetail{}, errors.New("problem not found: " + slug)
		}
		return twoSumDetail(), nil
	}
	t.Cleanup(func() { fetchProblem = old })
}

func stubFetchNoCall(t *testing.T) {
	t.Helper()
	old := fetchProblem
	fetchProblem = func(ctx context.Context, slug string) (domain.ProblemDetail, error) {
		t.Errorf("fetchProblem should not be called, got %q", slug)
		return domain.ProblemDetail{}, errors.New("no fetch")
	}
	t.Cleanup(func() { fetchProblem = old })
}

func chdir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Errorf("restore working dir: %v", err)
		}
	})
}

func capture(t *testing.T, fn func() int) (string, string, int) {
	t.Helper()
	or, ow, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	er, ew, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = ow, ew
	code := fn()
	os.Stdout, os.Stderr = oldOut, oldErr
	_ = ow.Close()
	_ = ew.Close()
	out, _ := io.ReadAll(or)
	errOut, _ := io.ReadAll(er)
	return string(out), string(errOut), code
}

func TestRunDispatch(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"-h"}, {"--help"}} {
		_, _, code := capture(t, func() int { return Run(args) })
		if code != 0 {
			t.Errorf("Run(%v) = %d, want 0", args, code)
		}
	}
	_, errOut, code := capture(t, func() int { return Run([]string{"frobnicate"}) })
	if code != 2 {
		t.Errorf("unknown command = %d, want 2", code)
	}
	if !strings.Contains(errOut, "unknown command") || !strings.Contains(errOut, "leetcli init") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunInitValidation(t *testing.T) {
	stubFetchNoCall(t)
	chdir(t)
	cases := []struct {
		args []string
		want int
	}{
		{[]string{"init"}, 2},
		{[]string{"init", "two-sum", "extra"}, 2},
		{[]string{"init", "--lang"}, 2},
		{[]string{"init", "--lang", "java", "two-sum"}, 2},
		{[]string{"init", "--lang="}, 2},
		{[]string{"init", "two-sum", "--bogus"}, 2},
	}
	for _, c := range cases {
		_, _, code := capture(t, func() int { return Run(c.args) })
		if code != c.want {
			t.Errorf("Run(%v) = %d, want %d", c.args, code, c.want)
		}
	}
}

func TestRunInitCreatesFiles(t *testing.T) {
	stubFetch(t)
	chdir(t)

	_, errOut, code := capture(t, func() int { return Run([]string{"init", "nope"}) })
	if code != 1 {
		t.Fatalf("unknown slug = %d, want 1, stderr %q", code, errOut)
	}

	out, _, code := capture(t, func() int { return Run([]string{"init", "two-sum", "--lang", "python3"}) })
	if code != 0 {
		t.Fatalf("init = %d, want 0", code)
	}
	if !strings.Contains(out, "created solution.py") || !strings.Contains(out, "Two Sum") {
		t.Errorf("stdout = %q", out)
	}
	sol, err := os.ReadFile("solution.py")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sol), "class Solution") {
		t.Errorf("solution.py = %q", sol)
	}
	meta, err := os.ReadFile("leetcli.json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(meta), `"slug": "two-sum"`) || !strings.Contains(string(meta), `"lang": "python"`) {
		t.Errorf("leetcli.json = %q", meta)
	}

	_, errOut, code = capture(t, func() int { return Run([]string{"init", "two-sum", "--lang", "python3"}) })
	if code != 1 {
		t.Fatalf("second init = %d, want 1", code)
	}
	if !strings.Contains(errOut, "--force") {
		t.Errorf("stderr = %q", errOut)
	}

	_, _, code = capture(t, func() int { return Run([]string{"init", "two-sum", "--lang", "python", "--force"}) })
	if code != 0 {
		t.Fatalf("init --force = %d, want 0", code)
	}

	_, _, code = capture(t, func() int { return Run([]string{"init", "two-sum", "--lang", "golang", "--force"}) })
	if code != 0 {
		t.Fatalf("init go = %d, want 0", code)
	}
	goSrc, err := os.ReadFile("solution.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(goSrc), "package main") {
		t.Errorf("solution.go = %q", goSrc)
	}
}

func stubFetchDaily(t *testing.T, challenge domain.DailyChallenge, err error) {
	t.Helper()
	old := fetchDaily
	fetchDaily = func(ctx context.Context) (domain.DailyChallenge, error) {
		return challenge, err
	}
	t.Cleanup(func() { fetchDaily = old })
}

func TestRunInitDaily(t *testing.T) {
	stubFetch(t)
	stubFetchDaily(t, domain.DailyChallenge{
		Date:       "2026-10-07",
		Slug:       "two-sum",
		ID:         1,
		Title:      "Two Sum",
		Difficulty: "Easy",
	}, nil)
	chdir(t)

	out, _, code := capture(t, func() int { return Run([]string{"init", "daily", "--lang", "python3"}) })
	if code != 0 {
		t.Fatalf("init daily = %d, want 0", code)
	}
	if !strings.Contains(out, "Two Sum") {
		t.Errorf("stdout = %q", out)
	}
	meta, err := os.ReadFile("leetcli.json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(meta), `"slug": "two-sum"`) {
		t.Errorf("leetcli.json = %q", meta)
	}
	if _, err := os.Stat("solution.py"); err != nil {
		t.Fatal(err)
	}
}

func TestRunInitDailyFetchError(t *testing.T) {
	stubFetch(t)
	stubFetchDaily(t, domain.DailyChallenge{}, errors.New("daily down"))
	chdir(t)

	_, errOut, code := capture(t, func() int { return Run([]string{"init", "daily"}) })
	if code != 1 {
		t.Fatalf("init daily error = %d, want 1", code)
	}
	if !strings.Contains(errOut, "daily down") {
		t.Errorf("stderr = %q", errOut)
	}
	if _, err := os.Stat("solution.py"); err == nil {
		t.Error("no files should be created on daily fetch error")
	}
}

func TestRunTestFlow(t *testing.T) {
	r, ok := runner.For("python")
	if !ok || !r.Available() {
		t.Skip("python3 is not available")
	}
	stubFetch(t)
	chdir(t)
	if _, _, code := capture(t, func() int { return Run([]string{"init", "two-sum", "--lang", "python3"}) }); code != 0 {
		t.Fatalf("init = %d", code)
	}

	writeSolution(t, `class Solution:
    def twoSum(self, nums, target):
        seen = {}
        for i, n in enumerate(nums):
            if target - n in seen:
                return [seen[target - n], i]
            seen[n] = i
        return []
`)
	out, _, code := capture(t, func() int { return Run([]string{"test"}) })
	if code != 0 {
		t.Fatalf("test = %d, stdout %q", code, out)
	}
	if !strings.Contains(out, "PASS") || !strings.Contains(out, "passed 1/1") {
		t.Errorf("stdout = %q", out)
	}

	writeSolution(t, "class Solution:\n    def twoSum(self, nums, target):\n        return [1, 0]\n")
	out, _, code = capture(t, func() int { return Run([]string{"test"}) })
	if code != 1 {
		t.Fatalf("wrong solution: test = %d, want 1, stdout %q", code, out)
	}
	if !strings.Contains(out, "FAIL") || !strings.Contains(out, "want: [0,1]") || !strings.Contains(out, "got:  [1,0]") {
		t.Errorf("stdout = %q", out)
	}

	writeSolution(t, "class Solution:\n    def twoSum(self, nums, target):\n        raise ValueError('boom')\n")
	out, _, code = capture(t, func() int { return Run([]string{"test"}) })
	if code != 1 {
		t.Fatalf("crash: test = %d, want 1, stdout %q", code, out)
	}
	if !strings.Contains(out, "ERROR") || !strings.Contains(out, "boom") {
		t.Errorf("stdout = %q", out)
	}
}

func TestRunTestWithoutInit(t *testing.T) {
	stubFetchNoCall(t)
	chdir(t)
	_, errOut, code := capture(t, func() int { return Run([]string{"test"}) })
	if code != 1 {
		t.Fatalf("test = %d, want 1", code)
	}
	if !strings.Contains(errOut, "leetcli.json not found") {
		t.Errorf("stderr = %q", errOut)
	}
	if err := os.WriteFile("leetcli.json", []byte(`{"slug": "two-sum", "lang": "python"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_, errOut, code = capture(t, func() int { return Run([]string{"test"}) })
	if code != 1 {
		t.Fatalf("test without solution = %d, want 1", code)
	}
	if !strings.Contains(errOut, "solution.py not found") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunTestRejectsUnknownArgs(t *testing.T) {
	stubFetchNoCall(t)
	chdir(t)
	_, _, code := capture(t, func() int { return Run([]string{"test", "--bogus"}) })
	if code != 2 {
		t.Errorf("test --bogus = %d, want 2", code)
	}
}

func writeSolution(t *testing.T, content string) {
	t.Helper()
	if err := os.WriteFile("solution.py", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
