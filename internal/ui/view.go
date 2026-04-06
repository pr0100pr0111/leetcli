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
		return m.theme.Muted.Render(fmt.Sprintf("Error: %v", m.err))
	}

	barWidth := min(40, m.width-20)

	header := m.theme.Title.Render(
		fmt.Sprintf("LeetCode Dashboard • %s", m.profile.Username),
	)

	themeBadge := m.theme.Muted.Render(fmt.Sprintf("theme: %s", m.theme.Name))
	meta := fmt.Sprintf(
		"Rank: %d   Reputation: %d   Streak: %d days",
		m.profile.Rank,
		m.profile.Reputation,
		m.profile.Streak,
	)

	diff := m.renderPanel("📊 Difficulty", renderDifficulty(m, barWidth), PanelDifficulty)
	langs := m.renderPanel("💻 Languages", renderLanguages(m, barWidth), PanelLanguages)
	skills := m.renderPanel("🏷️  Skills", renderSkills(m, barWidth), PanelSkills)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, diff, "  ", langs, "  ", skills)

	help := ""
	if m.showHelp {
		help = m.theme.Muted.Render("[r] refresh  [t] theme  [tab/←→] panel  [h] help  [q] quit")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		themeBadge,
		meta,
		"",
		panels,
		"",
		help,
	)
}

func (m Model) renderPanel(title, content string, panel Panel) string {
	style := m.theme.Panel
	if m.activePanel == panel {
		style = style.BorderForeground(lipgloss.Color(m.getAccentColor()))
	}
	titleStyle := m.theme.Muted
	if m.activePanel == panel {
		titleStyle = m.theme.Selected
	}
	return style.Render(titleStyle.Render(title) + "\n" + content)
}

func (m Model) getAccentColor() string {
	switch m.theme.Name {
	case "dracula":
		return "#BD93F9"
	case "nord":
		return "#88C0D0"
	case "tokyo-night":
		return "#7AA2F7"
	case "catppuccin":
		return "#CBA6F7"
	default:
		return "#7C5CFF"
	}
}

func renderDifficulty(m Model, width int) string {
	total := float64(m.profile.Difficulty.Total)
	if total == 0 {
		return m.theme.Muted.Render("No data")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderBar("Easy", m.profile.Difficulty.Easy, total, width, m.theme.Easy),
		renderBar("Medium", m.profile.Difficulty.Medium, total, width, m.theme.Medium),
		renderBar("Hard", m.profile.Difficulty.Hard, total, width, m.theme.Hard),
	)
}

func renderLanguages(m Model, width int) string {
	if len(m.profile.Languages) == 0 {
		return m.theme.Muted.Render("No data")
	}

	max := float64(m.profile.Languages[0].Count)
	if max == 0 {
		return m.theme.Muted.Render("No data")
	}

	var lines []string
	for i, l := range m.profile.Languages {
		if i >= 8 {
			break
		}
		ratio := float64(l.Count) / max
		lines = append(lines, renderCustomBar(l.Name, ratio, width))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderSkills(m Model, width int) string {
	if len(m.profile.Skills) == 0 {
		return m.theme.Muted.Render("No data")
	}

	max := float64(m.profile.Skills[0].Count)
	if max == 0 {
		return m.theme.Muted.Render("No data")
	}

	var lines []string
	for i, s := range m.profile.Skills {
		if i >= 8 {
			break
		}
		ratio := float64(s.Count) / max
		lines = append(lines, renderCustomBar(s.Name, ratio, width))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderBar(label string, value int, total float64, width int, style lipgloss.Style) string {
	ratio := float64(value) / total
	filled := int(ratio * float64(width))

	bar := style.Render(strings.Repeat("█", filled)) +
		strings.Repeat("░", width-filled)

	pct := fmt.Sprintf("%.0f%%", ratio*100)
	return fmt.Sprintf("%-8s %s %5s %d", label, bar, pct, value)
}

func renderCustomBar(label string, ratio float64, width int) string {
	filled := int(ratio * float64(width))
	bar := strings.Repeat("█", filled) +
		strings.Repeat("░", width-filled)
	pct := fmt.Sprintf("%.0f%%", ratio*100)
	return fmt.Sprintf("%-14s %s %s", label, bar, pct)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
