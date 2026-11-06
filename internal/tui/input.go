package tui

import (
	"charm.land/lipgloss/v2"
)

func (m *model) layoutInputs() {
	width := max(1, m.width-2*focusMarginWidth)
	m.textInput.Prompt = m.inputPrompt(width)
	m.textInput.SetWidth(max(1, width-lipgloss.Width(m.textInput.Prompt)-1))
	m.textInput.SetCursor(m.textInput.Position())
	if m.settings != nil {
		for i := range m.settings.fields {
			label := fitTail(settingLabels[i]+": ", max(0, m.viewport.Width()/2-2), "…")
			field := &m.settings.fields[i]
			field.SetWidth(max(1, m.viewport.Width()-2-lipgloss.Width(label)-1))
			field.SetCursor(field.Position())
		}
	}
}

func (m model) inputPrompt(width int) string {
	limit := max(1, width/2)
	audio, status := m.audioLabel(), singleLine(m.sendStatus)
	text := audio
	if status != "" && audio != "" {
		// Keep both one-second notices visible; a long missing-player message
		// must not swallow a simultaneous save/send result. Normal-width mute
		// notices still fit in full, and the steady icon uses only one cell.
		audio = fitTail(singleLine(audio), max(1, limit/2), "…")
		budget := max(1, limit-lipgloss.Width(audio)-3)
		text = fitTail(status, budget, "…") + " · " + audio
	} else if status != "" {
		text = status
	}
	if text == "" {
		return "> "
	}
	return fitTail(singleLine(text), limit, "…") + " > "
}

// Status lives in the existing input row, including when a failed draft is non-empty.
func (m model) inputLine() string {
	ti := m.textInput
	width := max(1, m.width-2*focusMarginWidth)
	if m.settings != nil {
		hint := "Tab/↑↓ 选择 · 空格 切换 · Ctrl+S 保存 · Esc 取消"
		if m.settings.errorText != "" {
			hint = m.settings.errorText + " · Ctrl+S 重试 · Esc 取消"
		}
		if audio := m.audioLabel(); audio != "" {
			hint = audio + " · " + hint
		}
		return fitTail(hint, width, "…")
	}
	ti.Prompt = m.inputPrompt(width)
	ti.SetWidth(max(1, width-lipgloss.Width(ti.Prompt)-1))
	ti.SetCursor(ti.Position())
	return fit(ti.View(), width)
}
