package ui

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"leetcli/internal/domain"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func problemsModel(width, height, problemCount int) Model {
	m := testModel(width, height)
	m.appView = ViewProblems
	m.problemsLoaded = true
	m.problemsTotal = problemCount

	diffs := []string{"Easy", "Medium", "Hard"}
	titles := []string{
		"Two Sum",
		"Longest Substring Without Repeating Characters",
		"A",
		"Median of Two Sorted Arrays",
		"Regular Expression Matching",
	}
	for i := 0; i < problemCount; i++ {
		status := ""
		if i%2 == 0 {
			status = "Solved"
		}
		m.problems = append(m.problems, domain.Problem{
			ID:         i + 1,
			Title:      titles[i%len(titles)] + " #" + strconv.Itoa(i+1),
			Difficulty: diffs[i%3],
			Slug:       "slug",
			Status:     status,
		})
	}
	return m
}

func applyKey(m Model, key string) Model {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	result, _ := m.Update(msg)
	return result.(Model)
}

func dataLineColumns(line string) (id, difficulty, status int, ok bool) {
	stripped := stripANSI(line)
	rest := strings.TrimPrefix(stripped, "▶ ")
	rest = strings.TrimPrefix(rest, "  ")

	if rest == "" {
		return 0, 0, 0, false
	}

	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return 0, 0, 0, false
	}
	if _, err := strconv.Atoi(fields[0]); err != nil {
		return 0, 0, 0, false
	}

	id = lipgloss.Width(stripped) - lipgloss.Width(rest)

	diffIdx := indexAny(rest, []string{"Easy", "Medium", "Hard"})
	if diffIdx < 0 {
		return 0, 0, 0, false
	}
	difficulty = id + lipgloss.Width(rest[:diffIdx])

	statusIdx := indexAny(rest, []string{"✅", "❌"})
	if statusIdx < 0 {
		return 0, 0, 0, false
	}
	status = id + lipgloss.Width(rest[:statusIdx])

	return id, difficulty, status, true
}

func indexAny(s string, subs []string) int {
	best := -1
	for _, sub := range subs {
		idx := strings.Index(s, sub)
		if idx >= 0 && (best < 0 || idx < best) {
			best = idx
		}
	}
	return best
}

func TestProblemBrowserAlignment(t *testing.T) {
	for _, size := range [][2]int{{120, 30}, {80, 20}, {200, 50}} {
		width, height := size[0], size[1]
		m := problemsModel(width, height, 30)

		frame := m.View()
		lines := strings.Split(frame, "\n")

		for i, l := range lines {
			if lipgloss.Width(l) > width {
				t.Errorf("size=%dx%d line %d exceeds width: %d > %d", width, height, i, lipgloss.Width(l), width)
			}
		}

		if len(lines) > height {
			t.Errorf("size=%dx%d frame height %d exceeds terminal height %d", width, height, len(lines), height)
		}

		var rows [][3]int
		for _, l := range lines {
			if id, diff, status, ok := dataLineColumns(l); ok {
				rows = append(rows, [3]int{id, diff, status})
			}
		}

		if len(rows) == 0 {
			t.Fatalf("size=%dx%d no data rows found", width, height)
		}

		first := rows[0]
		for i, r := range rows {
			if r != first {
				t.Errorf("size=%dx%d row %d misaligned: cols (id=%d diff=%d status=%d), want (id=%d diff=%d status=%d)",
					width, height, i, r[0], r[1], r[2], first[0], first[1], first[2])
			}
		}
	}
}

func TestProblemBrowserCursorStaysOnScreen(t *testing.T) {
	width, height := 100, 20
	m := problemsModel(width, height, 100)

	pageSize := m.pageSize()
	if pageSize != height-9 {
		t.Errorf("pageSize = %d, want %d", pageSize, height-9)
	}

	for _, cursor := range []int{0, 1, 49, 50, 98, 99} {
		m.cursor = cursor
		frame := m.View()
		lines := strings.Split(frame, "\n")

		if len(lines) > height {
			t.Errorf("cursor=%d frame height %d exceeds terminal %d", cursor, len(lines), height)
		}

		expectedID := strconv.Itoa(cursor + 1)
		found := false
		for _, l := range lines {
			if !strings.Contains(l, "▶") {
				continue
			}
			stripped := stripANSI(l)
			rest := strings.TrimPrefix(stripped, "▶ ")
			rest = strings.TrimPrefix(rest, "  ")
			fields := strings.Fields(rest)
			if len(fields) > 0 && fields[0] == expectedID {
				found = true
			}
		}
		if !found {
			t.Errorf("cursor=%d: selected row with ID %s not visible on page\nframe:\n%s", cursor, expectedID, frame)
		}
	}
}

func TestProblemBrowserPageKeys(t *testing.T) {
	m := problemsModel(100, 20, 100)
	pageSize := m.pageSize()

	m.cursor = 0
	m = applyKey(m, "]")
	if m.cursor != pageSize {
		t.Errorf("after ] cursor = %d, want %d", m.cursor, pageSize)
	}

	m = applyKey(m, "]")
	if m.cursor != pageSize*2 {
		t.Errorf("after ]] cursor = %d, want %d", m.cursor, pageSize*2)
	}

	m = applyKey(m, "[")
	if m.cursor != pageSize {
		t.Errorf("after [ cursor = %d, want %d", m.cursor, pageSize)
	}

	m = applyKey(m, "G")
	if m.cursor != 99 {
		t.Errorf("after G cursor = %d, want 99", m.cursor)
	}

	m = applyKey(m, "]")
	if m.cursor != 99 {
		t.Errorf("cursor overflow: = %d, want 99", m.cursor)
	}

	m = applyKey(m, "g")
	if m.cursor != 0 {
		t.Errorf("after g cursor = %d, want 0", m.cursor)
	}

	m = applyKey(m, "[")
	if m.cursor != 0 {
		t.Errorf("cursor underflow: = %d, want 0", m.cursor)
	}
}

func TestProblemBrowserLazyLoadTriggers(t *testing.T) {
	m := problemsModel(100, 20, 100)
	m.problemsTotal = 4073
	pageSize := m.pageSize()

	m.cursor = 0
	updated, cmd := m.maybeLoadMore()
	if cmd != nil {
		t.Errorf("cursor at start should not trigger load")
	}
	if updated.problemsLoadingMore {
		t.Errorf("problemsLoadingMore should be false")
	}

	m.cursor = len(m.problems) - pageSize
	updated, cmd = m.maybeLoadMore()
	if cmd == nil {
		t.Errorf("cursor near end should trigger load")
	}
	if !updated.problemsLoadingMore {
		t.Errorf("problemsLoadingMore should be true after trigger")
	}

	m.problemsLoadingMore = true
	_, cmd = m.maybeLoadMore()
	if cmd != nil {
		t.Errorf("should not trigger second load while one is in flight")
	}

	m.problemsLoadingMore = false
	m.problems = append(m.problems, m.problems...)
	m.problemsTotal = len(m.problems)
	m.cursor = len(m.problems) - 1
	_, cmd = m.maybeLoadMore()
	if cmd != nil {
		t.Errorf("should not load when all problems are loaded")
	}
}

func TestProblemBrowserPageAppend(t *testing.T) {
	m := problemsModel(100, 20, 100)
	m.problemsTotal = 200
	firstID := m.problems[0].ID
	lastID := m.problems[len(m.problems)-1].ID

	m.cursor = 50
	msg := problemsMsg{
		problems: []domain.Problem{
			{ID: 101, Title: "Next page", Difficulty: "Easy"},
			{ID: 102, Title: "Next page 2", Difficulty: "Hard"},
		},
		total: 200,
		skip:  100,
	}

	result, _ := m.Update(msg)
	updated := result.(Model)

	if len(updated.problems) != 102 {
		t.Fatalf("problems count after append = %d, want 102", len(updated.problems))
	}
	if updated.problems[0].ID != firstID || updated.problems[99].ID != lastID {
		t.Errorf("first page was corrupted by append")
	}
	if updated.problems[100].ID != 101 {
		t.Errorf("appended problem missing")
	}
	if updated.cursor != 50 {
		t.Errorf("cursor changed on append: = %d, want 50", updated.cursor)
	}
	if updated.problemsTotal != 200 {
		t.Errorf("problemsTotal = %d, want 200", updated.problemsTotal)
	}
	if updated.problemsLoadingMore {
		t.Errorf("problemsLoadingMore should be cleared")
	}
}

func TestProblemURL(t *testing.T) {
	got := problemURL("two-sum")
	want := "https://leetcode.com/problems/two-sum/"
	if got != want {
		t.Errorf("problemURL = %q, want %q", got, want)
	}
}

func TestProblemBrowserEnterOpensProblem(t *testing.T) {
	m := problemsModel(100, 20, 10)
	m.cursor = 3

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(Model)

	if cmd == nil {
		t.Fatal("enter should return an open-browser command")
	}
	if updated.cursor != 3 {
		t.Errorf("cursor changed on enter: = %d, want 3", updated.cursor)
	}
	if updated.appView != ViewProblems {
		t.Errorf("view changed on enter: = %v", updated.appView)
	}

	m.cursor = -1
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Errorf("enter with out-of-range cursor should not open anything")
	}

	m.cursor = len(m.problems)
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Errorf("enter past the end should not open anything")
	}

	m = problemsModel(100, 20, 10)
	m.problems[0].Slug = ""
	m.cursor = 0
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Errorf("enter with empty slug should not open anything")
	}
}

func TestProblemBrowserShortPageSetsTotal(t *testing.T) {
	m := problemsModel(100, 20, 100)
	m.problemsTotal = 0
	m.cursor = 0

	msg := problemsMsg{
		problems: []domain.Problem{{ID: 1, Title: "Only", Difficulty: "Easy"}},
		total:    0,
		skip:     0,
	}

	result, _ := m.Update(msg)
	updated := result.(Model)

	if updated.problemsTotal != 1 {
		t.Errorf("problemsTotal = %d, want 1 (short first page means end of list)", updated.problemsTotal)
	}
}
