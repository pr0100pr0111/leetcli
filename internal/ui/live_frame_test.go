package ui

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"leetcli/internal/api"
	"leetcli/internal/config"
	"leetcli/internal/service"
)

func TestRenderLiveFrame(t *testing.T) {
	if os.Getenv("LEETCLI_LIVE") == "" {
		t.Skip("set LEETCLI_LIVE=1")
	}

	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	svc := service.NewProfileService(api.NewClient(), "pr0100pr0")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	profile, err := svc.GetProfile(ctx)
	if err != nil {
		t.Fatal(err)
	}

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	m := Model{
		spinner:     sp,
		theme:       DraculaTheme(),
		themeIndex:  1,
		activePanel: PanelDifficulty,
		showHelp:    true,
		config:      config.Config{Username: "pr0100pr0", Theme: "dracula"},
		width:       200,
		height:      45,
		profile:     profile,
	}

	if err := os.WriteFile("/tmp/frame_live.txt", []byte(m.View()), 0o644); err != nil {
		t.Fatal(err)
	}
}
