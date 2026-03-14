package tui

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/Builtbyjb/yay/pkg/lib"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateBrowse:
			return m.HandleBrowseKey(msg)
		case stateFilter:
			return m.SearchUpdate(msg)
		case stateRowFocus:
			return m.handleRowFocusKey(msg)
		}
		return m, nil
	case lib.CKeyMsg:
		return m.RecordKey(msg)
	}
	return m, nil
}

func (m model) RecordKey(msg lib.CKeyMsg) (tea.Model, tea.Cmd) {
	if m.recordingHotkey {
		m.keys = append(m.keys, msg.Event.Keycode)
		// m.debug = append(m.debug, int(msg.Event.Keycode))

		m.key = m.keys[0]
		m.keys = m.keys[1:]
		k, err := lib.RawcodeToString(m.key)

		if err != nil {
			m.errors = append(m.errors, fmt.Sprintf("Unknown modifier key: %s", k))
			m.recordingHotkey = false
		}

		switch msg.Event.EventType {
		case lib.EventKeyDown:
			if m.mod != "" {
				if len(m.searchedIndices) > 0 && m.cursor < len(m.searchedIndices) {
					hotkey := fmt.Sprintf("%s+%s", m.mod, k)
					idx := m.searchedIndices[m.cursor]
					// m.errors = append(m.errors, hotkey)
					m.settings[idx].HotKey = sql.NullString{String: hotkey, Valid: true}
					if err := m.db.UpdateHotkey(m.settings[idx].Id, m.settings[idx].HotKey); err != nil {
						m.errors = append(m.errors, err.Error())
					}
					m.recordingHotkey = false
					m.mod = ""

					return m, nil
				}
			}

		case lib.EventFlagsChanged:
			if lib.VerifiedModifier(k) {
				m.mod = k
			}
		}
	}

	return m, nil
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
	currentIdx := slices.Index(AvailableModes, prev)
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

func (m model) SearchUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case EXIT_KEY:
		return m, tea.Quit

	case CANCEL_KEY:
		m.state = stateBrowse
		m.searchInput.Blur()
		return m, nil

	case "up":
		m.moveCursor(-1)
		return m, nil

	case "down":
		m.moveCursor(1)
		return m, nil

	case "enter":
		if len(m.searchedIndices) > 0 {
			m.state = stateRowFocus
			m.activeCol = colKey
			m.recordingHotkey = false
			m.searchInput.Blur()
		}
		return m, nil
	}

	// Pass key to text input for filtering
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.updateFilter()
	// Ensure cursor is in bounds after filter changes
	if m.cursor >= len(m.searchedIndices) {
		if len(m.searchedIndices) > 0 {
			m.cursor = len(m.searchedIndices) - 1
		} else {
			m.cursor = 0
		}
	}
	return m, cmd
}

func (m *model) updateFilter() {
	query := strings.ToLower(m.searchInput.Value())
	m.searchedIndices = make([]int, 0, len(m.settings))
	for i, s := range m.settings {
		if query == "" || strings.Contains(strings.ToLower(s.Name), query) {
			m.searchedIndices = append(m.searchedIndices, i)
		}
	}
	// Clamp cursor to stay within the new filtered list
	if len(m.searchedIndices) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.searchedIndices) {
		m.cursor = len(m.searchedIndices) - 1
	}
}
