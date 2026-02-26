package macos

import (
	"fmt"
	"os"
	"strings"

	"github.com/Builtbyjb/yay/pkg/libyay/internal/helper"
)

func GetBinaryPath(dir string, app_name string) string {
	return dir + "/" + app_name + "/Contents/MacOS/"
}

func GetIconPath(dir string, app_name string) string {
	// NOTE: Icon names vary
	// This is a temporary fix
	return dir + "/" + app_name + "/Contents/Resources/apps.icns"
}

func GetSettings(database helper.Database, dirs []string) []helper.Setting {
	apps := getApps(dirs)
	settings := database.Refresh(apps)
	return settings
}

func getApps(dirs []string) []helper.App {
	apps := []helper.App{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("Error reading directory %s: %v\n", dir, err)
			continue
		}

		for _, entry := range entries {
			app_name, ok := strings.CutSuffix(entry.Name(), ".app")
			if ok {
				apps = append(apps, helper.App{
					Name:     app_name,
					Path:     GetBinaryPath(dir, app_name),
					IconPath: GetIconPath(dir, app_name),
				})
			}
		}
	}
	return apps
}
