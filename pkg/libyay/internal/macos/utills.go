package macos

import (
	"fmt"
	"os"
	"strings"

	"github.com/Builtbyjb/yay/pkg/libyay/internal/helper"
)

func GetSettings(database helper.Database, dirs []string) ([]helper.Setting, error) {
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

func getApps(dirs []string) []helper.App {
	apps := []helper.App{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("Error reading directory %s: %v\n", dir, err)
			continue
		}

		for _, entry := range entries {
			appName, ok := strings.CutSuffix(entry.Name(), ".app")
			if ok {
				apps = append(apps, helper.App{
					Name:     appName,
					Path:     getBinaryPath(dir, appName),
					IconPath: getIconPath(dir, appName),
				})
			}
		}
	}
	return apps
}
