package macos

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
	return dir + "/" + appName + "/Contents/MacOS/"
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

	return dir + "/" + appName + "/Contents/Resources/" + iconName
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
					Path:     getBinaryPath(dir, appName),
					IconPath: getIconPath(dir, appName),
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
