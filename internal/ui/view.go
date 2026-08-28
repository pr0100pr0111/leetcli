package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"leetcli/internal/domain"
)

func fit(s string, width int) string {
	if width <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > width {
			lines[i] = truncate.String(line, uint(width))
		}
	}
	return strings.Join(lines, "\n")
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (m Model) View() string {
	if m.loading {
		return "\n  " + m.spinner.View() + " Loading profile..."
	}

	if m.err != nil {
		return fit(m.theme.Muted.Render(fmt.Sprintf("Error: %v", m.err)), m.width)
	}

	if m.appView == ViewProblems {
		return m.renderProblemsView()
	}

	if m.detailView != DetailNone {
		return fit(m.renderDetailView(), m.width)
	}

	width := m.width
	if width <= 0 {
		width = 80
	}
	height := m.height
	if height <= 0 {
		height = 24
	}

	useHorizontal := width >= 145

	var barWidth int
	var maxItems int
	if useHorizontal {
		barWidth = (width - 97) / 3
		maxItems = clamp(height-10, 3, 8)
	} else {
		barWidth = width - 35
		maxItems = clamp((height-20)/2, 3, 8)
	}
	barWidth = clamp(barWidth, 15, 60)

	total := m.profile.Difficulty.Total

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

	renderContents := func(bw int) (string, string, string) {
		return renderDifficulty(m, bw),
			renderLanguages(m, bw, total, maxItems),
			renderSkills(m, bw, total, maxItems)
	}
	diffContent, langsContent, skillsContent := renderContents(barWidth)
	for i := 0; i < 60; i++ {
		ws := []int{lipgloss.Width(diffContent), lipgloss.Width(langsContent), lipgloss.Width(skillsContent)}
		need := 0
		if useHorizontal {
			need = ws[0] + ws[1] + ws[2] + 16
		} else {
			need = 0
			for _, w := range ws {
				if w > need {
					need = w
				}
			}
			need += 4
		}
		if need <= width || barWidth <= 15 {
			break
		}
		barWidth--
		diffContent, langsContent, skillsContent = renderContents(barWidth)
	}

	diff := m.renderPanel("Difficulty", diffContent, PanelDifficulty)
	langs := m.renderPanel("Languages", langsContent, PanelLanguages)
	skills := m.renderPanel("Skills", skillsContent, PanelSkills)

	var panels string
	if useHorizontal {
		panels = lipgloss.JoinHorizontal(lipgloss.Top, diff, "  ", langs, "  ", skills)
	} else {
		panels = lipgloss.JoinVertical(lipgloss.Left, diff, "", langs, "", skills)
	}

	help := ""
	if m.showHelp {
		help = m.theme.Muted.Render("[r] refresh  [t] theme  [b] problems  [tab/←→] panel  [enter] detail  [h] hide  [q] quit")
	}

	frame := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		themeBadge,
		meta,
		"",
		panels,
		"",
		help,
	)

	return fit(frame, width)
}

func (m Model) renderDetailView() string {
	switch m.detailView {
	case DetailDifficulty:
		return m.renderDetailDifficulty()
	case DetailLanguages:
		return m.renderDetailLanguages()
	case DetailSkills:
		return m.renderDetailSkills()
	default:
		return ""
	}
}

func (m Model) renderProblemsView() string {
	width := m.width
	if width <= 0 {
		width = 80
	}

	title := m.theme.Title.Render("📋 Problem Browser")
	backHint := m.theme.Muted.Render("[q/esc] back  [/] filter  [j/k] move  [[]/]] page  [g/G] top/bottom  [enter] open  [r] refresh")

	head := []string{title}
	if fl := m.renderFilterLine(); fl != "" {
		head = append(head, fl)
	}

	if m.problemsLoading {
		content := "\n  " + m.spinner.View() + " Loading problems..."
		return lipgloss.JoinVertical(lipgloss.Left, append(head, "", content, "", backHint)...)
	}

	if m.problemsErr != nil {
		content := m.theme.Muted.Render(fmt.Sprintf("Error: %v", m.problemsErr))
		return lipgloss.JoinVertical(lipgloss.Left, append(head, "", content, "", backHint)...)
	}

	if !m.problemsLoaded {
		content := m.theme.Muted.Render("Press [r] to load problems")
		return lipgloss.JoinVertical(lipgloss.Left, append(head, "", content, "", backHint)...)
	}

	if len(m.problems) == 0 {
		content := m.theme.Muted.Render("No problems found")
		return lipgloss.JoinVertical(lipgloss.Left, append(head, "", content, "", backHint)...)
	}

	if m.visibleCount() == 0 {
		content := m.theme.Muted.Render(fmt.Sprintf("No problems match %q", m.query))
		return lipgloss.JoinVertical(lipgloss.Left, append(head, "", content, "", backHint)...)
	}

	visible := m.visibleCount()
	pageSize := m.pageSize()
	totalPages := (visible-1)/pageSize + 1

	cursor := m.cursor
	if cursor >= visible {
		cursor = visible - 1
	}
	if cursor < 0 {
		cursor = 0
	}
	page := cursor / pageSize
	start := page * pageSize
	end := start + pageSize
	if end > visible {
		end = visible
	}

	idW := len("ID")
	for _, p := range m.problems {
		if d := len(strconv.Itoa(p.ID)); d > idW {
			idW = d
		}
	}
	diffW := len("Difficulty")
	statusW := len("Status")
	gap := 2
	prefixW := 2

	titleW := width - prefixW - idW - diffW - statusW - gap*3
	if titleW < 10 {
		titleW = 10
	}

	headerLine := m.theme.Muted.Render(
		strings.Repeat(" ", prefixW) +
			padRight("ID", idW) + strings.Repeat(" ", gap) +
			padRight("Title", titleW) + strings.Repeat(" ", gap) +
			padRight("Difficulty", diffW) + strings.Repeat(" ", gap) +
			padRight("Status", statusW),
	)

	sepW := prefixW + idW + gap + titleW + gap + diffW + gap + statusW
	if sepW > width-2 {
		sepW = width - 2
	}
	if sepW < 10 {
		sepW = 10
	}
	sep := m.theme.Muted.Render(strings.Repeat("─", sepW))

	var lines []string
	lines = append(lines, headerLine, sep)

	for i := start; i < end; i++ {
		lines = append(lines, m.renderProblemLine(m.problemAt(i), i == cursor, idW, titleW, diffW, statusW, gap, width))
	}

	var progress string
	if m.query != "" {
		progress = m.theme.Muted.Render(fmt.Sprintf(
			"  Page %d/%d  •  matches %d–%d of %d",
			page+1, totalPages, start+1, end, visible,
		))
	} else {
		knownTotal := m.problemsTotal
		if knownTotal < len(m.problems) {
			knownTotal = len(m.problems)
		}
		progress = m.theme.Muted.Render(fmt.Sprintf(
			"  Page %d/%d  •  problems %d–%d of %d",
			page+1, totalPages, start+1, end, knownTotal,
		))
	}
	if m.problemsLoadingMore {
		progress += " " + m.theme.Muted.Render(fmt.Sprintf("%s loading…", m.spinner.View()))
	}

	return fit(lipgloss.JoinVertical(
		lipgloss.Left,
		append(head, "", lipgloss.JoinVertical(lipgloss.Left, lines...), "", progress, "", backHint)...,
	), width)
}

func (m Model) renderFilterLine() string {
	if m.searchMode {
		return m.theme.Selected.Render("/ ") + m.theme.Title.Render(m.query) +
			m.theme.Muted.Render(fmt.Sprintf("▌  %d matches", m.visibleCount()))
	}
	if m.query == "" {
		return ""
	}
	return m.theme.Muted.Render("filter: ") + m.theme.Accent.Render(m.query) +
		m.theme.Muted.Render(fmt.Sprintf("  %d matches  [/] edit [esc] clear", m.visibleCount()))
}

func (m Model) renderProblemLine(p domain.Problem, selected bool, idW, titleW, diffW, statusW, gap int, width int) string {
	prefix := "  "
	if selected {
		prefix = m.theme.Selected.Render("▶ ")
	}

	idStr := padRight(strconv.Itoa(p.ID), idW)
	titleStr := padRight(trunc(p.Title, titleW), titleW)
	if selected {
		titleStr = m.theme.Selected.Render(titleStr)
	}

	var diffStyle lipgloss.Style
	switch p.Difficulty {
	case "Easy":
		diffStyle = m.theme.Easy
	case "Medium":
		diffStyle = m.theme.Medium
	case "Hard":
		diffStyle = m.theme.Hard
	default:
		diffStyle = m.theme.Muted
	}
	diffStr := diffStyle.Render(padRight(p.Difficulty, diffW))

	status := "❌"
	if p.Status == "Solved" {
		status = "✅"
	}
	statusStr := padRight(status, statusW)

	line := prefix + idStr + strings.Repeat(" ", gap) +
		titleStr + strings.Repeat(" ", gap) +
		diffStr + strings.Repeat(" ", gap) +
		statusStr

	return fit(line, width)
}

func padRight(s string, w int) string {
	d := lipgloss.Width(s)
	if d >= w {
		return s
	}
	return s + strings.Repeat(" ", w-d)
}

func trunc(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	return truncate.String(s, uint(w-1)) + "…"
}

func (m Model) renderDetailDifficulty() string {
	total := float64(m.profile.Difficulty.Total)
	if total == 0 {
		return m.centered("No data")
	}

	leetcodeTotals := m.profile.LeetCodeTotals
	if leetcodeTotals.Easy == 0 && leetcodeTotals.Medium == 0 && leetcodeTotals.Hard == 0 {
		leetcodeTotals.Easy = 800
		leetcodeTotals.Medium = 1600
		leetcodeTotals.Hard = 700
	}
	leetcodeTotal := leetcodeTotals.Easy + leetcodeTotals.Medium + leetcodeTotals.Hard

	easyPct := float64(m.profile.Difficulty.Easy) / float64(leetcodeTotals.Easy) * 100
	mediumPct := float64(m.profile.Difficulty.Medium) / float64(leetcodeTotals.Medium) * 100
	hardPct := float64(m.profile.Difficulty.Hard) / float64(leetcodeTotals.Hard) * 100

	solved := m.profile.Difficulty.Easy + m.profile.Difficulty.Medium + m.profile.Difficulty.Hard
	totalPct := float64(solved) / float64(leetcodeTotal) * 100

	barWidth := clamp(m.width-30, 15, 60)

	title := m.theme.Title.Render("📊 Difficulty Breakdown")
	backHint := m.theme.Muted.Render("[enter/esc] back  [esc] close")

	rows := []string{
		renderDetailBar("Easy", m.profile.Difficulty.Easy, leetcodeTotals.Easy, easyPct, barWidth, m.theme.Easy, m.theme.BarTrack),
		renderDetailBar("Medium", m.profile.Difficulty.Medium, leetcodeTotals.Medium, mediumPct, barWidth, m.theme.Medium, m.theme.BarTrack),
		renderDetailBar("Hard", m.profile.Difficulty.Hard, leetcodeTotals.Hard, hardPct, barWidth, m.theme.Hard, m.theme.BarTrack),
		"",
		renderDetailBar("Total", solved, leetcodeTotal, totalPct, barWidth, m.theme.Title, m.theme.BarTrack),
	}

	stats := fmt.Sprintf(
		"Easy: %d/%d (%.1f%%)   Medium: %d/%d (%.1f%%)   Hard: %d/%d (%.1f%%)",
		m.profile.Difficulty.Easy, leetcodeTotals.Easy, easyPct,
		m.profile.Difficulty.Medium, leetcodeTotals.Medium, mediumPct,
		m.profile.Difficulty.Hard, leetcodeTotals.Hard, hardPct,
	)
	statsLine := m.theme.Muted.Render(stats)

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	content = lipgloss.JoinVertical(lipgloss.Left, content, "", statsLine)

	panelWidth := clamp(m.width-4, 40, 200)
	panel := m.theme.panelStyle(panelWidth, true).
		Render(fit(content, panelWidth-6))

	return lipgloss.JoinVertical(lipgloss.Center, "", title, "", panel, "", backHint)
}

func (m Model) renderDetailLanguages() string {
	if len(m.profile.Languages) == 0 {
		return m.centered("No data")
	}

	total := m.profile.Difficulty.Total
	if total == 0 {
		return m.centered("No data")
	}

	title := m.theme.Title.Render("💻 Languages Breakdown")
	backHint := m.theme.Muted.Render("[enter/esc] back  [esc] close")

	barWidth := clamp(m.width-40, 15, 70)
	maxItems := clamp(m.height-14, 3, 20)

	var names []string
	for i, l := range m.profile.Languages {
		if i >= maxItems {
			break
		}
		names = append(names, l.Name)
	}
	labelW := barLabelWidth(names, 20, 24, barWidth)

	var rows []string
	for i, l := range m.profile.Languages {
		if i >= maxItems {
			break
		}
		ratio := float64(l.Count) / float64(total) * 100
		rows = append(rows, renderDetailLangBar(l.Name, l.Count, ratio, barWidth, labelW, m.theme.Accent, m.theme.BarTrack))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	top3 := len(m.profile.Languages)
	if top3 > 3 {
		top3 = 3
	}
	topLangs := make([]string, top3)
	for i := 0; i < top3; i++ {
		topLangs[i] = fmt.Sprintf("%s(%d)", m.profile.Languages[i].Name, m.profile.Languages[i].Count)
	}
	summary := m.theme.Muted.Render(fmt.Sprintf("Top: %s   •   Total languages: %d", strings.Join(topLangs, ", "), len(m.profile.Languages)))

	content = lipgloss.JoinVertical(lipgloss.Left, content, "", summary)

	panelWidth := clamp(m.width-4, 40, 200)
	panel := m.theme.panelStyle(panelWidth, true).
		Render(fit(content, panelWidth-6))

	return lipgloss.JoinVertical(lipgloss.Center, "", title, "", panel, "", backHint)
}

func (m Model) renderDetailSkills() string {
	if len(m.profile.Skills) == 0 {
		return m.centered("No data")
	}

	total := m.profile.Difficulty.Total
	if total == 0 {
		return m.centered("No data")
	}

	title := m.theme.Title.Render("🏷️  Skills / Tags Breakdown")
	backHint := m.theme.Muted.Render("[enter/esc] back  [esc] close")

	barWidth := clamp(m.width-40, 15, 70)
	maxItems := clamp(m.height-14, 3, 25)

	var names []string
	for i, s := range m.profile.Skills {
		if i >= maxItems {
			break
		}
		names = append(names, s.Name)
	}
	labelW := barLabelWidth(names, 20, 24, barWidth)

	var rows []string
	for i, s := range m.profile.Skills {
		if i >= maxItems {
			break
		}
		ratio := float64(s.Count) / float64(total) * 100
		rows = append(rows, renderDetailLangBar(s.Name, s.Count, ratio, barWidth, labelW, m.theme.Accent, m.theme.BarTrack))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	dsTags := map[string]bool{"Array": true, "String": true, "Hash Table": true, "Linked List": true, "Stack": true, "Queue": true, "Tree": true, "Binary Tree": true, "Heap": true, "Graph": true, "Trie": true}
	algoTags := map[string]bool{"Dynamic Programming": true, "Binary Search": true, "Two Pointers": true, "Sliding Window": true, "Greedy": true, "Backtracking": true, "Divide and Conquer": true, "Sorting": true, "Search": true, "Recursion": true, "Bit Manipulation": true, "Math": true, "Geometry": true}

	dsCount := 0
	algoCount := 0
	otherCount := 0
	for _, s := range m.profile.Skills {
		if dsTags[s.Name] {
			dsCount += s.Count
		} else if algoTags[s.Name] {
			algoCount += s.Count
		} else {
			otherCount += s.Count
		}
	}

	var catParts []string
	if dsCount > 0 {
		catParts = append(catParts, fmt.Sprintf("Data Structures: %d (%.1f%%)", dsCount, float64(dsCount)/float64(total)*100))
	}
	if algoCount > 0 {
		catParts = append(catParts, fmt.Sprintf("Algorithms: %d (%.1f%%)", algoCount, float64(algoCount)/float64(total)*100))
	}
	if otherCount > 0 {
		catParts = append(catParts, fmt.Sprintf("Other: %d (%.1f%%)", otherCount, float64(otherCount)/float64(total)*100))
	}
	summary := m.theme.Muted.Render(fmt.Sprintf("%s   •   Total tags: %d", strings.Join(catParts, "   "), len(m.profile.Skills)))

	content = lipgloss.JoinVertical(lipgloss.Left, content, "", summary)

	panelWidth := clamp(m.width-4, 40, 200)
	panel := m.theme.panelStyle(panelWidth, true).
		Render(fit(content, panelWidth-6))

	return lipgloss.JoinVertical(lipgloss.Center, "", title, "", panel, "", backHint)
}

func (m Model) centered(s string) string {
	w := m.width
	h := m.height
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, m.theme.Muted.Render(s))
}

func (m Model) renderPanel(title, content string, panel Panel) string {
	style := m.theme.panelStyle(0, m.activePanel == panel)
	titleStyle := m.theme.Muted
	if m.activePanel == panel {
		titleStyle = m.theme.Selected
	}
	return style.Render(titleStyle.Render(title) + "\n" + content)
}

func barLabelWidth(names []string, minW, maxW, barWidth int) int {
	w := minW
	if barWidth < 30 && w > 10 {
		w = 10
	}
	for _, n := range names {
		if d := lipgloss.Width(n); d > w {
			w = d
		}
	}
	if w > maxW {
		w = maxW
	}
	return w
}

func filledCells(ratio float64, width int) int {
	filled := int(ratio * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return filled
}

func renderDifficulty(m Model, width int) string {
	total := float64(m.profile.Difficulty.Total)
	if total == 0 {
		return m.theme.Muted.Render("No data")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderBar("Easy", m.profile.Difficulty.Easy, total, width, m.theme.Easy, m.theme.BarTrack),
		renderBar("Medium", m.profile.Difficulty.Medium, total, width, m.theme.Medium, m.theme.BarTrack),
		renderBar("Hard", m.profile.Difficulty.Hard, total, width, m.theme.Hard, m.theme.BarTrack),
	)
}

func renderLanguages(m Model, width, total, maxItems int) string {
	if len(m.profile.Languages) == 0 || total == 0 {
		return m.theme.Muted.Render("No data")
	}

	var names []string
	for i, l := range m.profile.Languages {
		if i >= maxItems {
			break
		}
		names = append(names, l.Name)
	}
	labelW := barLabelWidth(names, 14, 20, width)

	var lines []string
	for i, l := range m.profile.Languages {
		if i >= maxItems {
			break
		}
		ratio := float64(l.Count) / float64(total)
		lines = append(lines, renderCustomBar(l.Name, l.Count, ratio, width, labelW, m.theme.Accent, m.theme.BarTrack))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderSkills(m Model, width, total, maxItems int) string {
	if len(m.profile.Skills) == 0 || total == 0 {
		return m.theme.Muted.Render("No data")
	}

	var names []string
	for i, s := range m.profile.Skills {
		if i >= maxItems {
			break
		}
		names = append(names, s.Name)
	}
	labelW := barLabelWidth(names, 14, 20, width)

	var lines []string
	for i, s := range m.profile.Skills {
		if i >= maxItems {
			break
		}
		ratio := float64(s.Count) / float64(total)
		lines = append(lines, renderCustomBar(s.Name, s.Count, ratio, width, labelW, m.theme.Accent, m.theme.BarTrack))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderBar(label string, value int, total float64, width int, style, track lipgloss.Style) string {
	ratio := float64(value) / total
	filled := filledCells(ratio, width)

	bar := style.Render(strings.Repeat("█", filled)) +
		track.Render(strings.Repeat(" ", width-filled))

	pct := fmt.Sprintf("%.0f%%", ratio*100)
	labelWidth := 8
	if width < 25 {
		labelWidth = 5
	}
	labelStr := padRight(trunc(label, labelWidth), labelWidth)
	return fmt.Sprintf("%s %s %5s %d", labelStr, bar, pct, value)
}

func renderCustomBar(label string, count int, ratio float64, width, labelWidth int, style, track lipgloss.Style) string {
	filled := filledCells(ratio, width)
	bar := style.Render(strings.Repeat("█", filled)) +
		track.Render(strings.Repeat(" ", width-filled))
	pct := fmt.Sprintf("%.0f%%", ratio*100)
	labelStr := padRight(trunc(label, labelWidth), labelWidth)
	return fmt.Sprintf("%s %s %4s %d", labelStr, bar, pct, count)
}

func renderDetailBar(label string, solved, total int, pct float64, width int, style, track lipgloss.Style) string {
	filled := filledCells(pct/100, width)
	bar := style.Render(strings.Repeat("█", filled)) +
		track.Render(strings.Repeat(" ", width-filled))
	pctStr := fmt.Sprintf("%.1f%%", pct)
	labelStr := padRight(trunc(label, 8), 8)
	return fmt.Sprintf("%s %s %7s %4d/%d", labelStr, bar, pctStr, solved, total)
}

func renderDetailLangBar(label string, count int, pct float64, width, labelWidth int, style, track lipgloss.Style) string {
	filled := filledCells(pct/100, width)
	bar := style.Render(strings.Repeat("█", filled)) +
		track.Render(strings.Repeat(" ", width-filled))
	pctStr := fmt.Sprintf("%.1f%%", pct)
	labelStr := padRight(trunc(label, labelWidth), labelWidth)
	return fmt.Sprintf("%s %s %7s %5d", labelStr, bar, pctStr, count)
}
