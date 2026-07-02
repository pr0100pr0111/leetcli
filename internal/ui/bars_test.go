package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func countStart(row string) int {
	plain := strings.TrimRight(stripANSI(row), " ")
	i := len(plain)
	for i > 0 && plain[i-1] >= '0' && plain[i-1] <= '9' {
		i--
	}
	return lipgloss.Width(plain[:i])
}

func TestBarColumnsAligned(t *testing.T) {
	old := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(old)

	m := testModel(200, 45)
	total := m.profile.Difficulty.Total

	for name, content := range map[string]string{
		"difficulty": lipgloss.JoinVertical(lipgloss.Left,
			renderBar("Easy", m.profile.Difficulty.Easy, float64(total), 34, m.theme.Easy, m.theme.Muted),
			renderBar("Medium", m.profile.Difficulty.Medium, float64(total), 34, m.theme.Medium, m.theme.Muted),
			renderBar("Hard", m.profile.Difficulty.Hard, float64(total), 34, m.theme.Hard, m.theme.Muted)),
		"languages": renderLanguages(m, 34, total, 8),
		"skills":    renderSkills(m, 34, total, 8),
	} {
		starts := map[int]bool{}
		for _, row := range strings.Split(content, "\n") {
			if !strings.Contains(stripANSI(row), "%") {
				continue
			}
			starts[countStart(row)] = true
			if !strings.Contains(row, "\x1b[") {
				t.Errorf("%s row has no ANSI color: %q", name, row)
			}
		}
		if len(starts) != 1 {
			t.Errorf("%s: count column not aligned, positions: %v", name, starts)
		}
	}
}

func TestBarLabelTruncatedToColumn(t *testing.T) {
	m := testModel(200, 45)
	total := m.profile.Difficulty.Total

	content := renderSkills(m, 34, total, 8)
	labelW := barLabelWidth([]string{"Dynamic Programming", "Bit Manipulation"}, 14, 20, 34)
	for _, row := range strings.Split(content, "\n") {
		plain := stripANSI(row)
		if !strings.Contains(plain, "█") && !strings.Contains(plain, "░") {
			continue
		}
		rest := plain
		if idx := strings.Index(rest, "█"); idx >= 0 {
			rest = rest[:idx]
		} else if idx := strings.Index(rest, "░"); idx >= 0 {
			rest = rest[:idx]
		}
		if got := lipgloss.Width(strings.TrimRight(rest, " ")); got > labelW {
			t.Errorf("label column overflows: %d > %d: %q", got, labelW, plain)
		}
	}
}
