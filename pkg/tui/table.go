package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

func (m model) TableView() string {
	contents := []string{}

	// Set table height
	maxRows := m.height - 20
	if maxRows < 3 {
		maxRows = 3
	}

	// Calculate scroll window
	startIdx := 0
	if len(m.filteredIndices) > maxRows {
		if m.cursor >= maxRows {
			startIdx = m.cursor - maxRows + 1
		}
		if startIdx+maxRows > len(m.filteredIndices) {
			startIdx = len(m.filteredIndices) - maxRows
		}
	}

	endIdx := startIdx + maxRows
	if endIdx > len(m.filteredIndices) {
		endIdx = len(m.filteredIndices)
	}

	if len(m.filteredIndices) == 0 {
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
					if (col == 1 && m.activeCol == colMod) ||
						(col == 2 && m.activeCol == colKey) ||
						(col == 3 && m.activeCol == colMode) ||
						(col == 4 && m.activeCol == colEnabled) {
						style = ActiveCellStyle
					}
				}

				style = style.Align(lipgloss.Left).Padding(0, 1)

				return style
			})

		table.Headers("Application", "Modifier", "Key", "Mode", "Enabled")

		// Add only the visible rows
		for i := startIdx; i < endIdx; i++ {
			idx := m.filteredIndices[i]
			s := m.settings[idx]

			isCursor := i == m.cursor
			isFocused := isCursor && m.state == stateRowFocus

			prefix := "  "
			if isCursor {
				prefix = "> "
			}

			name := truncate(s.Name, colWidthName-3) // -3 for safety + prefix
			name = prefix + name

			hotkey := displayHotkey(s.HotKey)

			// Special case: recording hotkey
			if isFocused && m.activeCol == colKey && m.recordingHotkey {
				hotkeyDisplay := "recording..."
				if m.pendingMods != "" {
					hotkeyDisplay = m.pendingMods + "+..."
				}
				hotkey = hotkeyDisplay
			}

			mode := s.Mode
			enabled := formatBool(s.Enabled)

			mod := "---"

			table.Row(
				name,
				mod,
				hotkey,
				mode,
				enabled,
			)
		}

		contents = append(contents, lipgloss.JoinVertical(
			lipgloss.Left,
			table.Render(),
		))

		// Scroll indicator
		if len(m.filteredIndices) >= endIdx {
			contents = append(contents, lipgloss.JoinVertical(
				lipgloss.Left,
				DimStyle.Render(fmt.Sprintf("showing %d-%d of %d", startIdx+1, endIdx, len(m.filteredIndices))),
			))
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		contents...,
	)
}
