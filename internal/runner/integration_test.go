package runner

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"leetcli/internal/domain"
)

const twoSumGo = `func twoSum(nums []int, target int) []int {
	m := map[int]int{}
	for i, x := range nums {
		if j, ok := m[target-x]; ok {
			return []int{j, i}
		}
		m[x] = i
	}
	return nil
}`

const twoSumPython = `class Solution:
    def twoSum(self, nums, target):
        seen = {}
        for i, x in enumerate(nums):
            if target - x in seen:
                return [seen[target - x], i]
            seen[x] = i
        return []
`

const twoSumPythonBroken = `class Solution:
    def twoSum(self, nums, target):
        raise ValueError("""boom
now""")
`

const twoSumPythonWrong = `class Solution:
    def twoSum(self, nums, target):
        return [1, 0]
`

const twoSumJS = `class Solution {
    twoSum(nums, target) {
        const m = new Map();
        for (let i = 0; i < nums.length; i++) {
            const j = m.get(target - nums[i]);
            if (j !== undefined) return [j, i];
            m.set(nums[i], i);
        }
        return [];
    }
}`

const twoSumCpp = `class Solution {
public:
    vector<int> twoSum(vector<int>& nums, int target) {
        unordered_map<int, int> seen;
        for (int i = 0; i < (int)nums.size(); i++) {
            int need = target - nums[i];
            if (seen.count(need)) return {seen[need], i};
            seen[nums[i]] = i;
        }
        return {};
    }
};`

func twoSumTask(t *testing.T, lang, solution string) Task {
	t.Helper()
	inputs := mustParse(t, "[2,7,11,15]\n9", 2)
	return Task{
		Slug:     "two-sum",
		Lang:     lang,
		Solution: solution,
		Sig: domain.Signature{
			Name:   "twoSum",
			Params: []domain.Param{{Name: "nums", Type: "integer[]"}, {Name: "target", Type: "integer"}},
			Return: "integer[]",
		},
		Cases: []Case{{Inputs: inputs, Expected: "[0,1]"}},
	}
}

func runOne(t *testing.T, r Runner, task Task) Result {
	t.Helper()
	if !r.Available() {
		t.Skipf("%s is not available", r.Name())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	results, err := r.Run(ctx, task, "")
	if err != nil {
		t.Fatalf("%s run: %v", r.Name(), err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	return results[0]
}

func TestRunGoPass(t *testing.T) {
	r, _ := For("go")
	res := runOne(t, r, twoSumTask(t, "go", twoSumGo))
	if res.Status != StatusPass {
		t.Fatalf("status = %s, got %q, want %q, msg %q", res.Status, res.Got, res.Want, res.Message)
	}
}

func TestRunGoFail(t *testing.T) {
	r, _ := For("go")
	wrong := twoSumTask(t, "go", "func twoSum(nums []int, target int) []int { return []int{1, 0} }")
	res := runOne(t, r, wrong)
	if res.Status != StatusFail {
		t.Fatalf("status = %s, want fail", res.Status)
	}
	if res.Got != "[1,0]" {
		t.Errorf("got %q", res.Got)
	}
}

func TestRunGoBuildError(t *testing.T) {
	r, _ := For("go")
	if !r.Available() {
		t.Skip("go is not available")
	}
	broken := twoSumTask(t, "go", "func twoSum(nums []int, target int) []int { return ")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := r.Run(ctx, broken, "")
	var be *BuildError
	if !errors.As(err, &be) {
		t.Fatalf("want BuildError, got %v", err)
	}
	if strings.TrimSpace(be.Output) == "" {
		t.Errorf("build output should not be empty")
	}
}

func TestRunGoPanic(t *testing.T) {
	r, _ := For("go")
	panicky := twoSumTask(t, "go", `func twoSum(nums []int, target int) []int {
	panic("kaboom")
}`)
	res := runOne(t, r, panicky)
	if res.Status != StatusRuntime {
		t.Fatalf("status = %s, want runtime", res.Status)
	}
	if !strings.Contains(res.Message, "kaboom") {
		t.Errorf("message = %q", res.Message)
	}
}

func TestRunPythonPass(t *testing.T) {
	r, _ := For("python")
	res := runOne(t, r, twoSumTask(t, "python", twoSumPython))
	if res.Status != StatusPass {
		t.Fatalf("status = %s, got %q, want %q, msg %q", res.Status, res.Got, res.Want, res.Message)
	}
}

func TestRunPythonFail(t *testing.T) {
	r, _ := For("python")
	res := runOne(t, r, twoSumTask(t, "python", twoSumPythonWrong))
	if res.Status != StatusFail {
		t.Fatalf("status = %s, want fail", res.Status)
	}
	if res.Got != "[1,0]" {
		t.Errorf("got %q", res.Got)
	}
}

func TestRunPythonRuntime(t *testing.T) {
	r, _ := For("python")
	res := runOne(t, r, twoSumTask(t, "python", twoSumPythonBroken))
	if res.Status != StatusRuntime {
		t.Fatalf("status = %s, want runtime", res.Status)
	}
	if !strings.Contains(res.Message, "ValueError") || !strings.Contains(res.Message, "boom") {
		t.Errorf("message = %q", res.Message)
	}
	if strings.Contains(res.Message, "\n") {
		t.Errorf("message should be single line: %q", res.Message)
	}
}

func TestRunJSPass(t *testing.T) {
	r, _ := For("javascript")
	res := runOne(t, r, twoSumTask(t, "javascript", twoSumJS))
	if res.Status != StatusPass {
		t.Fatalf("status = %s, got %q, want %q, msg %q", res.Status, res.Got, res.Want, res.Message)
	}
}

func TestRunCppPass(t *testing.T) {
	r, _ := For("cpp")
	res := runOne(t, r, twoSumTask(t, "cpp", twoSumCpp))
	if res.Status != StatusPass {
		t.Fatalf("status = %s, got %q, want %q, msg %q", res.Status, res.Got, res.Want, res.Message)
	}
}

func TestRunCppBuildError(t *testing.T) {
	r, _ := For("cpp")
	if !r.Available() {
		t.Skip("cpp compiler is not available")
	}
	broken := twoSumTask(t, "cpp", "class Solution { vector<int> twoSum(vector<int>& nums) { return {}; }")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	_, err := r.Run(ctx, broken, "")
	var be *BuildError
	if !errors.As(err, &be) {
		t.Fatalf("want BuildError, got %v", err)
	}
}

func TestRunCppLongReturn(t *testing.T) {
	r, _ := For("cpp")
	if !r.Available() {
		t.Skip("cpp compiler is not available")
	}
	inputs := mustParse(t, "21", 1)
	task := Task{
		Slug: "long-return",
		Lang: "cpp",
		Solution: `class Solution {
public:
    long solve(int n) { return (long)n * 2; }
};`,
		Sig: domain.Signature{
			Name:   "solve",
			Params: []domain.Param{{Name: "n", Type: "integer"}},
			Return: "integer",
		},
		Cases: []Case{{Inputs: inputs, Expected: "42"}},
	}
	res := runOne(t, r, task)
	if res.Status != StatusPass {
		t.Fatalf("status = %s, got %q, want %q, msg %q", res.Status, res.Got, res.Want, res.Message)
	}
}

func TestRunCppBoolVector(t *testing.T) {
	r, _ := For("cpp")
	if !r.Available() {
		t.Skip("cpp compiler is not available")
	}
	inputs := mustParse(t, "[true,false,true]", 1)
	task := Task{
		Slug: "bool-vector",
		Lang: "cpp",
		Solution: `class Solution {
public:
    vector<bool> solve(vector<bool>& flags) { return flags; }
};`,
		Sig: domain.Signature{
			Name:   "solve",
			Params: []domain.Param{{Name: "flags", Type: "boolean[]"}},
			Return: "boolean[]",
		},
		Cases: []Case{{Inputs: inputs, Expected: "[true,false,true]"}},
	}
	res := runOne(t, r, task)
	if res.Status != StatusPass {
		t.Fatalf("status = %s, got %q, want %q, msg %q", res.Status, res.Got, res.Want, res.Message)
	}
}

func TestRunPythonLinkedAndTree(t *testing.T) {
	r, _ := For("python")
	if !r.Available() {
		t.Skip("python is not available")
	}
	solution := `class ListNode:
    def __init__(self, val=0, next=None):
        self.val = val
        self.next = next

class TreeNode:
    def __init__(self, val=0, left=None, right=None):
        self.val = val
        self.left = left
        self.right = right

class Solution:
    def solve(self, head, root):
        out = []
        n = head
        while n:
            out.append(n.val)
            n = n.next
        return out
`
	inputs := mustParse(t, "[1,2,3]\n[1,null,2]", 2)
	task := Task{
		Slug:     "list-tree",
		Lang:     "python",
		Solution: solution,
		Sig: domain.Signature{
			Name:   "solve",
			Params: []domain.Param{{Name: "head", Type: "ListNode"}, {Name: "root", Type: "TreeNode"}},
			Return: "integer[]",
		},
		Cases: []Case{{Inputs: inputs, Expected: "[1,2,3]"}},
	}
	res := runOne(t, r, task)
	if res.Status != StatusPass {
		t.Fatalf("status = %s, got %q, want %q, msg %q", res.Status, res.Got, res.Want, res.Message)
	}
}

func TestRunGoLinkedAndTree(t *testing.T) {
	r, _ := For("go")
	if !r.Available() {
		t.Skip("go is not available")
	}
	solution := `type ListNode struct {
	Val  int
	Next *ListNode
}

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func solve(head *ListNode, root *TreeNode) []int {
	out := []int{}
	for n := head; n != nil; n = n.Next {
		out = append(out, n.Val)
	}
	return out
}`
	inputs := mustParse(t, "[1,2,3]\n[1,null,2]", 2)
	task := Task{
		Slug:     "list-tree",
		Lang:     "go",
		Solution: solution,
		Sig: domain.Signature{
			Name:   "solve",
			Params: []domain.Param{{Name: "head", Type: "ListNode"}, {Name: "root", Type: "TreeNode"}},
			Return: "integer[]",
		},
		Cases: []Case{{Inputs: inputs, Expected: "[1,2,3]"}},
	}
	res := runOne(t, r, task)
	if res.Status != StatusPass {
		t.Fatalf("status = %s, got %q, want %q, msg %q", res.Status, res.Got, res.Want, res.Message)
	}
}

func TestRunPythonVoid(t *testing.T) {
	r, _ := For("python")
	if !r.Available() {
		t.Skip("python is not available")
	}
	solution := `class Solution:
    def solve(self, nums):
        nums.append(4)
`
	inputs := mustParse(t, "[1,2,3]", 1)
	task := Task{
		Slug:     "void",
		Lang:     "python",
		Solution: solution,
		Sig: domain.Signature{
			Name:   "solve",
			Params: []domain.Param{{Name: "nums", Type: "integer[]"}},
			Return: "void",
		},
		Cases: []Case{{Inputs: inputs}},
	}
	res := runOne(t, r, task)
	if res.Status != StatusPass {
		t.Fatalf("status = %s, msg %q", res.Status, res.Message)
	}
}

func TestRunTimeout(t *testing.T) {
	r, _ := For("python")
	if !r.Available() {
		t.Skip("python is not available")
	}
	solution := `class Solution:
    def twoSum(self, nums, target):
        while True:
            pass
`
	task := twoSumTask(t, "python", solution)
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	results, err := r.Run(ctx, task, "")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if results[0].Status != StatusRuntime || results[0].Message != "timeout" {
		t.Fatalf("result = %s / %q", results[0].Status, results[0].Message)
	}
}

func TestScaffoldThenRunGo(t *testing.T) {
	r, _ := For("go")
	if !r.Available() {
		t.Skip("go is not available")
	}
	task := twoSumTask(t, "go", twoSumGo)
	dir := t.TempDir()
	path, err := r.Scaffold(task, dir)
	if err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	data, err := readFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(data, "package main") {
		t.Fatalf("scaffold = %q", data)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	results, err := r.Run(ctx, task, dir)
	if err != nil {
		t.Fatalf("run in scaffold dir: %v", err)
	}
	if results[0].Status != StatusPass {
		t.Fatalf("status = %s, msg %q", results[0].Status, results[0].Message)
	}
	if _, err := readFile(filepath.Join(dir, "run.go")); err != nil {
		t.Errorf("run.go should be left in work dir: %v", err)
	}
}

func TestNewTaskTwoSumLiveFixture(t *testing.T) {
	detail := domain.ProblemDetail{
		Slug: "two-sum",
		Signature: domain.Signature{
			Name:   "twoSum",
			Params: []domain.Param{{Name: "nums", Type: "integer[]"}, {Name: "target", Type: "integer"}},
			Return: "integer[]",
		},
		Snippets:    []domain.CodeSnippet{{LangSlug: "python3", Code: twoSumPython}},
		SampleInput: "[2,7,11,15]\n9",
		Examples:    []domain.Example{{Number: 1, Output: "[0,1]"}},
	}
	task, err := NewTask(detail, "python3")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	r, _ := For(task.Lang)
	res := runOne(t, r, task)
	if res.Status != StatusPass {
		t.Fatalf("status = %s, got %q, want %q, msg %q", res.Status, res.Got, res.Want, res.Message)
	}
}

func TestAssembleFloatCompareFromHarness(t *testing.T) {
	task := Task{
		Sig:   domain.Signature{Name: "f", Return: "double"},
		Cases: []Case{{Expected: "0.5"}},
	}
	results := assemble(task, "CASE 0\tOK\t0.500000\n", "", nil)
	if results[0].Status != StatusPass {
		t.Fatalf("status = %s, got %q", results[0].Status, results[0].Got)
	}
}
