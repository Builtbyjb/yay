package macos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Builtbyjb/yay/pkg/libyay/internal/helper"
)

// createTempAppDir creates a temporary directory structure mimicking a macOS
// applications directory. It returns the path to the temp dir and a cleanup
// function. The created structure per app looks like:
//
//	<tmpdir>/<appName>.app/Contents/MacOS/
//	<tmpdir>/<appName>.app/Contents/Resources/<iconFile>
func createTempAppDir(t *testing.T, apps map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()

	for appName, iconFileName := range apps {
		macosDir := filepath.Join(tmpDir, appName+".app", "Contents", "MacOS")
		if err := os.MkdirAll(macosDir, 0755); err != nil {
			t.Fatalf("Failed to create MacOS dir: %v", err)
		}

		if iconFileName != "" {
			resourcesDir := filepath.Join(tmpDir, appName+".app", "Contents", "Resources")
			if err := os.MkdirAll(resourcesDir, 0755); err != nil {
				t.Fatalf("Failed to create Resources dir: %v", err)
			}
			iconPath := filepath.Join(resourcesDir, iconFileName)
			if err := os.WriteFile(iconPath, []byte("fake icon"), 0644); err != nil {
				t.Fatalf("Failed to create icon file: %v", err)
			}
		}
	}

	return tmpDir
}

// ---------------------------------------------------------------------------
// getBinaryPath tests
// ---------------------------------------------------------------------------

func TestGetBinaryPath(t *testing.T) {
	result := getBinaryPath("/Applications", "Safari")
	expected := "/Applications/Safari/Contents/MacOS/"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetBinaryPathEmptyDir(t *testing.T) {
	result := getBinaryPath("", "Safari")
	expected := "/Safari/Contents/MacOS/"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetBinaryPathEmptyAppName(t *testing.T) {
	result := getBinaryPath("/Applications", "")
	expected := "/Applications//Contents/MacOS/"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetBinaryPathNestedDir(t *testing.T) {
	result := getBinaryPath("/Users/me/Applications", "MyApp")
	expected := "/Users/me/Applications/MyApp/Contents/MacOS/"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// ---------------------------------------------------------------------------
// getIconPath tests
// ---------------------------------------------------------------------------

func TestGetIconPathNonExistentDir(t *testing.T) {
	result := getIconPath("/nonexistent/path/that/does/not/exist", "SomeApp")
	if result != "" {
		t.Errorf("Expected empty string for non-existent dir, got %q", result)
	}
}

func TestGetIconPathDirWithNoIcnsFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some non-icns files in the dir
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hi"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "icon.png"), []byte("png"), 0644)

	result := getIconPath(tmpDir, "SomeApp")
	if result != "" {
		t.Errorf("Expected empty string when no .icns files, got %q", result)
	}
}

func TestGetIconPathFindsIcnsFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Place an .icns file directly in tmpDir (this is what the function reads)
	os.WriteFile(filepath.Join(tmpDir, "AppIcon.icns"), []byte("icon"), 0644)

	result := getIconPath(tmpDir, "MyApp")
	expected := tmpDir + "/MyApp/Contents/Resources/AppIcon.icns"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetIconPathMultipleIcnsFilesReturnsLast(t *testing.T) {
	tmpDir := t.TempDir()

	// The function iterates all files and keeps overwriting iconName,
	// so the last .icns file alphabetically by ReadDir order wins.
	os.WriteFile(filepath.Join(tmpDir, "Alpha.icns"), []byte("icon"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "Beta.icns"), []byte("icon"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "Zeta.icns"), []byte("icon"), 0644)

	result := getIconPath(tmpDir, "TestApp")

	// os.ReadDir returns entries sorted by name, so "Zeta.icns" is last
	expected := tmpDir + "/TestApp/Contents/Resources/Zeta.icns"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetIconPathEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	result := getIconPath(tmpDir, "SomeApp")
	if result != "" {
		t.Errorf("Expected empty string for empty dir, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// getApps tests
// ---------------------------------------------------------------------------

func TestGetAppsEmptyDirs(t *testing.T) {
	apps := getApps([]string{})
	if len(apps) != 0 {
		t.Errorf("Expected 0 apps, got %d", len(apps))
	}
}

func TestGetAppsNonExistentDir(t *testing.T) {
	apps := getApps([]string{"/nonexistent/dir/abc123"})
	if len(apps) != 0 {
		t.Errorf("Expected 0 apps for non-existent dir, got %d", len(apps))
	}
}

func TestGetAppsEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	apps := getApps([]string{tmpDir})
	if len(apps) != 0 {
		t.Errorf("Expected 0 apps for empty dir, got %d", len(apps))
	}
}

func TestGetAppsIgnoresNonAppEntries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some non-.app entries
	os.MkdirAll(filepath.Join(tmpDir, "SomeFolder"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hi"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "script.sh"), []byte("#!/bin/sh"), 0755)

	apps := getApps([]string{tmpDir})
	if len(apps) != 0 {
		t.Errorf("Expected 0 apps, got %d", len(apps))
	}
}

func TestGetAppsFindsAppEntries(t *testing.T) {
	tmpDir := createTempAppDir(t, map[string]string{
		"Safari":  "",
		"Firefox": "",
	})

	apps := getApps([]string{tmpDir})

	if len(apps) != 2 {
		t.Fatalf("Expected 2 apps, got %d", len(apps))
	}

	nameSet := make(map[string]bool)
	for _, app := range apps {
		nameSet[app.Name] = true
	}

	if !nameSet["Safari"] {
		t.Error("Expected Safari to be found")
	}
	if !nameSet["Firefox"] {
		t.Error("Expected Firefox to be found")
	}
}

func TestGetAppsSetsCorrectBinaryPath(t *testing.T) {
	tmpDir := createTempAppDir(t, map[string]string{
		"MyApp": "",
	})

	apps := getApps([]string{tmpDir})
	if len(apps) != 1 {
		t.Fatalf("Expected 1 app, got %d", len(apps))
	}

	expected := tmpDir + "/MyApp/Contents/MacOS/"
	if apps[0].Path != expected {
		t.Errorf("Expected path %q, got %q", expected, apps[0].Path)
	}
}

func TestGetAppsMultipleDirs(t *testing.T) {
	dir1 := createTempAppDir(t, map[string]string{
		"App1": "",
	})
	dir2 := createTempAppDir(t, map[string]string{
		"App2": "",
		"App3": "",
	})

	apps := getApps([]string{dir1, dir2})
	if len(apps) != 3 {
		t.Fatalf("Expected 3 apps, got %d", len(apps))
	}

	nameSet := make(map[string]bool)
	for _, app := range apps {
		nameSet[app.Name] = true
	}

	for _, name := range []string{"App1", "App2", "App3"} {
		if !nameSet[name] {
			t.Errorf("Expected %s to be found", name)
		}
	}
}

func TestGetAppsMixedValidAndInvalidDirs(t *testing.T) {
	tmpDir := createTempAppDir(t, map[string]string{
		"GoodApp": "",
	})

	apps := getApps([]string{"/nonexistent/dir", tmpDir})
	if len(apps) != 1 {
		t.Fatalf("Expected 1 app, got %d", len(apps))
	}

	if apps[0].Name != "GoodApp" {
		t.Errorf("Expected name %q, got %q", "GoodApp", apps[0].Name)
	}
}

func TestGetAppsMixedAppAndNonAppEntries(t *testing.T) {
	tmpDir := createTempAppDir(t, map[string]string{
		"RealApp": "",
	})

	// Add a non-.app directory alongside the .app
	os.MkdirAll(filepath.Join(tmpDir, "NotAnApp"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("data"), 0644)

	apps := getApps([]string{tmpDir})
	if len(apps) != 1 {
		t.Fatalf("Expected 1 app, got %d", len(apps))
	}

	if apps[0].Name != "RealApp" {
		t.Errorf("Expected name %q, got %q", "RealApp", apps[0].Name)
	}
}

func TestGetAppsNilDirs(t *testing.T) {
	apps := getApps(nil)
	if len(apps) != 0 {
		t.Errorf("Expected 0 apps for nil dirs, got %d", len(apps))
	}
}

// ---------------------------------------------------------------------------
// GetSettings tests (integration with a real in-memory database)
// ---------------------------------------------------------------------------

func setupTestDatabase(t *testing.T) *helper.Database {
	t.Helper()
	db, err := helper.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}
	return db
}

func TestGetSettingsEmptyDir(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	tmpDir := t.TempDir()

	settings, err := GetSettings(*db, []string{tmpDir})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(settings) != 0 {
		t.Errorf("Expected 0 settings, got %d", len(settings))
	}
}

func TestGetSettingsWithApps(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	tmpDir := createTempAppDir(t, map[string]string{
		"TestApp1": "",
		"TestApp2": "",
	})

	settings, err := GetSettings(*db, []string{tmpDir})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(settings) != 2 {
		t.Fatalf("Expected 2 settings, got %d", len(settings))
	}

	nameSet := make(map[string]bool)
	for _, s := range settings {
		nameSet[s.Name] = true

		// Verify default values
		if s.HotKey != "" {
			t.Errorf("Expected empty hotkey for %s, got %q", s.Name, s.HotKey)
		}
		if s.Mode != "default" {
			t.Errorf("Expected mode 'default' for %s, got %q", s.Name, s.Mode)
		}
		if !s.Enabled {
			t.Errorf("Expected enabled=true for %s", s.Name)
		}
		if s.Id == 0 {
			t.Errorf("Expected non-zero id for %s", s.Name)
		}
	}

	if !nameSet["TestApp1"] {
		t.Error("Expected TestApp1 to be in settings")
	}
	if !nameSet["TestApp2"] {
		t.Error("Expected TestApp2 to be in settings")
	}
}

func TestGetSettingsRefreshRemovesStaleApps(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	tmpDir := createTempAppDir(t, map[string]string{
		"App1": "",
		"App2": "",
	})

	// First call seeds both apps
	settings, err := GetSettings(*db, []string{tmpDir})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(settings) != 2 {
		t.Fatalf("Expected 2 settings, got %d", len(settings))
	}

	// Remove App1.app from the filesystem
	os.RemoveAll(filepath.Join(tmpDir, "App1.app"))

	// Second call should remove App1 from the database
	settings, err = GetSettings(*db, []string{tmpDir})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(settings) != 1 {
		t.Fatalf("Expected 1 setting, got %d", len(settings))
	}
	if settings[0].Name != "App2" {
		t.Errorf("Expected remaining app to be App2, got %q", settings[0].Name)
	}
}

func TestGetSettingsNonExistentDir(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	settings, err := GetSettings(*db, []string{"/nonexistent/dir/xyz"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(settings) != 0 {
		t.Errorf("Expected 0 settings, got %d", len(settings))
	}
}

func TestGetSettingsMultipleDirs(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	dir1 := createTempAppDir(t, map[string]string{
		"AppA": "",
	})
	dir2 := createTempAppDir(t, map[string]string{
		"AppB": "",
		"AppC": "",
	})

	settings, err := GetSettings(*db, []string{dir1, dir2})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(settings) != 3 {
		t.Fatalf("Expected 3 settings, got %d", len(settings))
	}

	nameSet := make(map[string]bool)
	for _, s := range settings {
		nameSet[s.Name] = true
	}

	for _, name := range []string{"AppA", "AppB", "AppC"} {
		if !nameSet[name] {
			t.Errorf("Expected %s to be in settings", name)
		}
	}
}

func TestGetSettingsNilDirs(t *testing.T) {
	db := setupTestDatabase(t)
	defer db.Close()

	settings, err := GetSettings(*db, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(settings) != 0 {
		t.Errorf("Expected 0 settings for nil dirs, got %d", len(settings))
	}
}
