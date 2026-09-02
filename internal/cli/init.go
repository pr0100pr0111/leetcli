package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"leetcli/internal/runner"
)

func runInit(args []string) int {
	langArg := "python3"
	force := false
	slug := ""
	positional := 0
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--lang":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--lang requires a value")
				return 2
			}
			i++
			langArg = args[i]
		case strings.HasPrefix(a, "--lang="):
			langArg = strings.TrimPrefix(a, "--lang=")
		case a == "--force":
			force = true
		case a == "-h" || a == "--help":
			fmt.Print(usage)
			return 0
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "unknown flag %q\n", a)
			return 2
		default:
			positional++
			if positional > 1 {
				fmt.Fprintln(os.Stderr, "usage: leetcli init <slug> [--lang <name>] [--force]")
				return 2
			}
			slug = a
		}
	}
	if slug == "" {
		fmt.Fprintln(os.Stderr, "usage: leetcli init <slug> [--lang <name>] [--force]")
		return 2
	}
	lang, leetSlug, err := runner.ResolveLang(langArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if slug == "daily" {
		d, err := fetchDaily(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		slug = d.Slug
	}
	detail, err := fetchProblem(ctx, slug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	task, err := runner.NewTask(detail, leetSlug)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	r, ok := runner.For(lang)
	if !ok {
		fmt.Fprintf(os.Stderr, "unsupported language %q\n", lang)
		return 2
	}
	solution := "solution." + r.Extension()
	var existing []string
	if fileExists(solution) {
		existing = append(existing, solution)
	}
	if fileExists("leetcli.json") {
		existing = append(existing, "leetcli.json")
	}
	if len(existing) > 0 && !force {
		fmt.Fprintf(os.Stderr, "%s already exists — use --force to overwrite\n", strings.Join(existing, ", "))
		return 1
	}
	path, err := r.Scaffold(task, ".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("created %s and leetcli.json\n", path)
	fmt.Printf("problem: %s (%s)\n", detail.Title, detail.Difficulty)
	fmt.Println("edit the solution, then run: leetcli test")
	return 0
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
