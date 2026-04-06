package ui

import (
	"context"
	"leetcli/internal/config"
	"leetcli/internal/domain"
	"leetcli/internal/service"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)

type Panel int

const (
	PanelDifficulty Panel = iota
	PanelLanguages
	PanelSkills
	PanelContests
	PanelCount
)

type Model struct {
	service       *service.ProfileService
	profile       domain.Profile
	err           error
	width         int
	height        int

	spinner       spinner.Model
	loading       bool

	theme         Theme
	themeIndex    int
	activePanel   Panel
	showHelp      bool
	config        config.Config
}

func NewModel(s *service.ProfileService, cfg config.Config) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	theme := GetTheme(cfg.Theme)
	themeIndex := 0
	for i, t := range Themes {
		if t.Name == cfg.Theme {
			themeIndex = i
			break
		}
	}

	return Model{
		service:     s,
		spinner:     sp,
		loading:     true,
		theme:       theme,
		themeIndex:  themeIndex,
		activePanel: PanelDifficulty,
		showHelp:    true,
		config:      cfg,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetch())
}

func (m Model) fetch() tea.Cmd {
	return func() tea.Msg {
		p, err := m.service.GetProfile(context.Background())
		return profileMsg{p, err}
	}
}

type profileMsg struct {
	profile domain.Profile
	err     error
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case profileMsg:
		m.loading = false
		m.profile = msg.profile
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			cmds = append(cmds, m.fetch())
		case "t":
			m.themeIndex = (m.themeIndex + 1) % len(Themes)
			m.theme = Themes[m.themeIndex]
		case "tab", "right":
			m.activePanel = (m.activePanel + 1) % PanelCount
		case "shift+tab", "left":
			m.activePanel = (m.activePanel - 1 + PanelCount) % PanelCount
		case "h", "?":
			m.showHelp = !m.showHelp
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) nextTheme() Theme {
	m.themeIndex = (m.themeIndex + 1) % len(Themes)
	return Themes[m.themeIndex]
}