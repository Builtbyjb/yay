package darwin

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

/*
Note:
* Create launch and activate functions.The launch function opens a new instance of the application.
The activate switches to the app or focuses it if it's already running.
*/

func Launch(path string, app string, mode string) error {
	binPath := filepath.Join(path, app)
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
	script := fmt.Sprintf(` tell application %q to activate `, binPath)
	exec.Command("osascript", "-e", script).Run()
	// exec.Command("open", "-a", binPath).Run()
	fmt.Println(binPath)
	return nil
}

func LaunchDockApps(pos uint16) error {
	fmt.Println("Opening a dock app at position: ", pos)
	return nil
}

// set theAppName to "Safari" -- ← change this

// if application theAppName is running then
//     tell application "System Events"
//         tell application process "Dock"
//             try
//                 click UI element theAppName of list 1
//             end try
//         end tell
//     end tell
// else
//     tell application theAppName to activate -- will also launch it
// end if

// set theAppName to "Safari"

// tell application theAppName
//     activate
// end tell

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
