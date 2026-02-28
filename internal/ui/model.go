package ui

import (
	"context"
	"leetcli/internal/domain"
	"leetcli/internal/service"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
)

type Model struct {
	service *service.ProfileService
	profile domain.Profile
	err     error
	width   int

	spinner spinner.Model
	loading bool

	theme Theme
}

func NewModel(s *service.ProfileService) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return Model{
		service: s,
		spinner: sp,
		loading: true,
		theme:   DefaultTheme(),
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

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case profileMsg:
		m.loading = false
		m.profile = msg.profile
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, m.fetch()
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}