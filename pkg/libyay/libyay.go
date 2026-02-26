package libyay

import (
	"fmt"
	"runtime"

	"github.com/Builtbyjb/yay/pkg/libyay/helper"
	"github.com/Builtbyjb/yay/pkg/libyay/macos"
)

func Fetch() []helper.App {
	switch runtime.GOOS {
	case "darwin":
		fmt.Println("Fetching macOS applications...")
		// Create a apps list
		// Return the apps list
		dirs := macos.AppDirectories
		apps := macos.GetApps(dirs)
		return apps

	case "windows":
		fmt.Println("Fetching Windows applications...")
		// Fetch Windows applications
		// apps := windows.FetchApps()
		// return apps
		return nil
	default:
		fmt.Println("Unsupported operating system.")
		return nil
	}
}
