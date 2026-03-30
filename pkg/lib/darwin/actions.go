package darwin

import (
	"fmt"
	"os/exec"
)

/*
Note:
* Create launch and activate functions.The launch function opens a new instance of the application.
The activate switches to the app or focuses it if it's already running.
*/

func Launch(app string, mode string) error {
	script := fmt.Sprintf(`
		set appName to %q

		-- Open an application if minimized
		try
			tell application "System Events"
				tell application process "Dock"
					click UI element appName of list 1
				end tell
			end tell
		end try

		tell application appName
			reopen
			activate
		end tell
		`, app)
	err := exec.Command("osascript", "-e", script).Run()
	if err != nil {
		fmt.Println("Error launching application:", err)
		fmt.Println(err.Error())
	}
	return nil
}

/* Open applications on the dock */
func LaunchDockApps(pos uint16) error {
	fmt.Println("Opening a dock app at position:", pos)
	script := fmt.Sprintf(`
		tell application "System Events"
			tell application process "Dock"
				set dockItems to every UI element of list 1
				if %d ≤ (count of dockItems) then
					click UI element %d of list 1
				else
					error "Dock position out of range"
				end if
			end tell
		end tell
	`, pos, pos)
	err := exec.Command("osascript", "-e", script).Run()
	if err != nil {
		return fmt.Errorf("error clicking dock app at position %d: %w", pos, err)
	}
	return nil
}

func SwitchToDefaultDesktop() {
	script := `
		tell application "System Events"
   			-- key code 18 using {control down} -- key code 18 = "1"
        	-- keystroke "1" using control down
        	tell application "Finder" to activate
         end tell
        `
	exec.Command("osascript", "-e", script).Run()
}

// delay 0.2 -- small breathing room

// tell application "System Events"
//     tell process theAppName
//         if exists window 1 then
//             set isFullScreen to value of attribute "AXFullScreen" of window 1

//             if not isFullScreen then
//                 -- Method A: most compatible (simulates ⌃⌘F)
//                 keystroke "f" using {control down, command down}

//                 -- Method B: if the app supports the property directly (Safari, Mail, etc.)
//                 -- set value of attribute "AXFullScreen" of window 1 to true
//             end if
//         end if
//     end tell
// end tell
