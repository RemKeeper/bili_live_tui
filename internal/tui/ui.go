package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// keyString keeps modified keys (Alt+N, Ctrl+S) intact. Bubble Tea v2's
// String() drops the modifier when the event also carries text.
func keyString(msg tea.KeyPressMsg) string {
	if msg.Mod != 0 {
		return msg.Keystroke()
	}
	return msg.String()
}

func screen(content string) tea.View {
	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "bililive"
	return v
}

func fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return ansi.Truncate(s, width, "")
}

func fitTail(s string, width int, tail string) string {
	if width <= 0 {
		return ""
	}
	return ansi.Truncate(s, width, tail)
}

func fitBlock(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		lines[i] = fit(line, width)
	}
	return strings.Join(lines, "\n")
}

// repeatCell fills a row without assuming a glyph is one cell wide.
func repeatCell(ch string, width int) string {
	if width <= 0 || ch == "" {
		return ""
	}
	cell := lipgloss.Width(ch)
	if cell <= 0 {
		return ""
	}
	n := width / cell
	return strings.Repeat(ch, n) + strings.Repeat(" ", width-n*cell)
}
