package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"leetcli/internal/domain"
)

func dailyModel(width, height int) Model {
	m := testModel(width, height)
	m.daily = domain.DailyChallenge{
		Date:       "2026-10-07",
		Slug:       "remove-invalid-parentheses",
		ID:         301,
		Title:      "Remove Invalid Parentheses",
		Difficulty: "Hard",
		Topics: []domain.Topic{
			{Name: "String", Slug: "string"},
			{Name: "Backtracking", Slug: "backtracking"},
		},
	}
	m.dailyOK = true
	return m
}

func TestDailyChallengeDashboardLine(t *testing.T) {
	m := dailyModel(120, 30)
	frame := stripANSI(m.View())
	if !strings.Contains(frame, "Daily #301") {
		t.Errorf("dashboard should show daily id:\n%s", frame)
	}
	if !strings.Contains(frame, "Remove Invalid Parentheses") {
		t.Errorf("dashboard should show daily title:\n%s", frame)
	}
	if !strings.Contains(frame, "(Hard)") {
		t.Errorf("dashboard should show daily difficulty:\n%s", frame)
	}

	m2 := dailyModel(120, 30)
	m2.dailyOK = false
	if strings.Contains(stripANSI(m2.View()), "Daily #") {
		t.Error("dashboard should hide daily line when unavailable")
	}
}

func TestDailyChallengeDetail(t *testing.T) {
	m := dailyModel(120, 30)
	m = applyKey(m, "d")
	if m.detailView != DetailDaily {
		t.Fatalf("detailView = %v, want DetailDaily", m.detailView)
	}

	frame := stripANSI(m.View())
	for _, want := range []string{
		"Daily Challenge",
		"#301",
		"Remove Invalid Parentheses",
		"Hard",
		"2026-10-07",
		"String, Backtracking",
		"leetcode.com/problems/remove-invalid-parentheses/",
		"[enter] open",
	} {
		if !strings.Contains(frame, want) {
			t.Errorf("detail frame missing %q:\n%s", want, frame)
		}
	}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("enter should open the daily problem in the browser")
	}
	updated := result.(Model)
	if updated.detailView != DetailDaily {
		t.Errorf("enter should keep the detail view: %v", updated.detailView)
	}

	updated = pressKey(updated, tea.KeyEsc)
	if updated.detailView != DetailNone {
		t.Errorf("esc should close the detail view: %v", updated.detailView)
	}
}

func TestDailyChallengeKeyWithoutData(t *testing.T) {
	m := testModel(120, 30)
	m = applyKey(m, "d")
	if m.detailView != DetailNone {
		t.Errorf("d without daily data should do nothing: %v", m.detailView)
	}
}

func TestDailyChallengeDashToggleGuard(t *testing.T) {
	m := dailyModel(120, 30)
	m.detailView = DetailDifficulty
	m = applyKey(m, "d")
	if m.detailView != DetailDifficulty {
		t.Errorf("d inside another detail view should do nothing: %v", m.detailView)
	}
}

func TestDailyChallengeFramesFit(t *testing.T) {
	for _, size := range [][2]int{{120, 30}, {80, 20}, {200, 50}} {
		width, height := size[0], size[1]

		m := dailyModel(width, height)
		for name, frame := range map[string]string{
			"dashboard": m.View(),
			"detail":    applyKey(m, "d").View(),
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
