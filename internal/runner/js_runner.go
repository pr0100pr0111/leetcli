package runner

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
)

type JSRunner struct{}

func (JSRunner) Name() string      { return "javascript" }
func (JSRunner) Extension() string { return "js" }

func (JSRunner) Available() bool {
	return jsBin() != ""
}

func (JSRunner) Scaffold(task Task, dir string) (string, error) {
	return writeScaffold(JSRunner{}, task, dir)
}

func (JSRunner) Run(ctx context.Context, task Task, workDir string) ([]Result, error) {
	return runTask(ctx, JSRunner{}, task, workDir)
}

func init() { Register(JSRunner{}) }

func jsBin() string {
	for _, name := range []string{"node", "nodejs"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

func jsProgram(task Task) (program, error) {
	src, err := jsSource(task)
	if err != nil {
		return program{}, err
	}
	bin := jsBin()
	if bin == "" {
		bin = "node"
	}
	return program{
		Files: map[string]string{"run.js": src},
		Exec:  []string{bin, "run.js"},
	}, nil
}

func jsSource(task Task) (string, error) {
	var b strings.Builder
	b.WriteString(ensureNewline(task.Solution))
	b.WriteString("\n")
	b.WriteString(jsConstants)
	for i, c := range task.Cases {
		_, args, err := argSpecs(task, c.Inputs)
		if err != nil {
			return "", err
		}
		call, err := callExpr(task, args)
		if err != nil {
			return "", err
		}
		b.WriteString(jsCaseBlock(i, call))
	}
	return b.String(), nil
}

func jsCaseBlock(i int, call string) string {
	idx := strconv.Itoa(i)
	return "var _lc_s" + idx + ";\n" +
		"try {\n" +
		"    _lc_s" + idx + " = \"OK\\t\" + _lc_canon(" + call + ");\n" +
		"} catch (_lc_e" + idx + ") {\n" +
		"    _lc_s" + idx + " = \"ERR\\t\" + _lc_msg(_lc_e" + idx + ");\n" +
		"}\n" +
		"_lc_write(\"CASE " + idx + "\\t\" + _lc_s" + idx + " + \"\\n\");\n"
}

const jsConstants = `
function _lc_msg(e) {
    var s = (e instanceof Error) ? (e.name + ": " + e.message) : String(e);
    return s.replace(/\n/g, " ").replace(/\t/g, " ");
}

function _lc_list(v) {
    var out = [];
    var n = 0;
    while (v && n < 10000) {
        out.push(v.val);
        v = v.next;
        n++;
    }
    return out;
}

function _lc_tree(v) {
    if (!v) {
        return [];
    }
    var out = [];
    var q = [v];
    var n = 0;
    while (q.length && n < 10000) {
        n++;
        var c = q.shift();
        if (c === null || c === undefined) {
            out.push(null);
            continue;
        }
        out.push(c.val);
        q.push(c.left === undefined ? null : c.left);
        q.push(c.right === undefined ? null : c.right);
    }
    while (out.length && out[out.length - 1] === null) {
        out.pop();
    }
    return out;
}

function lcMakeList(v) {
    if (!v || v.length === 0) {
        return null;
    }
    var head = new ListNode(v[0]);
    var cur = head;
    for (var i = 1; i < v.length; i++) {
        cur.next = new ListNode(v[i]);
        cur = cur.next;
    }
    return head;
}

function lcMakeTree(v) {
    if (!v || v.length === 0) {
        return null;
    }
    var root = new TreeNode(v[0]);
    var q = [root];
    var i = 1;
    while (q.length && i < v.length) {
        var n = q.shift();
        if (v[i] !== null && v[i] !== undefined) {
            n.left = new TreeNode(v[i]);
            q.push(n.left);
        }
        i++;
        if (i < v.length && v[i] !== null && v[i] !== undefined) {
            n.right = new TreeNode(v[i]);
            q.push(n.right);
        }
        i++;
    }
    return root;
}

function _lc_canon(v) {
    if (v === undefined || v === null) {
        return "";
    }
    if (typeof v === "object" && "val" in v && "next" in v) {
        return JSON.stringify(_lc_list(v));
    }
    if (typeof v === "object" && "val" in v && "left" in v) {
        return JSON.stringify(_lc_tree(v));
    }
    try {
        var s = JSON.stringify(v);
        return s === undefined ? String(v) : s;
    } catch (e) {
        return String(v);
    }
}

function _lc_write(s) {
    process.stdout.write(s);
}
`
