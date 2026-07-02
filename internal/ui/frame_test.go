package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestFrameFitsWidth(t *testing.T) {
	for _, width := range []int{145, 160, 200, 250} {
		m := testModel(width, 45)
		frame := m.View()
		for i, l := range strings.Split(frame, "\n") {
			if w := lipgloss.Width(l); w > width {
				t.Errorf("width=%d line %d overflows: %d > %d: %q", width, i, w, width, l)
			}
		}
	}

	for _, width := range []int{80, 100, 120, 144} {
		m := testModel(width, 40)
		frame := m.View()
		for i, l := range strings.Split(frame, "\n") {
			if w := lipgloss.Width(l); w > width {
				t.Errorf("width=%d line %d overflows: %d > %d: %q", width, i, w, width, l)
			}
		}
	}
}
