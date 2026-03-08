package tui

import (
	"database/sql"
	"fmt"

	"github.com/Builtbyjb/yay/pkg/lib"
	"github.com/Builtbyjb/yay/pkg/lib/darwin"
	tea "github.com/charmbracelet/bubbletea"
)

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
		case darwin.EventKeyDown:
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

		case darwin.EventFlagsChanged:
			if lib.VerifiedModifier(k) {
				m.mod = k
			}
		}
	}

	return m, nil
}
