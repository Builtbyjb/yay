package macos

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

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
