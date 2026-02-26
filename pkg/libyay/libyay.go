package libyay

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/Builtbyjb/yay/pkg/libyay/internal/helper"
	"github.com/Builtbyjb/yay/pkg/libyay/internal/macos"
)

// Setting is re-exported from the internal helper package so external
type Setting = helper.Setting

func Fetch() ([]helper.Setting, error) {

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	switch runtime.GOOS {
	case "darwin":
		dbPath := filepath.Join(usr.HomeDir, "Library", "Application Support", "Yay", "db.sqlite3")

		dbDir := filepath.Dir(dbPath)
		// 0755
		// │││└─ Others: 5 (read + execute)
		// ││└── Group: 5 (read + execute)
		// │└─── Owner: 7 (read + write + execute)
		// └──── Octal prefix: 0
		err = os.MkdirAll(dbDir, 0755)
		if err != nil {
			return nil, err
		}

		database, err := helper.NewDatabase(dbPath)
		if err != nil {
			return nil, err
		}
		defer database.Close()

		if err := database.Init(); err != nil {
			return nil, err
		}

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
