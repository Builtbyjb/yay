package tui

import (
	"database/sql"
	"testing"

	"github.com/Builtbyjb/yay/pkg/lib/core"
	tea "github.com/charmbracelet/bubbletea"
)

func setupTestDatabase(t *testing.T) *core.Database {
	t.Helper()
	database, err := core.NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	if err := database.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	return database
}

// testSettings returns a reusable slice of settings for tests.
func testSettings(t *testing.T, db *core.Database) []core.Setting {
	var err error
	err = db.Insert("Firefox", "/path/to/firefox", "/icon/firefox.icns", sql.NullString{String: "ctrl+1", Valid: true}, "default", true)
	err = db.Insert("Terminal", "/path/to/terminal", "/icon/terminal.icns", sql.NullString{String: "", Valid: false}, "desktop", true)
	err = db.Insert("Finder", "/path/to/finder", "/icon/finder.icns", sql.NullString{String: "ctrl+3", Valid: true}, "desktop", false)
	err = db.Insert("Safari", "/path/to/safari", "/icon/safari.icns", sql.NullString{String: "", Valid: false}, "default", true)
	err = db.Insert("Notes", "/path/to/notes", "/icon/notes.icns", sql.NullString{String: "alt+n", Valid: true}, "default", true)

	if err != nil {
		t.Fatalf("Failed to insert test settings: %v", err)
	}

	settings, getErr := db.GetAllSettings()

	if getErr != nil {
		t.Fatalf("Failed to get test settings: %v", getErr)
	}

	return settings
}

func keyMsg(key string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
}

func specialKeyMsg(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

// sendKey is a small helper that calls Update with a key message and returns
// the resulting model, failing the test if the type assertion fails.
func sendKey(t *testing.T, m model, key string) model {
	t.Helper()
	var result tea.Model
	switch key {
	case "up":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	case "down":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	case "enter":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	case "esc":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	case "backspace":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	case "tab":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	case "ctrl+c":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	case "ctrl+q":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlQ})
	case " ":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	case "right":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	case "left":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	default:
		result, _ = m.Update(keyMsg(key))
	}
	return result.(model)
}

func TestNewModel_DefaultState(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")

	if m.state != stateBrowse {
		t.Errorf("expected initial state stateBrowse (%d), got %d", stateBrowse, m.state)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
	if m.activeCol != colNone {
		t.Errorf("expected activeCol colNone, got %d", m.activeCol)
	}
	if m.version != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %s", m.version)
	}
}

func TestNewModel_AllSettingsVisible(t *testing.T) {
	database := setupTestDatabase(t)
	settings := testSettings(t, database)
	m := NewModel(nil, settings, "0.1.0")

	if len(m.searchedIndices) != len(settings) {
		t.Errorf("expected %d filtered indices, got %d", len(settings), len(m.searchedIndices))
	}
}

func TestNewModel_EmptySettings(t *testing.T) {
	m := NewModel(nil, []core.Setting{}, "1.0.0")

	if len(m.searchedIndices) != 0 {
		t.Errorf("expected 0 filtered indices, got %d", len(m.searchedIndices))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

func TestNewModel_PreservesSettingData(t *testing.T) {
	database := setupTestDatabase(t)
	settings := testSettings(t, database)
	m := NewModel(nil, settings, "0.1.0")

	if m.settings[0].Name != "Finder" {
		t.Errorf("expected first setting name Finder, got %s", m.settings[0].Name)
	}
	if m.settings[2].Enabled != true {
		t.Errorf("expected Finder to be enabled")
	}
	if m.settings[1].Mode != "default" {
		t.Errorf("expected Terminal mode default, got %s", m.settings[1].Mode)
	}
}

// ─── Helper ──────────────────────────────────────────────────────

func containsAny(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) > 0 && contains(haystack, needle)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
