package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type CppRunner struct{}

func (CppRunner) Name() string      { return "cpp" }
func (CppRunner) Extension() string { return "cpp" }

func (CppRunner) Available() bool {
	return cppCompiler() != ""
}

func (CppRunner) Scaffold(task Task, dir string) (string, error) {
	return writeScaffold(CppRunner{}, task, dir)
}

func (CppRunner) Run(ctx context.Context, task Task, workDir string) ([]Result, error) {
	return runTask(ctx, CppRunner{}, task, workDir)
}

func init() { Register(CppRunner{}) }

func cppCompiler() string {
	for _, name := range []string{"g++", "clang++"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

func cppHeader() string {
	return `#include <algorithm>
#include <array>
#include <cctype>
#include <climits>
#include <cmath>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <deque>
#include <exception>
#include <iomanip>
#include <list>
#include <map>
#include <numeric>
#include <optional>
#include <queue>
#include <set>
#include <sstream>
#include <stack>
#include <string>
#include <unordered_map>
#include <unordered_set>
#include <vector>

using namespace std;

`
}

func cppProgram(task Task) (program, error) {
	src, err := cppSource(task)
	if err != nil {
		return program{}, err
	}
	compiler := cppCompiler()
	if compiler == "" {
		compiler = "g++"
	}
	return program{
		Files: map[string]string{"run.cpp": src},
		Build: []string{compiler, "-O2", "-std=c++17", "-o", binName(), "run.cpp"},
		Exec:  []string{"./" + binName()},
	}, nil
}

func cppSource(task Task) (string, error) {
	harness, err := cppHarness(task)
	if err != nil {
		return "", err
	}
	return cppHeader() + ensureNewline(task.Solution) + "\n" + harness, nil
}

func cppHarness(task Task) (string, error) {
	var b strings.Builder
	b.WriteString(cppMsgFunc())
	b.WriteString(cppQuoteFunc())
	b.WriteString(cppScalarVals())
	if usesType(task, "ListNode") {
		b.WriteString(cppListVals())
	}
	if usesType(task, "TreeNode") {
		b.WriteString(cppTreeVals())
	}
	b.WriteString(cppVectorVal())
	b.WriteString("int main() {\n")
	b.WriteString("    setvbuf(stdout, nullptr, _IONBF, 0);\n")
	for i, c := range task.Cases {
		decls, args, err := argSpecs(task, c.Inputs)
		if err != nil {
			return "", err
		}
		call, err := callExpr(task, args)
		if err != nil {
			return "", err
		}
		idx := strconv.Itoa(i)
		var body strings.Builder
		for _, d := range decls {
			body.WriteString("        " + d + "\n")
		}
		if strings.EqualFold(task.Sig.Return, "void") {
			body.WriteString("        " + call + ";\n")
			body.WriteString("        printf(\"CASE " + idx + "\\tOK\\n\");\n")
		} else {
			body.WriteString("        auto _r" + idx + " = " + call + ";\n")
			body.WriteString("        printf(\"CASE " + idx + "\\tOK\\t%s\\n\", lcVal(_r" + idx + ").c_str());\n")
		}
		b.WriteString("    try {\n")
		b.WriteString(body.String())
		b.WriteString("    } catch (const exception& _lc_e) {\n")
		fmt.Fprintf(&b, "        printf(\"CASE %s\\tERR\\t%%s\\n\", lcMsg(_lc_e.what()).c_str());\n", idx)
		b.WriteString("    } catch (...) {\n")
		fmt.Fprintf(&b, "        printf(\"CASE %s\\tERR\\tunknown error\\n\");\n", idx)
		b.WriteString("    }\n")
	}
	b.WriteString("    return 0;\n}\n")
	return b.String(), nil
}

func cppMsgFunc() string {
	return `
static string lcMsg(const string& s) {
    string out;
    for (size_t i = 0; i < s.size(); i++) {
        char c = s[i];
        if (c == '\n' || c == '\r' || c == '\t') {
            out += ' ';
        } else {
            out += c;
        }
    }
    return out;
}
`
}

func cppQuoteFunc() string {
	return `
static string lcQuote(const string& v) {
    string out = "\"";
    for (size_t i = 0; i < v.size(); i++) {
        char c = v[i];
        if (c == '\"') {
            out += "\\\"";
        } else if (c == '\\') {
            out += "\\\\";
        } else if (c == '\n') {
            out += "\\n";
        } else if (c == '\t') {
            out += "\\t";
        } else if (c == '\r') {
            out += "\\r";
        } else {
            out += c;
        }
    }
    out += "\"";
    return out;
}
`
}

func cppScalarVals() string {
	return `
static string lcVal(const string& v) { return lcQuote(v); }
static string lcVal(const char* v) { return lcQuote(string(v)); }
static string lcVal(char v) { return lcQuote(string(1, v)); }
static string lcVal(long long v) { return to_string(v); }
static string lcVal(long v) { return to_string(v); }
static string lcVal(int v) { return to_string(v); }
static string lcVal(short v) { return to_string((long long)v); }
static string lcVal(unsigned long long v) { return to_string(v); }
static string lcVal(unsigned long v) { return to_string(v); }
static string lcVal(unsigned v) { return to_string(v); }
static string lcVal(double v) { ostringstream o; o << setprecision(15) << v; return o.str(); }
static string lcVal(bool v) { return v ? "true" : "false"; }
`
}

func cppListVals() string {
	return `
static ListNode* lcMakeList(const vector<int>& v) {
    if (v.empty()) return nullptr;
    ListNode* head = new ListNode(v[0]);
    ListNode* cur = head;
    for (size_t i = 1; i < v.size(); i++) {
        cur->next = new ListNode(v[i]);
        cur = cur->next;
    }
    return head;
}

static string lcVal(ListNode* n) {
    string out = "[";
    int guard = 0;
    while (n != nullptr && guard < 10000) {
        if (guard > 0) out += ",";
        out += lcVal((long long)n->val);
        n = n->next;
        guard++;
    }
    out += "]";
    return out;
}
`
}

func cppTreeVals() string {
	return `
static TreeNode* lcMakeTree(const vector<optional<long long>>& v) {
    if (v.empty() || !v[0].has_value()) return nullptr;
    TreeNode* root = new TreeNode((int)*v[0]);
    vector<TreeNode*> q = {root};
    size_t i = 1;
    size_t qi = 0;
    while (qi < q.size() && i < v.size()) {
        TreeNode* n = q[qi++];
        if (v[i].has_value()) {
            n->left = new TreeNode((int)*v[i]);
            q.push_back(n->left);
        }
        i++;
        if (i < v.size() && v[i].has_value()) {
            n->right = new TreeNode((int)*v[i]);
            q.push_back(n->right);
        }
        i++;
    }
    return root;
}

static string lcVal(TreeNode* t) {
    if (t == nullptr) return "";
    vector<TreeNode*> q = {t};
    vector<optional<string>> out;
    size_t qi = 0;
    size_t guard = 0;
    while (qi < q.size() && guard < 10000) {
        guard++;
        TreeNode* n = q[qi++];
        if (n == nullptr) {
            out.push_back(nullopt);
            continue;
        }
        out.push_back(lcVal((long long)n->val));
        q.push_back(n->left);
        q.push_back(n->right);
    }
    while (!out.empty() && !out.back().has_value()) {
        out.pop_back();
    }
    string s = "[";
    for (size_t i = 0; i < out.size(); i++) {
        if (i > 0) s += ",";
        s += out[i].has_value() ? *out[i] : string("null");
    }
    s += "]";
    return s;
}
`
}

func cppVectorVal() string {
	return `
template <class T>
static string lcVal(const vector<T>& v) {
    string out = "[";
    for (size_t i = 0; i < v.size(); i++) {
        if (i > 0) out += ",";
        out += lcVal(v[i]);
    }
    out += "]";
    return out;
}
`
}
