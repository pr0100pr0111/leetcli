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

func TestRenderLiveDailyFrame(t *testing.T) {
	if os.Getenv("LEETCLI_LIVE") == "" {
		t.Skip("set LEETCLI_LIVE=1")
	}

	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	svc := service.NewProfileService(api.NewClient(), "pr0100pr0")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	daily, err := svc.GetDailyChallenge(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("daily: #%d %s (%s) %s topics=%d", daily.ID, daily.Title, daily.Difficulty, daily.Date, len(daily.Topics))

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
		daily:       daily,
		dailyOK:     true,
	}

	if err := os.WriteFile("/tmp/frame_daily_dashboard.txt", []byte(m.View()), 0o644); err != nil {
		t.Fatal(err)
	}

	m.detailView = DetailDaily
	if err := os.WriteFile("/tmp/frame_daily_detail.txt", []byte(m.View()), 0o644); err != nil {
		t.Fatal(err)
	}

	m.width = 80
	m.height = 24
	m.detailView = DetailNone
	if err := os.WriteFile("/tmp/frame_daily_small.txt", []byte(m.View()), 0o644); err != nil {
		t.Fatal(err)
	}
}
