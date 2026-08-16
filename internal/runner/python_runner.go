package runner

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
)

type PythonRunner struct{}

func (PythonRunner) Name() string      { return "python" }
func (PythonRunner) Extension() string { return "py" }

func (PythonRunner) Available() bool {
	return pythonBin() != ""
}

func (PythonRunner) Scaffold(task Task, dir string) (string, error) {
	return writeScaffold(PythonRunner{}, task, dir)
}

func (PythonRunner) Run(ctx context.Context, task Task, workDir string) ([]Result, error) {
	return runTask(ctx, PythonRunner{}, task, workDir)
}

func init() { Register(PythonRunner{}) }

func pythonBin() string {
	for _, name := range []string{"python3", "python"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

func pythonProgram(task Task) (program, error) {
	src, err := pythonSource(task)
	if err != nil {
		return program{}, err
	}
	bin := pythonBin()
	if bin == "" {
		bin = "python3"
	}
	return program{
		Files: map[string]string{"run.py": src},
		Exec:  []string{bin, "run.py"},
	}, nil
}

func pythonSource(task Task) (string, error) {
	var b strings.Builder
	b.WriteString(ensureNewline(task.Solution))
	b.WriteString("\n")
	b.WriteString(pyConstants)
	for i, c := range task.Cases {
		_, args, err := argSpecs(task, c.Inputs)
		if err != nil {
			return "", err
		}
		call, err := callExpr(task, args)
		if err != nil {
			return "", err
		}
		b.WriteString(pyCaseBlock(i, call))
	}
	return b.String(), nil
}

func pyCaseBlock(i int, call string) string {
	idx := strconv.Itoa(i)
	return "try:\n" +
		"    _lc_v" + idx + " = " + call + "\n" +
		"    _lc_s" + idx + " = \"OK\\t\" + _lc_canon(_lc_v" + idx + ")\n" +
		"except Exception as _lc_e" + idx + ":\n" +
		"    _lc_s" + idx + " = \"ERR\\t\" + _lc_msg(_lc_e" + idx + ")\n" +
		"_lc_out.write(\"CASE " + idx + "\\t\" + _lc_s" + idx + " + \"\\n\")\n" +
		"_lc_out.flush()\n"
}

const pyConstants = `
import json as _lc_json
import sys as _lc_sys

_lc_out = _lc_sys.stdout

def _lc_msg(e):
    return (type(e).__name__ + ": " + str(e)).replace("\n", " ").replace("\t", " ")

def lc_make_list(v):
    if not v:
        return None
    head = ListNode(v[0])
    cur = head
    for x in v[1:]:
        cur.next = ListNode(x)
        cur = cur.next
    return head

def lc_make_tree(v):
    if not v:
        return None
    root = TreeNode(v[0])
    q = [root]
    i = 1
    while q and i < len(v):
        n = q.pop(0)
        if v[i] is not None:
            n.left = TreeNode(v[i])
            q.append(n.left)
        i += 1
        if i < len(v) and v[i] is not None:
            n.right = TreeNode(v[i])
            q.append(n.right)
        i += 1
    return root

def _lc_list(v):
    out = []
    while v is not None and len(out) < 10000:
        out.append(v.val)
        v = v.next
    return out

def _lc_tree(v):
    if v is None:
        return []
    out = []
    q = [v]
    while q and len(out) < 10000:
        n = q.pop(0)
        if n is None:
            out.append(None)
            continue
        out.append(n.val)
        q.append(n.left)
        q.append(n.right)
    while out and out[-1] is None:
        out.pop()
    return out

def _lc_canon(v):
    if v is None:
        return ""
    if hasattr(v, "val") and hasattr(v, "next"):
        return _lc_json.dumps(_lc_list(v), separators=(",", ":"), ensure_ascii=False)
    if hasattr(v, "val") and hasattr(v, "left"):
        return _lc_json.dumps(_lc_tree(v), separators=(",", ":"), ensure_ascii=False)
    try:
        return _lc_json.dumps(v, separators=(",", ":"), ensure_ascii=False)
    except TypeError:
        return str(v)
`
