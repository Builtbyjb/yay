package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m model) StatusLineView() string {
	var content string

	switch m.state {
	case stateRowFocus:
		colName := "none"
		switch m.activeCol {
		case colKey:
			if m.recordingHotkey {
				colName = "hotkey (recording)"
			} else {
				colName = "hotkey"
			}
		case colMode:
			colName = "mode"
		case colEnabled:
			colName = "enabled"
		}
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			StatusStyle.Render(fmt.Sprintf("EDITING ROW  |  column: %s", colName)),
		)
	case stateFilter:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			StatusStyle.Render("SEARCH MODE"),
		)
	default:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			StatusStyle.Render("BROWSE"),
		)
	}

	return content
}
