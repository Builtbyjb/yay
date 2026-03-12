package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (m model) View() string {
	contents := []string{}
	contents = append(contents, m.HeaderView())
	contents = append(contents, m.SearchView())
	contents = append(contents, m.TableView())
	contents = append(contents, m.StatusLineView())
	contents = append(contents, m.HelpView())

	return lipgloss.JoinVertical(lipgloss.Left, contents...)
}

func (m model) HeaderView() string {
	var ascii_logo = `
██    ██  █████  ██    ██
 ██  ██  ██   ██  ██  ██
  ████   ███████   ████
   ██    ██   ██    ██
   ██    ██   ██    ██` + "	v" + m.version

	logo := lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Split(LogoStyle.Render(ascii_logo), "\n")...,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		logo,
		"\n",
	)
}

func (m model) SearchView() string {

	contents := []string{}

	contents = append(contents, lipgloss.JoinVertical(
		lipgloss.Left,
		FilterLabelStyle.Render("Search: "),
	))

	if m.state == stateFilter {
		contents = append(contents, m.searchInput.View())
	} else {
		// Show filter text but not focused
		filterText := m.searchInput.Value()
		if filterText == "" {
			filterText = DimStyle.Render("(press / to Search)")
		} else {
			filterText = NormalRowStyle.Render(filterText)
		}

		contents = append(contents, filterText)
	}

	contents = append(contents, "\n")

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		contents...,
	)
}

func (m model) TableView() string {
	contents := []string{}

	// Set table height
	maxRows := max(m.height-20, 3)

	// Calculate scroll window
	startIdx := 0
	if len(m.searchedIndices) > maxRows {
		if m.cursor >= maxRows {
			startIdx = m.cursor - maxRows + 1
		}
		if startIdx+maxRows > len(m.searchedIndices) {
			startIdx = len(m.searchedIndices) - maxRows
		}
	}

	endIdx := min(startIdx+maxRows, len(m.searchedIndices))

	if len(m.searchedIndices) == 0 {
		contents = append(contents, lipgloss.JoinVertical(
			lipgloss.Left,
			DimStyle.Render("No matching applications."),
			"\n",
		))
	} else {
		table := table.New().
			Border(lipgloss.NormalBorder()).
			Width(m.width).
			StyleFunc(func(row, col int) lipgloss.Style {
				isHeader := row == table.HeaderRow
				isCursorRow := row >= 0 && (startIdx+row) == m.cursor
				isFocused := isCursorRow && m.state == stateRowFocus

				base := lipgloss.NewStyle().
					Padding(0, 1).
					Foreground(lipgloss.Color(PRIMARY_COLOR))

				if isHeader {
					return base.Bold(true).Foreground(lipgloss.Color(PRIMARY_COLOR))
				}

				// Default row style
				style := base
				if isCursorRow {
					style = CursorRowStyle // your existing cursor style
				} else {
					style = NormalRowStyle // your normal row style
				}

				// When row-focused → per-cell highlighting
				if isFocused {
					style = FocusedRowStyle // base for whole row

					// Active (editing/recording) column gets stronger highlight
					if (col == 1 && m.activeCol == colKey) ||
						(col == 2 && m.activeCol == colMode) ||
						(col == 3 && m.activeCol == colEnabled) {
						style = ActiveCellStyle
					}
				}

				style = style.Align(lipgloss.Left).Padding(0, 1)

				return style
			})

		table.Headers("Application", "HotKey", "Mode", "Enabled")

		// Add only the visible rows
		for i := startIdx; i < endIdx; i++ {
			idx := m.searchedIndices[i]
			s := m.settings[idx]

			isCursor := i == m.cursor
			isFocused := isCursor && m.state == stateRowFocus

			prefix := "  "
			if isCursor {
				prefix = "> "
			}

			name := truncate(s.Name, colWidthName-3) // -3 for safety + prefix
			name = prefix + name

			hotkey := displayKey(s.HotKey.String)

			// Special case: recording hotkey
			if isFocused && m.activeCol == colKey && m.recordingHotkey {
				hotkeyDisplay := "recording..."
				hotkey = hotkeyDisplay
			}

			mode := s.Mode
			enabled := formatBool(s.Enabled)

			table.Row(name, hotkey, mode, enabled)
		}

		contents = append(contents, lipgloss.JoinVertical(
			lipgloss.Left,
			table.Render(),
		))

		// Scroll indicator
		if len(m.searchedIndices) >= endIdx {
			contents = append(contents, lipgloss.JoinVertical(
				lipgloss.Left,
				DimStyle.Render(fmt.Sprintf("showing %d-%d of %d", startIdx+1, endIdx, len(m.searchedIndices))),
			))
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		contents...,
	)
}

func (m model) StatusLineView() string {
	var content string

	switch m.state {
	case stateRowFocus:
		colName := "none"
		switch m.activeCol {
		case colKey:
			if m.recordingHotkey {
				colName = "HotKey (recording)"
			} else {
				colName = "HotKey"
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

func (m model) HelpView() string {
	var content string

	switch m.state {
	case stateBrowse:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			HelpStyle.Render("↑/↓/j/k: Navigate | enter: Edit Row | /: Search | esc/ctrl+c: Quit"),
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
