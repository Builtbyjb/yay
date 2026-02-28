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
		case colMod:
			colName = "Modifier"
		case colKey:
			if m.recordingHotkey {
				colName = "Key (recording)"
			} else {
				colName = "Key"
			}
		case colMode:
			colName = "Mode"
		case colEnabled:
			colName = "Enabled"
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
