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

func TestRenderLiveContestsFrame(t *testing.T) {
	if os.Getenv("LEETCLI_LIVE") == "" {
		t.Skip("set LEETCLI_LIVE=1")
	}

	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	svc := service.NewProfileService(api.NewClient(), "Radewoosh")
	history, err := svc.GetContests(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("contests: rating=%.2f top=%.2f results=%d attended=%d",
		history.Rating, history.TopPercent, len(history.Results), history.AttendedCount())

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	m := Model{
		spinner:     sp,
		theme:       DraculaTheme(),
		themeIndex:  1,
		activePanel: PanelContests,
		showHelp:    true,
		config:      config.Config{Username: "pr0100pr0", Theme: "dracula"},
		width:       140,
		height:      40,
		contests:    history,
		contestsOK:  true,
	}

	if err := os.WriteFile("/tmp/frame_contests_dashboard.txt", []byte(m.View()), 0o644); err != nil {
		t.Fatal(err)
	}

	m.detailView = DetailContests
	if err := os.WriteFile("/tmp/frame_contests_detail.txt", []byte(m.View()), 0o644); err != nil {
		t.Fatal(err)
	}

	m.width = 80
	m.height = 24
	if err := os.WriteFile("/tmp/frame_contests_detail_small.txt", []byte(m.View()), 0o644); err != nil {
		t.Fatal(err)
	}

	own := service.NewProfileService(api.NewClient(), "pr0100pr0")
	ownHistory, err := own.GetContests(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("own contests: hasRating=%v results=%d", ownHistory.HasRating, len(ownHistory.Results))

	m.width = 140
	m.height = 40
	m.contests = ownHistory
	if err := os.WriteFile("/tmp/frame_contests_empty.txt", []byte(m.View()), 0o644); err != nil {
		t.Fatal(err)
	}

	daily, err := svc.GetDailyChallenge(ctx)
	if err != nil {
		t.Fatal(err)
	}
	m.daily = daily
	m.dailyOK = true
	m.contests = history
	m.detailView = DetailNone
	if err := os.WriteFile("/tmp/frame_header_both.txt", []byte(m.View()), 0o644); err != nil {
		t.Fatal(err)
	}
}
