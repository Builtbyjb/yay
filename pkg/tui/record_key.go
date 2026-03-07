package tui

import (
	"database/sql"
	"fmt"

	"github.com/Builtbyjb/yay/pkg/lib"
	tea "github.com/charmbracelet/bubbletea"
	hook "github.com/robotn/gohook"
)

func (m model) RecordKey(msg lib.CustomKeyMsg) (tea.Model, tea.Cmd) {
	if m.recordingHotkey && msg.Event.Kind == hook.KeyDown {

		m.keys = append(m.keys, uint16(msg.Event.Rawcode))
		m.debug = append(m.debug, int(msg.Event.Rawcode))

		if len(m.keys) == 3 {
			var hotkey string
			mod, err := lib.RawcodeToString(m.keys[1])
			if err != nil {
				m.errors = append(m.errors, fmt.Sprintf("Unknown modifier key: %d", m.keys[1]))
				m.recordingHotkey = false
			}

			key, err := lib.RawcodeToString(m.keys[2])
			if err != nil {
				m.errors = append(m.errors, fmt.Sprintf("Unknown key: %d", m.keys[2]))
				m.recordingHotkey = false
			}

			if lib.VerifiedModifier(mod) {
				if len(m.searchedIndices) > 0 && m.cursor < len(m.searchedIndices) {
					hotkey = fmt.Sprintf("%s+%s", mod, key)
					idx := m.searchedIndices[m.cursor]
					m.errors = append(m.errors, hotkey)
					m.settings[idx].HotKey = sql.NullString{String: hotkey, Valid: true}
					if err := m.db.UpdateHotkey(m.settings[idx].Id, m.settings[idx].HotKey); err != nil {
						m.errors = append(m.errors, err.Error())
					}
					m.recordingHotkey = false

					// Reset keys
					m.keys = []uint16{}
					return m, nil
				}
			} else {
				m.errors = append(m.errors, fmt.Sprintf("Invalid modifier key: %s", mod))
				m.recordingHotkey = false
			}
		}
	}
	return m, nil
}
