package ui

import (
	"context"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"leetcli/internal/config"
	"leetcli/internal/domain"
	"leetcli/internal/service"
)

type Panel int

const (
	PanelDifficulty Panel = iota
	PanelLanguages
	PanelSkills
	PanelContests
	PanelCount
)

type DetailView int

const (
	DetailNone DetailView = iota
	DetailDifficulty
	DetailLanguages
	DetailSkills
)

type Model struct {
	service *service.ProfileService
	profile domain.Profile
	err     error
	width   int
	height  int

	spinner spinner.Model
	loading bool

	theme       Theme
	themeIndex  int
	activePanel Panel
	detailView  DetailView
	showHelp    bool
	config      config.Config
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
		detailView:  DetailNone,
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
			if m.detailView != DetailNone {
				m.detailView = DetailNone
				return m, nil
			}
			return m, tea.Quit
		case "r":
			m.loading = true
			cmds = append(cmds, m.fetch())
		case "t":
			m.themeIndex = (m.themeIndex + 1) % len(Themes)
			m.theme = Themes[m.themeIndex]
		case "tab", "right":
			if m.detailView == DetailNone {
				m.activePanel = (m.activePanel + 1) % PanelCount
			}
		case "shift+tab", "left":
			if m.detailView == DetailNone {
				m.activePanel = (m.activePanel - 1 + PanelCount) % PanelCount
			}
		case "enter":
			if m.detailView == DetailNone {
				switch m.activePanel {
				case PanelDifficulty:
					m.detailView = DetailDifficulty
				case PanelLanguages:
					m.detailView = DetailLanguages
				case PanelSkills:
					m.detailView = DetailSkills
				}
			} else {
				m.detailView = DetailNone
			}
		case "esc":
			m.detailView = DetailNone
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
