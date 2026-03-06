package lib

import (
	"fmt"
	"runtime"

	"github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/Builtbyjb/yay/pkg/lib/macos"
)

func Fetch() (*core.Database, []core.Setting, error) {

	switch runtime.GOOS {
	case "darwin":
		dbPath, err := macos.GetDatabasePath()
		if err != nil {
			return nil, nil, err
		}

		db, err := core.NewDatabase(dbPath)
		if err != nil {
			return nil, nil, err
		}

		if err := db.Init(); err != nil {
			return nil, nil, err
		}

		dirs := macos.AppDirectories
		settings, err := macos.GetSettings(*db, dirs)
		if err != nil {
			return nil, nil, err
		}
		return db, settings, nil

	case "windows":
		fmt.Println("Coming Soon...")
		return nil, nil, nil
	default:
		fmt.Println("Unsupported operating system.")
		return nil, nil, nil
	}
}
