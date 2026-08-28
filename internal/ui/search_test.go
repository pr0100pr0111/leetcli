package ui

import (
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"leetcli/internal/domain"
)

func pressKey(m Model, typ tea.KeyType) Model {
	result, _ := m.Update(tea.KeyMsg{Type: typ})
	return result.(Model)
}

func TestProblemBrowserSearchFilters(t *testing.T) {
	m := problemsModel(100, 20, 100)

	m = applyKey(m, "/")
	if !m.searchMode {
		t.Fatal("/ should enter search mode")
	}

	m = applyKey(m, "two")
	if m.query != "two" {
		t.Errorf("query = %q, want two", m.query)
	}
	if m.filtered == nil {
		t.Fatal("filtered should be set")
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}
	if m.visibleCount() != 40 {
		t.Fatalf("visibleCount = %d, want 40", m.visibleCount())
	}
	for i := 0; i < m.visibleCount(); i++ {
		p := m.problemAt(i)
		if !strings.Contains(strings.ToLower(p.Title), "two") {
			t.Errorf("visible %d = %q, title should contain two", i, p.Title)
		}
	}

	frame := m.View()
	stripped := stripANSI(frame)
	if !strings.Contains(stripped, "40 matches") {
		t.Errorf("frame should show match count:\n%s", frame)
	}
	if lines := strings.Split(frame, "\n"); len(lines) > 20 {
		t.Errorf("frame height = %d, want <= 20", len(lines))
	}
}

func TestProblemBrowserSearchModeKeys(t *testing.T) {
	m := problemsModel(100, 20, 10)
	m = applyKey(m, "/")

	m = applyKey(m, "фыва")
	if m.query != "фыва" {
		t.Fatalf("query = %q", m.query)
	}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = result.(Model)
	if m.query != "фыв" {
		t.Errorf("backspace should remove one rune: query = %q", m.query)
	}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = result.(Model)
	if cmd != nil {
		t.Error("space should not quit or navigate")
	}
	if m.query != "фыв " {
		t.Errorf("space: query = %q", m.query)
	}

	m = applyKey(m, "q")
	if m.appView != ViewProblems {
		t.Error("q typed in search mode should not leave the browser")
	}
	if !strings.HasSuffix(m.query, "q") {
		t.Errorf("q should be typed into query: %q", m.query)
	}

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should quit from search mode")
	}

	m = pressKey(m, tea.KeyEsc)
	if m.searchMode {
		t.Error("esc should leave search mode")
	}
	if m.query != "" || m.filtered != nil {
		t.Errorf("esc should clear query: %q", m.query)
	}
	if m.visibleCount() != 10 {
		t.Errorf("visibleCount after clear = %d, want 10", m.visibleCount())
	}

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(Model)
	if m.appView != ViewDashboard {
		t.Error("second esc should return to dashboard")
	}
}

func TestProblemBrowserSearchEnterKeepsFilter(t *testing.T) {
	m := problemsModel(100, 20, 100)
	m = applyKey(m, "/")
	m = applyKey(m, "two")
	m = pressKey(m, tea.KeyEnter)

	if m.searchMode {
		t.Error("enter should leave search mode")
	}
	if m.query != "two" || m.filtered == nil {
		t.Errorf("enter should keep filter: query=%q filtered=%v", m.query, m.filtered)
	}
	if m.visibleCount() != 40 {
		t.Errorf("visibleCount = %d, want 40", m.visibleCount())
	}

	frame := stripANSI(m.View())
	if !strings.Contains(frame, "filter: two") {
		t.Errorf("frame should show active filter:\n%s", frame)
	}
	if !strings.Contains(frame, "[esc] clear") {
		t.Errorf("frame should show clear hint:\n%s", frame)
	}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("enter with active filter should open the selected problem")
	}
	updated := result.(Model)
	if updated.query != "two" {
		t.Errorf("opening should keep the filter: query = %q", updated.query)
	}

	m = applyKey(m, "/")
	if !m.searchMode {
		t.Error("/ should re-enter search mode with existing query")
	}
	if m.query != "two" {
		t.Errorf("query after re-enter = %q", m.query)
	}
}

func TestProblemBrowserEscClearsFilterBeforeLeaving(t *testing.T) {
	m := problemsModel(100, 20, 100)
	m = applyKey(m, "/")
	m = applyKey(m, "median")
	m = pressKey(m, tea.KeyEnter)
	if m.visibleCount() == 0 || m.visibleCount() == 100 {
		t.Fatalf("precondition: visibleCount = %d", m.visibleCount())
	}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(Model)
	if m.appView != ViewProblems {
		t.Error("esc with active filter should stay in the browser")
	}
	if m.query != "" || m.visibleCount() != 100 {
		t.Errorf("esc should clear filter: query=%q visible=%d", m.query, m.visibleCount())
	}

	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = result.(Model)
	if m.appView != ViewDashboard {
		t.Error("esc without filter should return to dashboard")
	}
}

func TestProblemBrowserSearchNoMatches(t *testing.T) {
	m := problemsModel(100, 20, 100)
	m = applyKey(m, "/")
	m = applyKey(m, "zzzzz")

	if m.visibleCount() != 0 {
		t.Fatalf("visibleCount = %d, want 0", m.visibleCount())
	}
	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.cursor)
	}

	frame := stripANSI(m.View())
	if !strings.Contains(frame, `No problems match "zzzzz"`) {
		t.Errorf("frame should show no-match state:\n%s", frame)
	}
	if lines := strings.Split(m.View(), "\n"); len(lines) > 20 {
		t.Errorf("no-match frame height = %d, want <= 20", len(lines))
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter with no matches should not open anything")
	}

	m = applyKey(m, "j")
	m = applyKey(m, "k")
	if m.cursor != 0 {
		t.Errorf("j/k with no matches: cursor = %d", m.cursor)
	}
}

func TestProblemBrowserSearchMatchesIDAndSlug(t *testing.T) {
	m := problemsModel(100, 20, 100)
	m = applyKey(m, "/")
	m = applyKey(m, "15")

	foundID15 := false
	for i := 0; i < m.visibleCount(); i++ {
		if m.problemAt(i).ID == 15 {
			foundID15 = true
		}
	}
	if !foundID15 {
		t.Error("query 15 should match problem with ID 15")
	}

	m.query = ""
	m.refilter()
	m = applyKey(m, "two sum #11")
	if m.visibleCount() != 1 {
		t.Fatalf("query by full title: visibleCount = %d, want 1", m.visibleCount())
	}
	if got := m.problemAt(0).ID; got != 11 {
		t.Errorf("matched ID = %d, want 11", got)
	}

	m.query = "SLUG"
	m.refilter()
	if m.visibleCount() != 100 {
		t.Errorf("slug query should match all: visibleCount = %d", m.visibleCount())
	}
}

func TestProblemBrowserLoadAllChain(t *testing.T) {
	m := problemsModel(100, 20, 100)
	m.problemsTotal = 4073

	m = applyKey(m, "/")
	if !m.loadingAll || !m.problemsLoadingMore {
		t.Fatalf("search should start bulk load: loadingAll=%v loadingMore=%v", m.loadingAll, m.problemsLoadingMore)
	}

	page := make([]domain.Problem, 100)
	for i := range page {
		page[i] = domain.Problem{ID: 101 + i, Title: "P " + strconv.Itoa(i), Difficulty: "Easy", Slug: "slug"}
	}
	result, cmd := m.Update(problemsMsg{problems: page, total: 4073, skip: 100})
	m = result.(Model)
	if cmd == nil {
		t.Fatal("chain should continue while not all problems are loaded")
	}
	if !m.loadingAll || !m.problemsLoadingMore {
		t.Errorf("chain state: loadingAll=%v loadingMore=%v", m.loadingAll, m.problemsLoadingMore)
	}
	if len(m.problems) != 200 {
		t.Errorf("loaded = %d, want 200", len(m.problems))
	}

	result, cmd = m.Update(problemsMsg{problems: nil, total: 0, skip: 200})
	m = result.(Model)
	if cmd != nil {
		t.Error("empty page should stop the chain")
	}
	if m.loadingAll || m.problemsLoadingMore {
		t.Errorf("chain should stop: loadingAll=%v loadingMore=%v", m.loadingAll, m.problemsLoadingMore)
	}
}

func TestProblemBrowserLoadAllCompletes(t *testing.T) {
	m := problemsModel(100, 20, 100)
	m.problemsTotal = 200

	m = applyKey(m, "/")
	page := make([]domain.Problem, 100)
	for i := range page {
		page[i] = domain.Problem{ID: 101 + i, Title: "P", Difficulty: "Easy", Slug: "slug"}
	}
	result, cmd := m.Update(problemsMsg{problems: page, total: 200, skip: 100})
	m = result.(Model)
	if cmd != nil {
		t.Error("chain should finish when total is reached")
	}
	if m.loadingAll || m.problemsLoadingMore {
		t.Errorf("flags after finish: loadingAll=%v loadingMore=%v", m.loadingAll, m.problemsLoadingMore)
	}
	if m.problemsTotal != 200 {
		t.Errorf("problemsTotal = %d, want 200", m.problemsTotal)
	}
}

func TestProblemBrowserFilterFrameFits(t *testing.T) {
	for _, size := range [][2]int{{120, 30}, {80, 20}, {200, 50}} {
		width, height := size[0], size[1]
		m := problemsModel(width, height, 100)

		for _, state := range []string{"typing", "active"} {
			m2 := applyKey(m, "/")
			m2 = applyKey(m2, "two")
			if state == "active" {
				m2 = pressKey(m2, tea.KeyEnter)
			}

			frame := m2.View()
			lines := strings.Split(frame, "\n")
			if len(lines) > height {
				t.Errorf("%s size=%dx%d frame height %d exceeds %d", state, width, height, len(lines), height)
			}
			for i, l := range lines {
				if lipgloss.Width(l) > width {
					t.Errorf("%s size=%dx%d line %d exceeds width: %d > %d", state, width, height, i, lipgloss.Width(l), width)
				}
			}

			var rows [][3]int
			for _, l := range lines {
				if id, diff, status, ok := dataLineColumns(l); ok {
					rows = append(rows, [3]int{id, diff, status})
				}
			}
			if len(rows) == 0 {
				t.Fatalf("%s size=%dx%d no data rows", state, width, height)
			}
			first := rows[0]
			for i, r := range rows {
				if r != first {
					t.Errorf("%s size=%dx%d row %d misaligned: %v vs %v", state, width, height, i, r, first)
				}
			}
		}
	}
}
