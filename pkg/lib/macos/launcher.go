package macos

import (
	"fmt"
	"path/filepath"
)

func Launch(path string, app string, mode string) error {
	binPath := filepath.Join(path, app)
	// cmd := exec.Command("open", binPath)
	// cmd.Run()
	fmt.Println(binPath)
	return nil
}

func LaunchDockApps(pos uint16) error {
	fmt.Println("Opening a dock app at position: ", pos)
	return nil
}
