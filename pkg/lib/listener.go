package lib

import (
	"fmt"

	hook "github.com/robotn/gohook"
)

type CustomKeyMsg struct {
	Event hook.Event
}

/*
Features:
-> If a key is mapped to an app pressing the same key toggles between opened windows of the app
-> Pressing command+shift on macOS activates the dock apps and opens them in order
->  You can assign hotkeys to applications and set mode
*/

func Listener() {
	eventChan := hook.Start()
	defer hook.End()

	fmt.Println("Listening for global keyboard events... (Ctrl+C to quit)")

	for event := range eventChan {
		if event.Kind == hook.KeyDown && event.Rawcode != 0 {
			fmt.Printf("Rawcode: %d \n", event.Rawcode)
		}
	}

	/*
		var mod string
		var alt string
		var key string

		for e in keyEvent:
			ks := RawcodeToKeyString(e.rawcode)
			if AcceptedModifier(Ks):
				mod = ks
				continue
			if mod != "":
				if AcceptedAlt(ks):
					alt = ks
					continue
				if alt != "":
					hotkey = fmt.Sprintf("%s+%s+%s", mod, alt, ks)
					DoSomething(hotkey)
				else:
					hotkey = fmt.Sprintf("%s+%s", mod, ks)
					DoSomethingElse(hotkey)

	*/
}
