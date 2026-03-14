package darwin

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Builtbyjb/yay/pkg/lib/core"
)

func GetSettings(database core.Database, dirs []string) ([]core.Setting, error) {
	apps := getApps(dirs)
	settings, err := database.Refresh(apps)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func getBinaryPath(dir string, appName string) string {
	return filepath.Join(dir, appName, "Contents", "MacOS")
}

func getIconPath(dir string, appName string) string {
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", dir, err)
		return ""
	}

	var iconName string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".icns") {
			iconName = file.Name()
		}
	}

	if iconName == "" {
		return ""
	}

	return filepath.Join(dir, appName, "Contents", "Resources", iconName)
}

func getApps(dirs []string) []core.App {
	apps := []core.App{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("Error reading directory %s: %v\n", dir, err)
			continue
		}

		for _, entry := range entries {
			appName, ok := strings.CutSuffix(entry.Name(), ".app")
			if ok {
				apps = append(apps, core.App{
					Name:     appName,
					Path:     getBinaryPath(dir, entry.Name()),
					IconPath: getIconPath(dir, entry.Name()),
				})
			}
		}
	}
	return apps
}

func GetDatabasePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	dbPath := filepath.Join(usr.HomeDir, "Library", "Application Support", "Yay", "db.sqlite3")

	dbDir := filepath.Dir(dbPath)
	// 0755
	// │││└─ Others: 5 (read + execute)
	// ││└── Group: 5 (read + execute)
	// │└─── Owner: 7 (read + write + execute)
	// └──── Octal prefix: 0
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return "", err
	}

	return dbPath, nil
}

// isModifierPressed checks whether the modifier flag corresponding to the
// given keycode is currently set in the CGEvent flags bitmask.
func IsModifierPressed(flags uint64, keycode uint16) bool {
	switch keycode {
	case 55, 54: // l-command, r-command
		return flags&0x100000 != 0
	case 56, 60: // l-shift, r-shift
		return flags&0x020000 != 0
	case 58, 61: // l-option, r-option
		return flags&0x080000 != 0
	case 59: // control
		return flags&0x040000 != 0
	default:
		return false
	}
}
