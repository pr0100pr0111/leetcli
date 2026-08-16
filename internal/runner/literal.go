package runner

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf8"
)

func supportedType(typ string) bool {
	base := typ
	for strings.HasSuffix(base, "[]") {
		base = strings.TrimSuffix(base, "[]")
	}
	switch base {
	case "integer", "int", "double", "float", "string",
		"boolean", "bool", "character", "char",
		"ListNode", "TreeNode":
		return true
	}
	return false
}

func lit(lang, typ string, v any) (string, error) {
	if strings.HasSuffix(typ, "[]") {
		return arrayLit(lang, strings.TrimSuffix(typ, "[]"), v)
	}
	switch typ {
	case "integer", "int", "double", "float":
		return numLit(v)
	case "string":
		return strLit(lang, v)
	case "boolean", "bool":
		return boolLit(lang, v)
	case "character", "char":
		return charLit(lang, v)
	case "ListNode":
		return listLit(lang, v)
	case "TreeNode":
		return treeLit(lang, v)
	}
	return "", fmt.Errorf("unsupported type %q", typ)
}

func numLit(v any) (string, error) {
	n, ok := v.(json.Number)
	if !ok {
		return "", fmt.Errorf("expected a number, got %v", v)
	}
	return n.String(), nil
}

func strLit(lang string, v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("expected a string, got %v", v)
	}
	switch lang {
	case "go":
		return strconv.Quote(s), nil
	case "python", "javascript":
		b, err := json.Marshal(s)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case "cpp":
		return cppQuote(s), nil
	}
	return "", fmt.Errorf("unsupported language %q", lang)
}

func boolLit(lang string, v any) (string, error) {
	b, ok := v.(bool)
	if !ok {
		return "", fmt.Errorf("expected a boolean, got %v", v)
	}
	if lang == "python" {
		if b {
			return "True", nil
		}
		return "False", nil
	}
	return strconv.FormatBool(b), nil
}

func charLit(lang string, v any) (string, error) {
	s, ok := v.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("expected a character, got %v", v)
	}
	r, _ := utf8.DecodeRuneInString(s)
	switch lang {
	case "go":
		return strconv.QuoteRune(r), nil
	case "cpp":
		var b strings.Builder
		b.WriteByte('\'')
		writeCharBody(&b, r, '\'')
		b.WriteByte('\'')
		return b.String(), nil
	}
	return strLit(lang, v)
}

func writeCharBody(b *strings.Builder, r rune, quote rune) {
	switch r {
	case '\\':
		b.WriteString(`\\`)
	case '\n':
		b.WriteString(`\n`)
	case '\t':
		b.WriteString(`\t`)
	case '\r':
		b.WriteString(`\r`)
	default:
		if r == quote {
			b.WriteByte('\\')
			b.WriteRune(r)
		} else if r < 0x20 {
			fmt.Fprintf(b, "\\u%04x", r)
		} else {
			b.WriteRune(r)
		}
	}
}

func cppQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, "\\u%04x", r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

func arrayLit(lang, inner string, v any) (string, error) {
	items, err := arrayItems(v)
	if err != nil {
		return "", err
	}
	elems := make([]string, 0, len(items))
	for i, it := range items {
		e, err := lit(lang, inner, it)
		if err != nil {
			return "", fmt.Errorf("element %d: %w", i, err)
		}
		elems = append(elems, e)
	}
	joined := strings.Join(elems, ", ")
	switch lang {
	case "go":
		return "[]" + goType(inner) + "{" + joined + "}", nil
	case "python", "javascript":
		return "[" + joined + "]", nil
	case "cpp":
		return "std::vector<" + cppType(inner) + ">{" + joined + "}", nil
	}
	return "", fmt.Errorf("unsupported language %q", lang)
}

func listLit(lang string, v any) (string, error) {
	items, err := arrayItems(v)
	if err != nil {
		return "", err
	}
	elems := make([]string, 0, len(items))
	for i, it := range items {
		s, err := numLit(it)
		if err != nil {
			return "", fmt.Errorf("linked list value %d must be a number: %w", i, err)
		}
		elems = append(elems, s)
	}
	joined := strings.Join(elems, ", ")
	switch lang {
	case "go":
		return "lcMakeList([]int{" + joined + "})", nil
	case "python":
		return "lc_make_list([" + joined + "])", nil
	case "javascript":
		return "lcMakeList([" + joined + "])", nil
	case "cpp":
		return "lcMakeList(std::vector<int>{" + joined + "})", nil
	}
	return "", fmt.Errorf("unsupported language %q", lang)
}

func treeLit(lang string, v any) (string, error) {
	items, err := arrayItems(v)
	if err != nil {
		return "", err
	}
	elems := make([]string, 0, len(items))
	for i, it := range items {
		if it == nil {
			if lang == "cpp" {
				elems = append(elems, "std::nullopt")
			} else {
				elems = append(elems, nullLit(lang))
			}
			continue
		}
		s, err := numLit(it)
		if err != nil {
			return "", fmt.Errorf("tree value %d must be a number: %w", i, err)
		}
		elems = append(elems, s)
	}
	joined := strings.Join(elems, ", ")
	switch lang {
	case "go":
		return "lcMakeTree([]any{" + joined + "})", nil
	case "python":
		return "lc_make_tree([" + joined + "])", nil
	case "javascript":
		return "lcMakeTree([" + joined + "])", nil
	case "cpp":
		return "lcMakeTree(std::vector<std::optional<long long>>{" + joined + "})", nil
	}
	return "", fmt.Errorf("unsupported language %q", lang)
}

func nullLit(lang string) string {
	switch lang {
	case "go":
		return "nil"
	case "python":
		return "None"
	default:
		return "null"
	}
}

func arrayItems(v any) ([]any, error) {
	if v == nil {
		return nil, nil
	}
	items, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("expected an array, got %v", v)
	}
	return items, nil
}

func goType(typ string) string {
	if strings.HasSuffix(typ, "[]") {
		return "[]" + goType(strings.TrimSuffix(typ, "[]"))
	}
	switch typ {
	case "integer", "int":
		return "int"
	case "double", "float":
		return "float64"
	case "string":
		return "string"
	case "boolean", "bool":
		return "bool"
	case "character", "char":
		return "rune"
	case "ListNode":
		return "*ListNode"
	case "TreeNode":
		return "*TreeNode"
	}
	return "any"
}

func cppType(typ string) string {
	if strings.HasSuffix(typ, "[]") {
		return "std::vector<" + cppType(strings.TrimSuffix(typ, "[]")) + ">"
	}
	switch typ {
	case "integer", "int":
		return "int"
	case "double", "float":
		return "double"
	case "string":
		return "std::string"
	case "boolean", "bool":
		return "bool"
	case "character", "char":
		return "char"
	case "ListNode":
		return "ListNode*"
	case "TreeNode":
		return "TreeNode*"
	}
	return "auto"
}

func argSpecs(task Task, inputs []any) ([]string, []string, error) {
	if len(inputs) != len(task.Sig.Params) {
		return nil, nil, fmt.Errorf("got %d input values, expected %d", len(inputs), len(task.Sig.Params))
	}
	var decls, args []string
	for i, p := range task.Sig.Params {
		l, err := lit(task.Lang, p.Type, inputs[i])
		if err != nil {
			return nil, nil, fmt.Errorf("argument %q: %w", p.Name, err)
		}
		if task.Lang == "cpp" {
			name := "p" + strconv.Itoa(i)
			decls = append(decls, cppType(p.Type)+" "+name+" = "+l+";")
			args = append(args, name)
		} else {
			args = append(args, l)
		}
	}
	return decls, args, nil
}

var (
	identRe         = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	classSolutionRe = regexp.MustCompile(`class\s+Solution\b`)
)

func callExpr(task Task, args []string) (string, error) {
	fn := task.Sig.Name
	if !identRe.MatchString(fn) {
		return "", fmt.Errorf("invalid function name %q", fn)
	}
	joined := strings.Join(args, ", ")
	switch task.Lang {
	case "go":
		return fn + "(" + joined + ")", nil
	case "python":
		if strings.Contains(task.Solution, "class Solution") {
			return "Solution()." + fn + "(" + joined + ")", nil
		}
		return fn + "(" + joined + ")", nil
	case "javascript":
		if classSolutionRe.MatchString(task.Solution) {
			return "new Solution()." + fn + "(" + joined + ")", nil
		}
		return fn + "(" + joined + ")", nil
	case "cpp":
		if classSolutionRe.MatchString(task.Solution) {
			return "Solution()." + fn + "(" + joined + ")", nil
		}
		return fn + "(" + joined + ")", nil
	}
	return "", fmt.Errorf("unsupported language %q", task.Lang)
}

func usesType(task Task, names ...string) bool {
	all := task.Sig.Return
	for _, p := range task.Sig.Params {
		all += " " + p.Type
	}
	for _, n := range names {
		if strings.Contains(all, n) {
			return true
		}
	}
	return false
}

func stripPackage(s string) string {
	trimmed := strings.TrimLeft(s, " \t\n")
	if strings.HasPrefix(trimmed, "package ") {
		if idx := strings.IndexByte(trimmed, '\n'); idx >= 0 {
			return trimmed[idx+1:]
		}
		return ""
	}
	return s
}

func ensureNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

func binName() string {
	if runtime.GOOS == "windows" {
		return "bin.exe"
	}
	return "bin"
}
