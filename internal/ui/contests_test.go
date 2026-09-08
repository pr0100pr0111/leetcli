package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"leetcli/internal/domain"
)

func contestsModel(width, height int) Model {
	m := testModel(width, height)
	m.contests = domain.ContestHistory{
		HasRating:  true,
		Rating:     1353.2,
		TopPercent: 92.6,
		Results: []domain.ContestResult{
			{Title: "Weekly Contest 212", Time: time.Unix(1603593000, 0), Rating: 1411.8, Ranking: 9091, Attended: true, HasRating: true},
			{Title: "Biweekly Contest 70", Time: time.Unix(1610000000, 0)},
			{Title: "Weekly Contest 296", Time: time.Unix(1654396200, 0), Rating: 1353.2, Ranking: 11056, Attended: true, HasRating: true},
		},
	}
	m.contestsOK = true
	return m
}

func TestContestsDashboardLine(t *testing.T) {
	m := contestsModel(140, 40)
	frame := stripANSI(m.View())
	for _, want := range []string{"Contests", "1353.2 rating", "92.6%", "2 attended"} {
		if !strings.Contains(frame, want) {
			t.Errorf("dashboard missing %q:\n%s", want, frame)
		}
	}

	m2 := contestsModel(140, 40)
	m2.contestsOK = false
	if strings.Contains(stripANSI(m2.View()), "Contests") {
		t.Error("dashboard should hide contests line when unavailable and inactive")
	}
}

func TestContestsLineActiveWhenNotLoaded(t *testing.T) {
	m := testModel(140, 40)
	m.activePanel = PanelContests
	frame := stripANSI(m.View())
	if !strings.Contains(frame, "Contests") || !strings.Contains(frame, "unavailable") {
		t.Errorf("active contests line should be visible:\n%s", frame)
	}
}

func TestContestsLineEmptyState(t *testing.T) {
	m := testModel(140, 40)
	m.contestsOK = true
	frame := stripANSI(m.View())
	if !strings.Contains(frame, "no contests yet") {
		t.Errorf("dashboard should show empty state:\n%s", frame)
	}
}

func TestContestsPanelNavigation(t *testing.T) {
	m := contestsModel(140, 40)
	m = applyKey(m, "tab")
	m = applyKey(m, "tab")
	m = applyKey(m, "tab")
	if m.activePanel != PanelContests {
		t.Fatalf("activePanel = %v, want PanelContests", m.activePanel)
	}

	m = applyKey(m, "enter")
	if m.detailView != DetailContests {
		t.Fatalf("detailView = %v, want DetailContests", m.detailView)
	}

	frame := stripANSI(m.View())
	for _, want := range []string{
		"Contest History",
		"Rating 1353.2",
		"Top 92.6%",
		"Attended 2",
		"max 1411.8",
		"min 1353.2",
		"Weekly Contest 296",
		"#11056",
		"[esc] back",
	} {
		if !strings.Contains(frame, want) {
			t.Errorf("detail frame missing %q:\n%s", want, frame)
		}
	}
	if !strings.Contains(frame, "█") {
		t.Errorf("detail frame should contain the rating graph:\n%s", frame)
	}
	if iRecent := strings.Index(frame, "Weekly Contest 296"); iRecent > strings.Index(frame, "Weekly Contest 212") {
		t.Error("recent contests should be listed first")
	}
	if !strings.Contains(frame, "Biweekly Contest 70") || !strings.Contains(frame, "—") {
		t.Errorf("unrated contest should be listed with dashes:\n%s", frame)
	}

	m = pressKey(m, tea.KeyEsc)
	if m.detailView != DetailNone {
		t.Errorf("esc should close detail view: %v", m.detailView)
	}
}

func TestContestsDetailStates(t *testing.T) {
	m := testModel(140, 40)
	m.activePanel = PanelContests
	m = applyKey(m, "enter")
	if frame := stripANSI(m.View()); !strings.Contains(frame, "unavailable") {
		t.Errorf("detail without data should show unavailable:\n%s", frame)
	}

	m2 := testModel(140, 40)
	m2.contestsOK = true
	m2.activePanel = PanelContests
	m2 = applyKey(m2, "enter")
	if frame := stripANSI(m2.View()); !strings.Contains(frame, "No contests yet") {
		t.Errorf("detail with empty history should show empty state:\n%s", frame)
	}
}

func TestContestsFramesFit(t *testing.T) {
	for _, size := range [][2]int{{140, 40}, {80, 24}, {200, 50}} {
		width, height := size[0], size[1]
		m := contestsModel(width, height)
		m.activePanel = PanelContests
		for name, frame := range map[string]string{
			"dashboard": m.View(),
			"detail":    applyKey(m, "enter").View(),
		} {
			for i, l := range strings.Split(frame, "\n") {
				if w := lipgloss.Width(l); w > width {
					t.Errorf("%s size=%dx%d line %d overflows: %d > %d", name, width, height, i, w, width)
				}
			}
			if lines := strings.Split(frame, "\n"); len(lines) > height {
				t.Errorf("%s size=%dx%d height %d exceeds %d", name, width, height, len(lines), height)
			}
		}
	}
}
