package lib

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/Builtbyjb/yay/pkg/lib/darwin"
)

type CKeyMsg struct {
	Event darwin.KeyEvent
}

// isModifierPressed checks whether the modifier flag corresponding to the
// given keycode is currently set in the CGEvent flags bitmask.
func isModifierPressed(flags uint64, keycode uint16) bool {
	switch keycode {
	case 55, 54: // l-command, r-command
		return flags&0x100000 != 0
	case 56, 60: // l-shift, r-shift
		return flags&0x020000 != 0
	case 58, 61: // l-option, r-option
		return flags&0x080000 != 0
	case 59: // control
		return flags&0x040000 != 0
	default:
		return false
	}
}

// Listener starts the global key event tap. An optional onEvent callback
// is called for every event (e.g. to forward to a tea.Program).
// This function blocks forever.
func Listener(db *core.Database, onEvent func(darwin.KeyEvent)) {
	var mod string
	var mu sync.Mutex

	darwin.SetKeyHandler(func(event darwin.KeyEvent) bool {
		// Forward to the TUI if a callback is provided
		if onEvent != nil {
			onEvent(event)
		}

		mu.Lock()
		defer mu.Unlock()

		k, err := RawcodeToString(event.Keycode)
		if err != nil {
			return false // unknown key, pass through
		}

		switch event.EventType {
		case darwin.EventFlagsChanged:
			if VerifiedModifier(k) {
				if isModifierPressed(event.Flags, event.Keycode) {
					if mod == "" {
						mod = k
					} else if (mod == "l-command" || mod == "r-command") &&
						(k == "l-shift" || k == "r-shift") {
						mod = "command+shift"
					}
				} else {
					mod = ""
				}
			}
			return false

		case darwin.EventKeyDown:
			if mod == "" {
				return false
			}

			if mod == "command+shift" {
				pos, err := strconv.ParseUint(k, 10, 16)
				if err == nil {
					go func() {
						if err := darwin.LaunchDockApps(uint16(pos)); err != nil {
							fmt.Println("Error launching dock app:", err)
						}
					}()
					return true
				}
			}

			hotkey := fmt.Sprintf("%s+%s", mod, k)

			if hotkey == "l-command+esc" || hotkey == "r-command+esc" {
				go func() {
					darwin.SwitchToDefaultDesktop()
				}()
			}

			setting, err := db.FindByHotkey(hotkey)
			if err != nil {
				fmt.Println("Error fetching setting:", err)
				return false
			}

			if setting != nil && setting.Enabled {
				go func() {
					if err := darwin.Launch(setting.Path, setting.Name, setting.Mode); err != nil {
						fmt.Println("Error launching application:", err)
					}
				}()
				return true
			}

			return false
		}

		return false
	})

	fmt.Println("Listening for global keyboard events... (Ctrl+C to quit)")
	darwin.StartEventTap()
}
