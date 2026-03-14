package darwin

import (
	"fmt"
	"slices"
	"strconv"
	"sync"

	"github.com/Builtbyjb/yay/pkg/lib/core"
)

// Listener starts the global key event tap. An optional onEvent callback
// is called for every event (e.g. to forward to a tea.Program).
// This function blocks forever.
func Listener(db *core.Database, onEvent func(KeyEvent)) {
	var mod string
	var mu sync.Mutex

	SetKeyHandler(func(event KeyEvent) bool {
		// Forward to the TUI if a callback is provided
		if onEvent != nil {
			onEvent(event)
		}

		mu.Lock()
		defer mu.Unlock()

		k, ok := RawToKeyDarwin[event.Keycode]
		if !ok {
			return false // unknown key, pass through
		}

		switch event.EventType {
		case EventFlagsChanged:
			if slices.Contains(ModifiersMacos, k) {
				if IsModifierPressed(event.Flags, event.Keycode) {
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

		case EventKeyDown:
			if mod == "" {
				return false
			}

			if mod == "command+shift" {
				pos, err := strconv.ParseUint(k, 10, 16)
				if err == nil {
					go func() {
						if err := LaunchDockApps(uint16(pos)); err != nil {
							fmt.Println("Error launching dock app:", err)
						}
					}()
					return true
				}
			}

			hotkey := fmt.Sprintf("%s+%s", mod, k)

			if hotkey == "l-command+esc" || hotkey == "r-command+esc" {
				go func() {
					SwitchToDefaultDesktop()
				}()
			}

			setting, err := db.FindByHotkey(hotkey)
			if err != nil {
				fmt.Println("Error fetching setting:", err)
				return false
			}

			if setting != nil && setting.Enabled {
				go func() {
					if err := Launch(setting.Path, setting.Name, setting.Mode); err != nil {
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
	StartEventTap()
}
