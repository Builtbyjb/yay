package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
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

			hotkey := displayKey(s.Key)

			// Special case: recording hotkey
			if isFocused && m.activeCol == colKey && m.recordingHotkey {
				hotkeyDisplay := "recording..."
				hotkey = hotkeyDisplay
			}

			mode := s.Mode
			enabled := formatBool(s.Enabled)
			mod := displayMod(s.Mod)

			table.Row(name, mod, hotkey, mode, enabled)
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
		if len(m.filteredIndices) > 0 {
			m.cursor = len(m.filteredIndices) - 1
		}
		return m, nil

	case "enter":
		if len(m.filteredIndices) > 0 {
			m.state = stateRowFocus
			m.activeCol = colMod
			m.recordingHotkey = false
		}
		return m, nil

	case SEARCH_KEY:
		m.state = stateFilter
		m.filterInput.Focus()
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
	case colMod:
		switch msg.String() {
		case "enter", " ", "right", "l":
			m.cycleModifierForward()
			return m, nil
		case "left", "h":
			m.cycleModifierBackward()
			return m, nil
		}
	case colKey:
		if msg.String() == "enter" || msg.String() == " " {
			m.recordingHotkey = true
			return m, nil
		}

	case colMode:
		switch msg.String() {
		case "enter", " ", "right", "l":
			m.cycleModeForward()
			return m, nil
		case "left", "h":
			m.cycleModeBackward()
			return m, nil
		}

	case colEnabled:
		if msg.String() == "enter" || msg.String() == " " {
			m.toggleEnabled()
			return m, nil
		}
	}

	return m, nil
}

func (m *model) cycleModifierForward() {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return
	}
	idx := m.filteredIndices[m.cursor]
	old := m.settings[idx].Mod
	current := indexOf(AvailableModifiersMacos, old)
	next := (current + 1) % len(AvailableModifiersMacos)
	m.settings[idx].Mod = AvailableModifiersMacos[next]
	m.changes = append(m.changes, changeEntry{
		Name:   m.settings[idx].Name,
		Field:  "Modifier",
		OldVal: displayMod(old),
		NewVal: m.settings[idx].Mod,
	})
}

func (m *model) cycleModifierBackward() {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return
	}
	idx := m.filteredIndices[m.cursor]
	old := m.settings[idx].Mode
	current := indexOf(AvailableModes, old)
	next := current - 1
	if next < 0 {
		next = len(AvailableModes) - 1
	}
	m.settings[idx].Mode = AvailableModes[next]
	m.changes = append(m.changes, changeEntry{
		Name:   m.settings[idx].Name,
		Field:  "mode",
		OldVal: old,
		NewVal: m.settings[idx].Mode,
	})
}

func (m *model) cycleModeForward() {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return
	}
	idx := m.filteredIndices[m.cursor]
	old := m.settings[idx].Mode
	current := indexOf(AvailableModes, old)
	next := (current + 1) % len(AvailableModes)
	m.settings[idx].Mode = AvailableModes[next]
	m.changes = append(m.changes, changeEntry{
		Name:   m.settings[idx].Name,
		Field:  "mode",
		OldVal: old,
		NewVal: m.settings[idx].Mode,
	})
}

func (m *model) cycleModeBackward() {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return
	}
	idx := m.filteredIndices[m.cursor]
	old := m.settings[idx].Mode
	current := indexOf(AvailableModes, old)
	next := current - 1
	if next < 0 {
		next = len(AvailableModes) - 1
	}
	m.settings[idx].Mode = AvailableModes[next]
	m.changes = append(m.changes, changeEntry{
		Name:   m.settings[idx].Name,
		Field:  "mode",
		OldVal: old,
		NewVal: m.settings[idx].Mode,
	})
}

func (m *model) cycleColumn() {
	switch m.activeCol {
	case colNone, colEnabled:
		m.activeCol = colMod
	case colMod:
		m.activeCol = colKey
	case colKey:
		m.activeCol = colMode
	case colMode:
		m.activeCol = colEnabled
	}
	m.recordingHotkey = false
}

func (m *model) moveCursor(delta int) {
	if len(m.filteredIndices) == 0 {
		m.cursor = 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.filteredIndices) {
		m.cursor = len(m.filteredIndices) - 1
	}
}

func (m *model) toggleEnabled() {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return
	}
	idx := m.filteredIndices[m.cursor]
	old := m.settings[idx].Enabled
	m.settings[idx].Enabled = !old
	m.changes = append(m.changes, changeEntry{
		Name:   m.settings[idx].Name,
		Field:  "enabled",
		OldVal: formatBool(old),
		NewVal: formatBool(m.settings[idx].Enabled),
	})
}

func (m model) handleHotkeyRecording(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Esc cancels recording
	if key == "esc" {
		m.recordingHotkey = false
		return m, nil
	}

	// Backspace clears the hotkey
	if key == "backspace" {
		if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
			idx := m.filteredIndices[m.cursor]
			old := m.settings[idx].Key
			m.settings[idx].Key = ""
			m.recordingHotkey = false
			if old != "" {
				m.changes = append(m.changes, changeEntry{
					Name:   m.settings[idx].Name,
					Field:  "hotkey",
					OldVal: old,
					NewVal: "(cleared)",
				})
			}
		}
		return m, nil
	}

	// Build the hotkey string from the tea.KeyMsg
	hotkey := msg.String()
	if hotkey == "" {
		return m, nil
	}

	if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
		idx := m.filteredIndices[m.cursor]
		old := m.settings[idx].Key
		m.settings[idx].Key = hotkey
		m.recordingHotkey = false
		m.changes = append(m.changes, changeEntry{
			Name:   m.settings[idx].Name,
			Field:  "hotkey",
			OldVal: displayKey(old),
			NewVal: hotkey,
		})
	}
	return m, nil
}
