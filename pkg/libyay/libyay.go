package libyay

import (
	"fmt"
	"runtime"

	"github.com/Builtbyjb/yay/pkg/libyay/internal/helper"
	"github.com/Builtbyjb/yay/pkg/libyay/internal/macos"
)

func Fetch() ([]helper.Setting, error) {

	switch runtime.GOOS {
	case "darwin":
		database := helper.NewDatabase(macos.DatabaseDirectory)
		defer database.Close()

		dirs := macos.AppDirectories
		settings, err := macos.GetSettings(*database, dirs)
		if err != nil {
			return nil, err
		}
		return settings, nil

	case "windows":
		fmt.Println("Coming Soon...")
		return nil, nil
	default:
		fmt.Println("Unsupported operating system.")
		return nil, nil
	}
}

// Update function
