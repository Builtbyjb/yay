package tui

import (
	"database/sql"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

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

func (m model) HandleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case EXIT_KEY, CANCEL_KEY:
		return m, tea.Quit

	case "up", "k":
		m.moveCursor(-1)
		return m, nil

	case "down", "j":
		m.moveCursor(1)
		return m, nil

	case "home", "g":
		m.cursor = 0
		return m, nil

	case "end", "G":
		if len(m.searchedIndices) > 0 {
			m.cursor = len(m.searchedIndices) - 1
		}
		return m, nil

	case "enter":
		if len(m.searchedIndices) > 0 {
			m.state = stateRowFocus
			m.activeCol = colKey
			m.recordingHotkey = false
		}
		return m, nil

	case SEARCH_KEY:
		m.state = stateFilter
		m.searchInput.Focus()
		return m, nil
	}

	return m, nil
}

func (m model) handleRowFocusKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If recording a hotkey, capture the key
	if m.recordingHotkey {
		return m.handleHotkeyRecording(msg)
	}

	switch msg.String() {
	case EXIT_KEY:
		return m, tea.Quit

	case CANCEL_KEY:
		m.state = stateBrowse
		m.activeCol = colNone
		m.recordingHotkey = false
		return m, nil

	case SWITCH_COLUMN_KEY:
		m.cycleColumn()
		return m, nil

	case "up", "k":
		m.moveCursor(-1)
		return m, nil

	case "down", "j":
		m.moveCursor(1)
		return m, nil
	}

	// Column-specific actions
	switch m.activeCol {
	case colKey:
		switch msg.String() {
		case "enter", " ":
			m.recordingHotkey = true
			return m, nil
		case "delete", "backspace":
			if len(m.searchedIndices) == 0 || m.cursor >= len(m.searchedIndices) {
				return m, nil
			}
			idx := m.searchedIndices[m.cursor]
			settingId := m.settings[idx].Id
			err := m.db.ClearHotkey(settingId)
			m.settings[idx].HotKey = sql.NullString{String: "", Valid: false}

			if err != nil {
				m.errors = append(m.errors, err.Error())
			}
			return m, nil
		}

	case colMode:
		switch msg.String() {
		case "enter", " ":
			m.cycleMode()
			return m, nil
		}

	case colEnabled:
		switch msg.String() {
		case "enter", " ":
			m.toggleEnabled()
			return m, nil
		}
	}

	return m, nil
}

func (m *model) moveCursor(delta int) {
	if len(m.searchedIndices) == 0 {
		m.cursor = 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.searchedIndices) {
		m.cursor = len(m.searchedIndices) - 1
	}
}

func (m *model) cycleColumn() {
	switch m.activeCol {
	case colNone, colEnabled:
		m.activeCol = colKey
	case colKey:
		m.activeCol = colMode
	case colMode:
		m.activeCol = colEnabled
	}
	m.recordingHotkey = false
}

func (m *model) cycleMode() {
	if len(m.searchedIndices) == 0 || m.cursor >= len(m.searchedIndices) {
		return
	}
	idx := m.searchedIndices[m.cursor]
	prev := m.settings[idx].Mode
	currentIdx := indexOf(AvailableModes, prev)
	nextIdx := (currentIdx + 1) % len(AvailableModes)
	m.settings[idx].Mode = AvailableModes[nextIdx]
	if err := m.db.UpdateMode(m.settings[idx].Id, m.settings[idx].Mode); err != nil {
		m.errors = append(m.errors, err.Error())
	}
}

func (m *model) toggleEnabled() {
	if len(m.searchedIndices) == 0 || m.cursor >= len(m.searchedIndices) {
		return
	}
	idx := m.searchedIndices[m.cursor]
	prev := m.settings[idx].Enabled
	m.settings[idx].Enabled = !prev
	if err := m.db.UpdateEnabled(m.settings[idx].Id, m.settings[idx].Enabled); err != nil {
		m.errors = append(m.errors, err.Error())
	}
}

func (m model) handleHotkeyRecording(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == CANCEL_KEY {
		m.recordingHotkey = false
		return m, nil
	}

	if key == EXIT_KEY {
		return m, tea.Quit
	}

	return m, nil
}
