package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name        string
	Title       lipgloss.Style
	Easy        lipgloss.Style
	Medium      lipgloss.Style
	Hard        lipgloss.Style
	Muted       lipgloss.Style
	Accent      lipgloss.Style
	Selected    lipgloss.Style
	PanelBorder string
	AccentHex   string
}

func (t Theme) panelStyle(width int, active bool) lipgloss.Style {
	border := t.PanelBorder
	if active {
		border = t.AccentHex
	}
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(border)).
		Padding(0, 1)
	if width > 0 {
		s = s.Width(width)
	}
	return s
}

func DefaultTheme() Theme {
	return Theme{
		Name:        "default",
		Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C5CFF")),
		Easy:        lipgloss.NewStyle().Foreground(lipgloss.Color("#00C896")),
		Medium:      lipgloss.NewStyle().Foreground(lipgloss.Color("#FFC107")),
		Hard:        lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4D4F")),
		Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("#888")),
		Accent:      lipgloss.NewStyle().Foreground(lipgloss.Color("#7C5CFF")),
		Selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("#7C5CFF")).Bold(true),
		PanelBorder: "#444",
		AccentHex:   "#7C5CFF",
	}
}

func DraculaTheme() Theme {
	return Theme{
		Name:        "dracula",
		Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9")),
		Easy:        lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
		Medium:      lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")),
		Hard:        lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
		Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")),
		Accent:      lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")),
		Selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true),
		PanelBorder: "#BD93F9",
		AccentHex:   "#BD93F9",
	}
}

func NordTheme() Theme {
	return Theme{
		Name:        "nord",
		Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#88C0D0")),
		Easy:        lipgloss.NewStyle().Foreground(lipgloss.Color("#A3BE8C")),
		Medium:      lipgloss.NewStyle().Foreground(lipgloss.Color("#EBCB8B")),
		Hard:        lipgloss.NewStyle().Foreground(lipgloss.Color("#BF616A")),
		Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("#4C566A")),
		Accent:      lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")),
		Selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("#88C0D0")).Bold(true),
		PanelBorder: "#4C566A",
		AccentHex:   "#88C0D0",
	}
}

func TokyoNightTheme() Theme {
	return Theme{
		Name:        "tokyo-night",
		Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7AA2F7")),
		Easy:        lipgloss.NewStyle().Foreground(lipgloss.Color("#9ECE6A")),
		Medium:      lipgloss.NewStyle().Foreground(lipgloss.Color("#E0AF68")),
		Hard:        lipgloss.NewStyle().Foreground(lipgloss.Color("#F7768E")),
		Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("#565F89")),
		Accent:      lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")),
		Selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7")).Bold(true),
		PanelBorder: "#7AA2F7",
		AccentHex:   "#7AA2F7",
	}
}

func CatppuccinTheme() Theme {
	return Theme{
		Name:        "catppuccin",
		Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CBA6F7")),
		Easy:        lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")),
		Medium:      lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF")),
		Hard:        lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")),
		Muted:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")),
		Accent:      lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")),
		Selected:    lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Bold(true),
		PanelBorder: "#CBA6F7",
		AccentHex:   "#CBA6F7",
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
