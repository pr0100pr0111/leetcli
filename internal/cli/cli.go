package cli

import (
	"context"
	"fmt"
	"os"

	"leetcli/internal/api"
	"leetcli/internal/domain"
	"leetcli/internal/service"
)

const usage = `leetcli — LeetCode dashboard and local test runner

Usage:
  leetcli                      open the dashboard
  leetcli init <slug> [flags]  create solution files in the current directory
                               (use "daily" for today's Daily Challenge)
  leetcli test [flags]         run local tests for the current directory
  leetcli help                 show this help

Init flags:
  --lang <name>  golang | python3 | javascript | cpp (default: python3)
  --force        overwrite existing files

Test flags:
  --keep         keep generated sources and print their path
`

var fetchProblem = func(ctx context.Context, slug string) (domain.ProblemDetail, error) {
	return service.NewProfileService(api.NewClient(), "").GetProblemDetail(ctx, slug)
}

var fetchDaily = func(ctx context.Context) (domain.DailyChallenge, error) {
	return service.NewProfileService(api.NewClient(), "").GetDailyChallenge(ctx)
}

func Run(args []string) int {
	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "test":
		return runTest(args[1:])
	case "help", "-h", "--help":
		fmt.Print(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}
