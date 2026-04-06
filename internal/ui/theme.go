package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name   string
	Title  lipgloss.Style
	Easy   lipgloss.Style
	Medium lipgloss.Style
	Hard   lipgloss.Style
	Muted  lipgloss.Style
	Accent lipgloss.Style
	Panel  lipgloss.Style
	Selected lipgloss.Style
}

func DefaultTheme() Theme {
	return Theme{
		Name:   "default",
		Title:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C5CFF")),
		Easy:   lipgloss.NewStyle().Foreground(lipgloss.Color("#00C896")),
		Medium: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFC107")),
		Hard:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4D4F")),
		Muted:  lipgloss.NewStyle().Foreground(lipgloss.Color("#888")),
		Accent: lipgloss.NewStyle().Foreground(lipgloss.Color("#7C5CFF")),
		Panel:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#444")).Padding(0, 1),
		Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#7C5CFF")).Bold(true),
	}
}

func DraculaTheme() Theme {
	return Theme{
		Name:   "dracula",
		Title:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")),
		Easy:   lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
		Medium: lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")),
		Hard:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
		Muted:  lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		Accent: lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
		Panel:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#BD93F9")).Padding(0, 1),
		Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true),
	}
}

func NordTheme() Theme {
	return Theme{
		Name:   "nord",
		Title:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#88C0D0")),
		Easy:   lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
		Medium: lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")),
		Hard:   lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),
		Muted:  lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
		Accent: lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")),
		Panel:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4C566A")).Padding(0, 1),
		Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Bold(true),
	}
}

func TokyoNightTheme() Theme {
	return Theme{
		Name:   "tokyo-night",
		Title:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7AA2F7")),
		Easy:   lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")),
		Medium: lipgloss.NewStyle().Foreground(lipgloss.Color("#E0AF68")),
		Hard:   lipgloss.NewStyle().Foreground(lipgloss.Color("#F7768E")),
		Muted:  lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")),
		Accent: lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")),
		Panel:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7AA2F7")).Padding(0, 1),
		Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")).Bold(true),
	}
}

func CatppuccinTheme() Theme {
	return Theme{
		Name:   "catppuccin",
		Title:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CBA6F7")),
		Easy:   lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")),
		Medium: lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF")),
		Hard:   lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")),
		Muted:  lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")),
		Accent: lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")),
		Panel:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#CBA6F7")).Padding(0, 1),
		Selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Bold(true),
	}
}

func GetTheme(name string) Theme {
	switch name {
	case "dracula":
		return DraculaTheme()
	case "nord":
		return NordTheme()
	case "tokyo-night":
		return TokyoNightTheme()
	case "catppuccin":
		return CatppuccinTheme()
	default:
		return DefaultTheme()
	}
}

var Themes = []Theme{
	DefaultTheme(),
	DraculaTheme(),
	NordTheme(),
	TokyoNightTheme(),
	CatppuccinTheme(),
}