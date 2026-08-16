package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type GoRunner struct{}

func (GoRunner) Name() string      { return "go" }
func (GoRunner) Extension() string { return "go" }

func (GoRunner) Available() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

func (GoRunner) Scaffold(task Task, dir string) (string, error) {
	return writeScaffold(GoRunner{}, task, dir)
}

func (GoRunner) Run(ctx context.Context, task Task, workDir string) ([]Result, error) {
	return runTask(ctx, GoRunner{}, task, workDir)
}

func init() { Register(GoRunner{}) }

func goProgram(task Task) (program, error) {
	src, err := goSource(task)
	if err != nil {
		return program{}, err
	}
	return program{
		Files: map[string]string{"run.go": src},
		Build: []string{"go", "build", "-o", binName(), "run.go"},
		Exec:  []string{"./" + binName()},
	}, nil
}

func goSource(task Task) (string, error) {
	harness, err := goHarness(task)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("package main\n\n")
	b.WriteString("import (\n\t\"bytes\"\n\t\"encoding/json\"\n\t\"fmt\"\n\t\"reflect\"\n\t\"strings\"\n)\n\n")
	b.WriteString(strings.TrimRight(stripPackage(task.Solution), "\n"))
	b.WriteString("\n\n")
	b.WriteString(harness)
	return b.String(), nil
}

func goHarness(task Task) (string, error) {
	var b strings.Builder
	b.WriteString(goCanonFunc(task))
	b.WriteString(goMsgFunc())
	b.WriteString(goRunCaseFunc())
	if usesType(task, "ListNode") {
		b.WriteString(goListFunc())
	}
	if usesType(task, "TreeNode") {
		b.WriteString(goTreeFunc())
	}
	b.WriteString("func main() {\n")
	for i, c := range task.Cases {
		_, args, err := argSpecs(task, c.Inputs)
		if err != nil {
			return "", err
		}
		call, err := callExpr(task, args)
		if err != nil {
			return "", err
		}
		if strings.EqualFold(task.Sig.Return, "void") {
			fmt.Fprintf(&b, "\tlcRunCase(%d, func() any { %s; return nil })\n", i, call)
		} else {
			fmt.Fprintf(&b, "\tlcRunCase(%d, func() any { return %s })\n", i, call)
		}
	}
	b.WriteString("}\n")
	return b.String(), nil
}

func goCanonFunc(task Task) string {
	needList := usesType(task, "ListNode")
	needTree := usesType(task, "TreeNode")
	var b strings.Builder
	b.WriteString(`
func lcCanon(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
`)
	if needList {
		b.WriteString(`	case *ListNode:
		out := make([]int, 0, 8)
		n := t
		for n != nil && len(out) < 10000 {
			out = append(out, n.Val)
			n = n.Next
		}
		data, _ := json.Marshal(out)
		return string(data)
`)
	}
	if needTree {
		b.WriteString(`	case *TreeNode:
		if t == nil {
			return ""
		}
		out := []any{}
		q := []*TreeNode{t}
		for len(q) > 0 && len(out) < 10000 {
			n := q[0]
			q = q[1:]
			if n == nil {
				out = append(out, nil)
				continue
			}
			out = append(out, n.Val)
			q = append(q, n.Left, n.Right)
		}
		for len(out) > 0 && out[len(out)-1] == nil {
			out = out[:len(out)-1]
		}
		data, _ := json.Marshal(out)
		return string(data)
`)
	}
	b.WriteString(`	default:
		rv := reflect.ValueOf(t)
		if rv.Kind() == reflect.Slice && rv.IsNil() {
			return "[]"
		}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(v); err != nil {
			return fmt.Sprint(v)
		}
		return strings.TrimRight(buf.String(), "\n")
	}
}
`)
	return b.String()
}

func goMsgFunc() string {
	return `
func lcMsg(v any) string {
	s := fmt.Sprint(v)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}
`
}

func goRunCaseFunc() string {
	return `
func lcRunCase(i int, fn func() any) {
	out := "OK"
	func() {
		defer func() {
			if r := recover(); r != nil {
				out = "ERR\t" + lcMsg(r)
			}
		}()
		out = "OK\t" + lcCanon(fn())
	}()
	fmt.Printf("CASE %d\t%s\n", i, out)
}
`
}

func goListFunc() string {
	return `
func lcMakeList(vals []int) *ListNode {
	if len(vals) == 0 {
		return nil
	}
	head := &ListNode{Val: vals[0]}
	cur := head
	for _, v := range vals[1:] {
		cur.Next = &ListNode{Val: v}
		cur = cur.Next
	}
	return head
}
`
}

func goTreeFunc() string {
	return `
func lcMakeTree(vals []any) *TreeNode {
	if len(vals) == 0 {
		return nil
	}
	root := &TreeNode{}
	lcTreeVal(&root.Val, vals[0])
	q := []*TreeNode{root}
	i := 1
	for len(q) > 0 && i < len(vals) {
		n := q[0]
		q = q[1:]
		if vals[i] != nil {
			left := &TreeNode{}
			lcTreeVal(&left.Val, vals[i])
			n.Left = left
			q = append(q, left)
		}
		i++
		if i < len(vals) && vals[i] != nil {
			right := &TreeNode{}
			lcTreeVal(&right.Val, vals[i])
			n.Right = right
			q = append(q, right)
		}
		i++
	}
	return root
}

func lcTreeVal(dst *int, v any) {
	switch t := v.(type) {
	case int:
		*dst = t
	case int64:
		*dst = int(t)
	case float64:
		*dst = int(t)
	}
}
`
}
