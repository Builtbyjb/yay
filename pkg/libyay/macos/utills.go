package macos

import (
	"fmt"
	"os"
	"strings"

	"github.com/Builtbyjb/yay/pkg/libyay/helper"
)

func GetBinaryPath(dir string, app_name string) string {
	return dir + "/" + app_name + "/Contents/MacOS/"
}

func GetApps(dirs []string) []helper.App {
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("Error reading directory %s: %v\n", dir, err)
			continue
		}

		for _, entry := range entries {
			// Store app name without the .app extension (it is the same as binary name)
			if strings.HasSuffix(entry.Name(), ".app") {
				part := GetBinaryPath(dir, entry.Name())
				fmt.Printf("Found application: %s in %s\n", entry.Name(), part)
			}
		}
	}

	return []helper.App{}
}
