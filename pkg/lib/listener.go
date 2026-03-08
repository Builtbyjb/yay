package lib

import (
	"fmt"
	"strconv"

	"github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/Builtbyjb/yay/pkg/lib/macos"
	hook "github.com/robotn/gohook"
)

type CustomKeyMsg struct {
	Event hook.Event
}

// TODO: Currently macos specific
func Listener(db *core.Database) {
	eventChan := hook.Start()
	defer hook.End()

	fmt.Println("Listening for global keyboard events... (Ctrl+C to quit)")

	// TODO: use a queue linked list
	var keys []uint16
	var mod string

	for event := range eventChan {
		keys = append(keys, uint16(event.Rawcode))

		// fmt.Printf("Current keys: %v\n", keys)
		// fmt.Printf("Rawcode: %d \n", event.Rawcode)

		// Tmp queue implementation to process keys in order
		key := keys[0]
		keys = keys[1:]

		if event.Kind == hook.KeyDown {

			k, err := RawcodeToString(key)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			if mod != "" {
				if mod == "l-command" || mod == "r-command" {
					if k == "l-shift" || k == "r-shift" {
						mod = "command+shift"
						continue
					}
				}

				if mod == "command+shift" {
					pos, err := strconv.ParseUint(k, 10, 16)
					if err != nil {
						fmt.Println("Error parsing position:", err)
						continue
					}

					if err := macos.LaunchDockApps(uint16(pos)); err != nil {
						fmt.Println("Error launching dock app:", err)
						continue
					}
				}

				hotkey := fmt.Sprintf("%s+%s", mod, k)

				setting, err := db.FindByHotkey(hotkey)
				if err != nil {
					fmt.Println("Error fetching setting:", err)
					continue
				}

				if setting == nil {
					continue
				}

				if setting.Enabled {
					if err := macos.Launch(setting.Path, setting.Name, setting.Mode); err != nil {
						fmt.Println("Error launching application:", err)
						continue
					}
				}
			}

			if VerifiedModifier(k) {
				mod = k
				continue
			}
		} else if event.Kind == hook.KeyUp {
			k, err := RawcodeToString(key)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			if VerifiedModifier(k) {
				mod = ""
				continue
			}
		}
	}
}
