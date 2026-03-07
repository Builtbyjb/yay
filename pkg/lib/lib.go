package lib

import (
	"fmt"
	"runtime"
	"slices"

	"github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/Builtbyjb/yay/pkg/lib/macos"
)

func GetDatabase() (*core.Database, error) {

	switch runtime.GOOS {
	case "darwin":
		dbPath, err := macos.GetDatabasePath()
		if err != nil {
			return nil, err
		}

		db, err := core.NewDatabase(dbPath)
		if err != nil {
			return nil, err
		}

		if err := db.Init(); err != nil {
			return nil, err
		}

		return db, nil
	case "windows":
		fmt.Println("Coming Soon...")
		return nil, nil
	case "linux":
		fmt.Println("Coming Soon...")
		return nil, nil
	default:
		fmt.Println("Unsupported operating system.")
		return nil, nil
	}
}

func Fetch() (*core.Database, []core.Setting, error) {

	switch runtime.GOOS {
	case "darwin":
		db, err := GetDatabase()
		if err != nil {
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

func RawcodeToString(rawcode uint16) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		key, ok := macos.RawToKeyDarwin[rawcode]
		if !ok {
			return "", fmt.Errorf("unknown rawcode: %d", rawcode)
		}
		return key, nil
	case "windows":
		return "", fmt.Errorf("windows not supported yet")
	case "linux":
		return "", fmt.Errorf("linux not supported yet")
	default:
		return "", fmt.Errorf("unsupported operating system")
	}
}

func VerifiedModifier(mod string) bool {
	switch runtime.GOOS {
	case "darwin":
		if slices.Contains(macos.ModifiersMacos, mod) {
			return true
		}
		return false
	case "windows":
		return false
	case "linux":
		return false
	default:
		return false
	}
}
