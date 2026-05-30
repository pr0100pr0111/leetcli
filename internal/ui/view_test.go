package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"leetcli/internal/config"
	"leetcli/internal/domain"
)

func testModel(width, height int) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := Model{
		spinner:     sp,
		loading:     false,
		theme:       DefaultTheme(),
		themeIndex:  0,
		activePanel: PanelDifficulty,
		detailView:  DetailNone,
		showHelp:    true,
		config:      config.Config{Theme: "default"},
		width:       width,
		height:      height,
	}
	m.profile = domain.Profile{
		Username:       "tester",
		Rank:           12345,
		Reputation:     678,
		Streak:         42,
		Difficulty:     domain.Difficulty{Easy: 120, Medium: 60, Hard: 20, Total: 200},
		LeetCodeTotals: domain.LeetCodeTotals{Easy: 800, Medium: 1600, Hard: 700},
		Languages: []domain.Language{
			{Name: "Go", Count: 100},
			{Name: "Python", Count: 60},
			{Name: "JavaScript", Count: 40},
		},
		Skills: []domain.Skill{
			{Name: "Array", Count: 100},
			{Name: "String", Count: 80},
			{Name: "Dynamic Programming", Count: 50},
			{Name: "Binary Search", Count: 40},
			{Name: "Graph", Count: 30},
		},
	}
	return m
}

func checkFrame(t *testing.T, name, frame string, width int) {
	t.Helper()
	lines := strings.Split(frame, "\n")
	for i, l := range lines {
		w := lipgloss.Width(l)
		if w > width {
			t.Errorf("%s: line %d wider than terminal: %d > %d: %q", name, i, w, width, l)
		}
	}
}

func TestPanelsVisibleAfterDetailExit(t *testing.T) {
	for _, width := range []int{120, 160, 200, 250} {
		m := testModel(width, 40)

		main := m.View()
		checkFrame(t, "main-initial", main, width)

		for _, title := range []string{"Difficulty", "Languages", "Skills"} {
			if !strings.Contains(main, title) {
				t.Errorf("width=%d initial main missing %q", width, title)
			}
		}

		m.activePanel = PanelLanguages
		m.detailView = DetailLanguages
		detail := m.View()
		checkFrame(t, "detail-langs", detail, width)

		m.detailView = DetailNone
		after := m.View()
		checkFrame(t, "main-after-detail", after, width)

		for _, title := range []string{"Difficulty", "Languages", "Skills"} {
			if !strings.Contains(after, title) {
				t.Errorf("width=%d main after langs detail missing %q\n---frame---\n%s", width, title, after)
			}
		}

		m.activePanel = PanelSkills
		m.detailView = DetailSkills
		_ = m.View()
		m.detailView = DetailNone
		after2 := m.View()
		checkFrame(t, "main-after-skill-detail", after2, width)
		for _, title := range []string{"Difficulty", "Languages", "Skills"} {
			if !strings.Contains(after2, title) {
				t.Errorf("width=%d main after skills detail missing %q\n---frame---\n%s", width, title, after2)
			}
		}
	}
}
