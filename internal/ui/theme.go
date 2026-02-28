package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Title  lipgloss.Style
	Easy   lipgloss.Style
	Medium lipgloss.Style
	Hard   lipgloss.Style
	Muted  lipgloss.Style
}

func DefaultTheme() Theme {
	return Theme{
		Title:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C5CFF")),
		Easy:   lipgloss.NewStyle().Foreground(lipgloss.Color("#00C896")),
		Medium: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFC107")),
		Hard:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4D4F")),
		Muted:  lipgloss.NewStyle().Foreground(lipgloss.Color("#888")),
	}
}