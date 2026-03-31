//go:build darwin

package lib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"text/template"
	"time"

	"github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/Builtbyjb/yay/pkg/lib/darwin"
)

func getBinaryPath() string {
	bin_path, _ := os.Executable()
	return bin_path
}

func GetDatabase() (*core.Database, error) {

	dbPath, err := darwin.GetDatabasePath()
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
}

func Fetch() (*core.Database, []core.Setting, error) {
	db, err := GetDatabase()
	if err != nil {
		return nil, nil, err
	}

	dirs := darwin.AppDirectories
	settings, err := darwin.GetSettings(*db, dirs)
	if err != nil {
		return nil, nil, err
	}
	return db, settings, nil
}

func RawcodeToString(rawcode uint16) (string, error) {
	key, ok := darwin.RawToKeyDarwin[rawcode]
	if !ok {
		return "", fmt.Errorf("unknown rawcode: %d", rawcode)
	}
	return key, nil
}

func KeyEventListener(db *core.Database, onEvent func(KeyEvent)) {
	darwin.Listener(db, func(de darwin.KeyEvent) {
		if onEvent != nil {
			onEvent(KeyEvent{
				Keycode:   de.Keycode,
				Flags:     de.Flags,
				EventType: de.EventType,
			})
		}
	})
}

func VerifiedModifier(key string) bool {
	return slices.Contains(darwin.ModifiersMacos, key)
}

func StartBackgroundService() {
	fmt.Println("Starting background service...")

	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, "/Library/logs", "yay")
	os.Mkdir(logDir, 0755)

	data := darwin.PlistData{
		Label:      darwin.LABEL,
		BinaryPath: getBinaryPath(),
		LogDir:     logDir,
		WorkingDir: filepath.Dir(getBinaryPath()),
	}

	// Create the launch agents' plist file
	t := template.Must(template.New("plist").Parse(darwin.PlistTemplate))
	f, err := os.Create(darwin.PlistPath())
	if err != nil {
		fmt.Println("Error creating plist file:", err)
		return
	}

	if err := t.Execute(f, data); err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	// Close before setting permissions
	f.Close()

	// Set permissions to 0644 (rw-r--r--)
	if err := os.Chmod(darwin.PlistPath(), 0644); err != nil {
		fmt.Println("Error setting plist permissions:", err)
		return
	}

	// uid := fmt.Sprintf("gui/%d", os.Getuid())
	serviceId := fmt.Sprintf("gui/%d/%s", os.Getuid(), darwin.LABEL)

	err = exec.Command("launchctl", "bootout", serviceId).Run()
	if err != nil {
		fmt.Println("No existing service to unload, proceeding to load new service.")
	}

	time.Sleep(500 * time.Millisecond)

	cmd := exec.Command("launchctl", "load", darwin.PlistPath())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Error loading plist:", err)
		os.Exit(1)
	}

	start := exec.Command("launchctl", "start", darwin.LABEL)
	start.Stdout = os.Stdout
	start.Stderr = os.Stderr

	if err := start.Run(); err != nil {
		fmt.Println("Error starting service:", err)
		os.Exit(1)
	}
}

func StopBackgroundService() {
	fmt.Println("Stopping background service...")

	serviceId := fmt.Sprintf("gui/%d/%s", os.Getuid(), darwin.LABEL)

	cmd := exec.Command("launchctl", "bootout", serviceId)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("Error unloading plist:", err)
		os.Exit(1)
	}
}

func CheckBackgroundServiceStatus() {
	out, _ := exec.Command("launchctl", "list", darwin.LABEL).CombinedOutput()
	fmt.Println(string(out))
}
