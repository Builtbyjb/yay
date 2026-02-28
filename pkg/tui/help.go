package tui

import "github.com/charmbracelet/lipgloss"

func (m model) HelpView() string {
	var content string

	switch m.state {
	case stateBrowse:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			HelpStyle.Render("↑/↓/j/k: Navigate | enter: Edit Row | /: Search | ctrl+c: Quit"),
		)
	case stateFilter:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			HelpStyle.Render("↑/↓: Navigate | enter: Edit Row | esc: Stop Searching | ctrl+c: Quit"),
		)
	case stateRowFocus:
		switch m.activeCol {
		case colKey:
			if m.recordingHotkey {
				content = lipgloss.JoinVertical(
					lipgloss.Left,
					HelpStyle.Render("Press any key to set key | backspace: Clear | esc: Cancel"),
				)
			} else {
				content = lipgloss.JoinVertical(
					lipgloss.Left,
					HelpStyle.Render("tab: Next Column | enter: Record Hotkey | esc: Un-focus"),
				)
			}
		case colMode:
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				HelpStyle.Render("space/enter/←/→: Cycle Mode | tab: Next Column | esc: Un-focus"),
			)
		case colEnabled:
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				HelpStyle.Render("space/enter: Toggle | tab: Next Column | esc: Un-focus"),
			)
		default:
			content = lipgloss.JoinVertical(
				lipgloss.Left,
				HelpStyle.Render("tab: select column | esc/ctrl+q: un-focus"),
			)
		}
	}

	return content
}
