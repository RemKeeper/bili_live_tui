package tui

import tea "charm.land/bubbletea/v2"

func press(name string) tea.KeyPressMsg {
	switch name {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "f2":
		return tea.KeyPressMsg{Code: tea.KeyF2}
	case "f6":
		return tea.KeyPressMsg{Code: tea.KeyF6}
	case "ctrl+s":
		return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
	case "alt+n":
		return tea.KeyPressMsg{Code: 'n', Mod: tea.ModAlt}
	case "alt+m":
		return tea.KeyPressMsg{Code: 'm', Mod: tea.ModAlt}
	default:
		return tea.KeyPressMsg{Code: tea.KeyExtended, Text: name}
	}
}
