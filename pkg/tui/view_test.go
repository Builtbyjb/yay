package tui

import (
	"testing"

	"github.com/Builtbyjb/yay/pkg/lib/core"
	tea "github.com/charmbracelet/bubbletea"
)

// ─── View Rendering ──────────────────────────────────────────────

func TestView_ContainsLogo(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "██") {
		t.Error("expected view to contain ASCII logo characters")
	}
}

func TestView_ContainsVersion(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "0.1.0") {
		t.Error("expected view to contain version string")
	}
}

func TestView_ContainsColumnHeaders(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	for _, header := range []string{"Application", "HotKey", "Mode", "Enabled"} {
		if !containsAny(view, header) {
			t.Errorf("expected view to contain header %q", header)
		}
	}
}

func TestView_ContainsSettingNames(t *testing.T) {
	database := setupTestDatabase(t)
	settings := testSettings(t, database)
	m := NewModel(nil, settings, "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	for _, s := range settings {
		if !containsAny(view, s.Name) {
			t.Errorf("expected view to contain setting name %q", s.Name)
		}
	}
}

func TestView_EmptyListMessage(t *testing.T) {
	m := NewModel(nil, []core.Setting{}, "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "No matching") {
		t.Error("expected view to show empty list message")
	}
}

func TestView_ShowsBrowseHelp(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "Navigate") {
		t.Error("expected browse mode help text")
	}
}

func TestView_ShowsFilterHelp(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "/")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "Stop Searching") {
		t.Error("expected filter mode help text")
	}
}

func TestView_ShowsRowFocusHelp(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, SWITCH_COLUMN_KEY) {
		t.Errorf("expected row focus help text with %s", SWITCH_COLUMN_KEY)
	}
}

func TestView_ShowsRecordingHotkeyHelp(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter") // focus row at colKey
	m = sendKey(t, m, " ")     // start recording
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "Press any key") {
		t.Error("expected hotkey recording help text")
	}
}

// ─── Exit Behavior ───────────────────────────────────────────────

func TestBrowse_CtrlCExits(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	_, cmd := m.Update(specialKeyMsg(tea.KeyCtrlC))

	if cmd == nil {
		t.Error("expected quit command from ctrl+c in browse mode")
		return
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestFilter_CtrlCExits(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "/") // enter filter mode

	_, cmd := m.Update(specialKeyMsg(tea.KeyCtrlC))
	if cmd == nil {
		t.Error("expected quit command from ctrl+c in filter mode")
		return
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestRowFocus_CtrlCExits(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter") // focus row

	_, cmd := m.Update(specialKeyMsg(tea.KeyCtrlC))
	if cmd == nil {
		t.Error("expected quit command from ctrl+c in row focus mode")
		return
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

// ─── Window Size ─────────────────────────────────────────────────

func TestWindowSize_UpdatesDimensions(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(model)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"hi", 5, "hi   "},
		{"hello", 5, "hello"},
		{"hello world", 5, "hello"},
		{"", 3, "   "},
	}
	for _, tc := range tests {
		result := padRight(tc.input, tc.width)
		if result != tc.expected {
			t.Errorf("padRight(%q, %d) = %q, want %q", tc.input, tc.width, result, tc.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"a long string", 8, "a lon..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
		{"hello", 5, "hello"},
	}
	for _, tc := range tests {
		result := truncate(tc.input, tc.maxLen)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.maxLen, result, tc.expected)
		}
	}
}
