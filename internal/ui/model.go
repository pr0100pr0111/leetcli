package ui

import (
	"context"
	"strconv"
	"strings"

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
	DetailDaily
)

type AppView int

const (
	ViewDashboard AppView = iota
	ViewProblems
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

	dailyOK bool
	daily   domain.DailyChallenge

	appView             AppView
	problems            []domain.Problem
	problemsLoaded      bool
	problemsErr         error
	problemsLoading     bool
	problemsLoadingMore bool
	problemsTotal       int
	cursor              int
	searchMode          bool
	query               string
	filtered            []int
	loadingAll          bool
}

const problemsPageSize = 100

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
		appView:     ViewDashboard,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetch(), m.fetchDaily())
}

func (m Model) fetch() tea.Cmd {
	return func() tea.Msg {
		p, err := m.service.GetProfile(context.Background())
		return profileMsg{p, err}
	}
}

func (m Model) fetchDaily() tea.Cmd {
	return func() tea.Msg {
		d, err := m.service.GetDailyChallenge(context.Background())
		return dailyMsg{d, err}
	}
}

func (m Model) fetchProblems(limit, skip int) tea.Cmd {
	return func() tea.Msg {
		problems, total, err := m.service.GetProblems(context.Background(), limit, skip)
		return problemsMsg{problems, total, skip, err}
	}
}

type profileMsg struct {
	profile domain.Profile
	err     error
}

type problemsMsg struct {
	problems []domain.Problem
	total    int
	skip     int
	err      error
}

type dailyMsg struct {
	daily domain.DailyChallenge
	err   error
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

	case dailyMsg:
		if msg.err == nil {
			m.daily = msg.daily
			m.dailyOK = true
		}
		return m, nil

	case problemsMsg:
		m.problemsLoading = false
		m.problemsLoadingMore = false
		if msg.err != nil {
			if msg.skip == 0 {
				m.problemsErr = msg.err
			}
			m.loadingAll = false
			return m, nil
		}
		m.problemsErr = nil
		if msg.skip == 0 {
			m.problems = msg.problems
			m.cursor = 0
		} else {
			m.problems = append(m.problems, msg.problems...)
		}
		if len(msg.problems) == 0 {
			m.problemsTotal = len(m.problems)
		} else if msg.total > 0 {
			m.problemsTotal = msg.total
		} else if len(msg.problems) < problemsPageSize {
			m.problemsTotal = msg.skip + len(msg.problems)
		}
		m.problemsLoaded = true
		m.refilter()
		return m, m.continueLoadAll()

	case tea.KeyMsg:
		if m.appView == ViewProblems {
			return m.updateProblemsView(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if m.detailView != DetailNone {
				m.detailView = DetailNone
				return m, nil
			}
			return m, tea.Quit
		case "r":
			m.loading = true
			cmds = append(cmds, m.fetch(), m.fetchDaily())
		case "d":
			if m.detailView == DetailNone && m.dailyOK {
				m.detailView = DetailDaily
			}
		case "t":
			m.themeIndex = (m.themeIndex + 1) % len(Themes)
			m.theme = Themes[m.themeIndex]
		case "b":
			m.appView = ViewProblems
			if !m.problemsLoaded && !m.problemsLoading {
				m.problemsLoading = true
				cmds = append(cmds, m.fetchProblems(problemsPageSize, 0))
			}
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
			} else if m.detailView == DetailDaily {
				if m.daily.Slug != "" {
					return m, openBrowser(problemURL(m.daily.Slug))
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

func (m Model) updateProblemsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchMode {
		return m.updateSearchInput(msg)
	}
	pageSize := m.pageSize()
	last := m.visibleCount() - 1

	switch msg.String() {
	case "q", "ctrl+c":
		m.leaveProblems()
		return m, nil
	case "esc":
		if m.query != "" {
			m.query = ""
			m.filtered = nil
			m.cursor = 0
			return m, nil
		}
		m.leaveProblems()
		return m, nil
	case "/":
		m.searchMode = true
		return m, m.startLoadAll()
	case "j", "down":
		if m.cursor < last {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g":
		m.cursor = 0
	case "G":
		m.cursor = last
		if m.cursor < 0 {
			m.cursor = 0
		}
	case "]", "pgdown":
		m.cursor += pageSize
		if m.cursor > last {
			m.cursor = last
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	case "[", "pgup":
		m.cursor -= pageSize
		if m.cursor < 0 {
			m.cursor = 0
		}
	case "r":
		m.problemsLoading = true
		m.problemsLoaded = false
		m.problemsTotal = 0
		return m, m.fetchProblems(problemsPageSize, 0)
	case "enter":
		if m.cursor >= 0 && m.cursor < m.visibleCount() {
			slug := m.problemAt(m.cursor).Slug
			if slug != "" {
				return m, openBrowser(problemURL(slug))
			}
		}
	}
	return m.maybeLoadMore()
}

func (m *Model) leaveProblems() {
	m.appView = ViewDashboard
	m.searchMode = false
	m.query = ""
	m.filtered = nil
	m.cursor = 0
}

func (m Model) updateSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEnter:
		m.searchMode = false
		return m, nil
	case tea.KeyEsc:
		m.searchMode = false
		m.query = ""
		m.filtered = nil
		m.cursor = 0
		return m, nil
	case tea.KeyBackspace:
		if m.query != "" {
			r := []rune(m.query)
			m.query = string(r[:len(r)-1])
			m.afterQueryChange()
		}
		return m, nil
	case tea.KeyDelete:
		m.query = ""
		m.afterQueryChange()
		return m, nil
	case tea.KeySpace:
		m.query += " "
		m.afterQueryChange()
		return m, nil
	case tea.KeyRunes:
		if len(msg.Runes) > 0 && !msg.Alt {
			m.query += string(msg.Runes)
			m.afterQueryChange()
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) afterQueryChange() {
	m.cursor = 0
	m.refilter()
}

func (m *Model) refilter() {
	if m.query == "" {
		m.filtered = nil
		return
	}
	m.filtered = filterIndices(m.problems, m.query)
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func filterIndices(problems []domain.Problem, query string) []int {
	q := strings.ToLower(query)
	out := make([]int, 0, len(problems))
	for i, p := range problems {
		if strings.Contains(strings.ToLower(p.Title), q) ||
			strings.Contains(strings.ToLower(p.Slug), q) ||
			strings.Contains(strconv.Itoa(p.ID), q) {
			out = append(out, i)
		}
	}
	return out
}

func (m Model) visibleCount() int {
	if m.filtered != nil {
		return len(m.filtered)
	}
	return len(m.problems)
}

func (m Model) problemAt(i int) domain.Problem {
	if m.filtered != nil {
		if i < 0 || i >= len(m.filtered) {
			return domain.Problem{}
		}
		return m.problems[m.filtered[i]]
	}
	if i < 0 || i >= len(m.problems) {
		return domain.Problem{}
	}
	return m.problems[i]
}

func (m *Model) startLoadAll() tea.Cmd {
	if !m.problemsLoaded {
		if !m.problemsLoading {
			m.problemsLoading = true
			m.loadingAll = true
			return m.fetchProblems(problemsPageSize, 0)
		}
		m.loadingAll = true
		return nil
	}
	if m.problemsTotal > len(m.problems) && !m.problemsLoadingMore && !m.loadingAll {
		m.loadingAll = true
		m.problemsLoadingMore = true
		return m.fetchProblems(problemsPageSize, len(m.problems))
	}
	return nil
}

func (m *Model) continueLoadAll() tea.Cmd {
	if !m.loadingAll {
		return nil
	}
	if m.problemsTotal <= len(m.problems) {
		m.loadingAll = false
		m.problemsLoadingMore = false
		return nil
	}
	m.problemsLoadingMore = true
	return m.fetchProblems(problemsPageSize, len(m.problems))
}

func (m Model) maybeLoadMore() (Model, tea.Cmd) {
	if m.problemsLoading || m.problemsLoadingMore || !m.problemsLoaded {
		return m, nil
	}
	if m.problemsTotal > 0 && len(m.problems) >= m.problemsTotal {
		return m, nil
	}
	if len(m.problems) == 0 {
		return m, nil
	}
	if m.cursor < len(m.problems)-m.pageSize() {
		return m, nil
	}
	m.problemsLoadingMore = true
	skip := len(m.problems)
	return m, m.fetchProblems(problemsPageSize, skip)
}

func (m Model) pageSize() int {
	p := m.height - 9
	if p < 1 {
		p = 1
	}
	return p
}

func (m Model) nextTheme() Theme {
	m.themeIndex = (m.themeIndex + 1) % len(Themes)
	return Themes[m.themeIndex]
}
