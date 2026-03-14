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
	// script := fmt.Sprintf(`
	// 	tell application %q to activate
	// 	tell application "System Events"
	// 		tell process %q
	//                keystroke "f" using {control down, command down}
	//            end tell
	// 	end tell
	//    `, binPath, binPath)
	// script := `
	// 	tell application "System Events"
	// 		set braveProcess to first process whose name is "Brave Browser"
	// 		set frontmost of braveProcess to true
	// 	end tell
	// `
	script := fmt.Sprintf(`
		set appName to %q

		-- Open an application if minimized
		tell application "System Events"
			tell application process "Dock"
				click UI element appName of list 1
			end tell
		end tell

		tell application appName
			activate
		end tell
		`, app)
	err := exec.Command("osascript", "-e", script).Run()
	if err != nil {
		fmt.Println(err.Error())
	}
	return nil
}

func LaunchDockApps(pos uint16) error {
	fmt.Println("Opening a dock app at position: ", pos)
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
