package ui

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"leetcli/internal/api"
	"leetcli/internal/service"
)

func TestRenderProblemsBrowserFrames(t *testing.T) {
	if os.Getenv("LEETCLI_LIVE") == "" {
		t.Skip("set LEETCLI_LIVE=1")
	}

	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	svc := service.NewProfileService(api.NewClient(), "pr0100pr0")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	problems, total, err := svc.GetProblems(ctx, 200, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) == 0 {
		t.Fatal("no problems fetched")
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	base := Model{
		spinner:        sp,
		theme:          DraculaTheme(),
		themeIndex:     1,
		appView:        ViewProblems,
		problems:       problems,
		problemsLoaded: true,
		problemsTotal:  total,
		width:          100,
		height:         24,
	}

	typing := applyKey(base, "/")
	typing = applyKey(typing, "two")
	active := pressKey(typing, tea.KeyEnter)

	frames := map[string]Model{
		"/tmp/frame_problems_plain.txt":  base,
		"/tmp/frame_problems_typing.txt": typing,
		"/tmp/frame_problems_filter.txt": active,
	}
	for path, m := range frames {
		frame := m.View()
		lines := strings.Split(frame, "\n")
		if len(lines) > m.height {
			t.Errorf("%s: height %d exceeds %d", path, len(lines), m.height)
		}
		for i, l := range lines {
			if lipgloss.Width(l) > m.width {
				t.Errorf("%s: line %d width %d exceeds %d", path, i, lipgloss.Width(l), m.width)
			}
		}
		if err := os.WriteFile(path, []byte(stripANSI(frame)+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("loaded %d of %d problems", len(problems), total)
}
