package ui

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {

	if m.loading {
		return "\n  " + m.spinner.View() + " Loading profile..."
	}

	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	barWidth := min(40, m.width-20)

	header := m.theme.Title.Render(
		fmt.Sprintf("LeetCode Dashboard • %s", m.profile.Username),
	)

	meta := fmt.Sprintf(
		"Rank: %d   Reputation: %d",
		m.profile.Rank,
		m.profile.Reputation,
	)

	diff := renderDifficulty(m, barWidth)
	langs := renderLanguages(m, barWidth)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		meta,
		"",
		diff,
		"",
		"💻 Languages",
		langs,
		"",
		m.theme.Muted.Render("[r] refresh • [q] quit"),
	)
}

func renderDifficulty(m Model, width int) string {
	total := float64(m.profile.Difficulty.Total)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderBar("Easy", m.profile.Difficulty.Easy, total, width, m.theme.Easy),
		renderBar("Medium", m.profile.Difficulty.Medium, total, width, m.theme.Medium),
		renderBar("Hard", m.profile.Difficulty.Hard, total, width, m.theme.Hard),
	)
}

func renderLanguages(m Model, width int) string {

	if len(m.profile.Languages) == 0 {
		return ""
	}

	max := float64(m.profile.Languages[0].Count)

	var lines []string
	for i, l := range m.profile.Languages {
		if i >= 5 {
			break
		}
		ratio := float64(l.Count) / max
		lines = append(lines,
			renderCustomBar(l.Name, ratio, width),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderBar(label string, value int, total float64, width int, style lipgloss.Style) string {
	ratio := float64(value) / total
	filled := int(ratio * float64(width))

	bar := style.Render(strings.Repeat("█", filled)) +
		strings.Repeat("░", width-filled)

	return fmt.Sprintf("%-8s %s", label, bar)
}

func renderCustomBar(label string, ratio float64, width int) string {
	filled := int(ratio * float64(width))
	bar := strings.Repeat("█", filled) +
		strings.Repeat("░", width-filled)
	return fmt.Sprintf("%-12s %s", label, bar)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}