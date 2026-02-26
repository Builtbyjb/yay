package libyay

import (
	"fmt"
	"runtime"

	"github.com/Builtbyjb/yay/pkg/libyay/internal/helper"
	"github.com/Builtbyjb/yay/pkg/libyay/internal/macos"
)

func Fetch() []helper.Setting {
	database := helper.NewDatabase("yay.db")

	switch runtime.GOOS {
	case "darwin":
		dirs := macos.AppDirectories
		settings := macos.GetSettings(*database, dirs)
		return settings

	case "windows":
		fmt.Println("Coming Soon...")
		return nil
	default:
		fmt.Println("Unsupported operating system.")
		return nil
	}
}

// Update function
