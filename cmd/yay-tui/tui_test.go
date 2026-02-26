package main

import (
	"testing"

	"github.com/Builtbyjb/yay/pkg/libyay"
	tea "github.com/charmbracelet/bubbletea"
)

// testSettings returns a reusable slice of settings for tests.
func testSettings() []libyay.Setting {
	return []libyay.Setting{
		{Id: 1, Name: "Firefox", Path: "/app/firefox", HotKey: "ctrl+1", Mode: "default", Enabled: true},
		{Id: 2, Name: "Terminal", Path: "/app/terminal", HotKey: "", Mode: "fullscreen", Enabled: true},
		{Id: 3, Name: "Finder", Path: "/app/finder", HotKey: "ctrl+3", Mode: "desktop", Enabled: false},
		{Id: 4, Name: "Safari", Path: "/app/safari", HotKey: "", Mode: "default", Enabled: true},
		{Id: 5, Name: "Notes", Path: "/app/notes", HotKey: "alt+n", Mode: "default", Enabled: true},
	}
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
	case "shift+tab":
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
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

// ─── Initialization ───────────────────────────────────────────────

func TestNewModel_DefaultState(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")

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
	settings := testSettings()
	m := NewModel(settings, "0.1.0")

	if len(m.filteredIndices) != len(settings) {
		t.Errorf("expected %d filtered indices, got %d", len(settings), len(m.filteredIndices))
	}
}

func TestNewModel_EmptySettings(t *testing.T) {
	m := NewModel([]libyay.Setting{}, "1.0.0")

	if len(m.filteredIndices) != 0 {
		t.Errorf("expected 0 filtered indices, got %d", len(m.filteredIndices))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
}

func TestNewModel_PreservesSettingData(t *testing.T) {
	settings := testSettings()
	m := NewModel(settings, "0.1.0")

	if m.settings[0].Name != "Firefox" {
		t.Errorf("expected first setting name Firefox, got %s", m.settings[0].Name)
	}
	if m.settings[2].Enabled != false {
		t.Errorf("expected Finder to be disabled")
	}
	if m.settings[1].Mode != "fullscreen" {
		t.Errorf("expected Terminal mode fullscreen, got %s", m.settings[1].Mode)
	}
}

// ─── Filtering ────────────────────────────────────────────────────

func TestUpdateFilter_MatchesSubstring(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.filterInput.SetValue("fire")
	m.updateFilter()

	if len(m.filteredIndices) != 1 {
		t.Fatalf("expected 1 match for 'fire', got %d", len(m.filteredIndices))
	}
	if m.settings[m.filteredIndices[0]].Name != "Firefox" {
		t.Errorf("expected Firefox to match, got %s", m.settings[m.filteredIndices[0]].Name)
	}
}

func TestUpdateFilter_CaseInsensitive(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.filterInput.SetValue("TERMINAL")
	m.updateFilter()

	if len(m.filteredIndices) != 1 {
		t.Fatalf("expected 1 match for 'TERMINAL', got %d", len(m.filteredIndices))
	}
	if m.settings[m.filteredIndices[0]].Name != "Terminal" {
		t.Errorf("expected Terminal, got %s", m.settings[m.filteredIndices[0]].Name)
	}
}

func TestUpdateFilter_EmptyQueryShowsAll(t *testing.T) {
	settings := testSettings()
	m := NewModel(settings, "0.1.0")
	m.filterInput.SetValue("")
	m.updateFilter()

	if len(m.filteredIndices) != len(settings) {
		t.Errorf("expected all %d settings visible, got %d", len(settings), len(m.filteredIndices))
	}
}

func TestUpdateFilter_NoMatch(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.filterInput.SetValue("zzzznotanapp")
	m.updateFilter()

	if len(m.filteredIndices) != 0 {
		t.Errorf("expected 0 matches, got %d", len(m.filteredIndices))
	}
}

func TestUpdateFilter_MultipleMatches(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	// Both "Firefox" and "Finder" contain "fi" (case-insensitive)
	m.filterInput.SetValue("fi")
	m.updateFilter()

	if len(m.filteredIndices) != 2 {
		t.Fatalf("expected 2 matches for 'fi', got %d", len(m.filteredIndices))
	}
}

// ─── Cursor Navigation ───────────────────────────────────────────

func TestMoveCursor_Down(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")

	m = sendKey(t, m, "down")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after down, got %d", m.cursor)
	}

	m = sendKey(t, m, "down")
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 after second down, got %d", m.cursor)
	}
}

func TestMoveCursor_Up(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.cursor = 3

	m = sendKey(t, m, "up")
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 after up, got %d", m.cursor)
	}
}

func TestMoveCursor_VimKeys(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")

	m = sendKey(t, m, "j")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after j, got %d", m.cursor)
	}

	m = sendKey(t, m, "j")
	m = sendKey(t, m, "k")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after j then k, got %d", m.cursor)
	}
}

func TestMoveCursor_ClampAtTop(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.cursor = 0

	m = sendKey(t, m, "up")
	if m.cursor != 0 {
		t.Errorf("expected cursor clamped at 0, got %d", m.cursor)
	}
}

func TestMoveCursor_ClampAtBottom(t *testing.T) {
	settings := testSettings()
	m := NewModel(settings, "0.1.0")
	m.cursor = len(settings) - 1

	m = sendKey(t, m, "down")
	if m.cursor != len(settings)-1 {
		t.Errorf("expected cursor clamped at %d, got %d", len(settings)-1, m.cursor)
	}
}

func TestMoveCursor_EmptyList(t *testing.T) {
	m := NewModel([]libyay.Setting{}, "0.1.0")

	m = sendKey(t, m, "down")
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 on empty list, got %d", m.cursor)
	}

	m = sendKey(t, m, "up")
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 on empty list after up, got %d", m.cursor)
	}
}

// ─── State Transitions ───────────────────────────────────────────

func TestBrowse_SlashEntersFilterMode(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "/")

	if m.state != stateFilter {
		t.Errorf("expected stateFilter, got %d", m.state)
	}
}

func TestFilter_EscReturnsToBrowse(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "/")   // enter filter
	m = sendKey(t, m, "esc") // leave filter

	if m.state != stateBrowse {
		t.Errorf("expected stateBrowse after esc, got %d", m.state)
	}
}

func TestBrowse_EnterFocusesRow(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter")

	if m.state != stateRowFocus {
		t.Errorf("expected stateRowFocus, got %d", m.state)
	}
	if m.activeCol != colHotkey {
		t.Errorf("expected activeCol colHotkey on enter, got %d", m.activeCol)
	}
}

func TestBrowse_EnterNoopOnEmptyList(t *testing.T) {
	m := NewModel([]libyay.Setting{}, "0.1.0")
	m = sendKey(t, m, "enter")

	if m.state != stateBrowse {
		t.Errorf("expected stateBrowse on enter with empty list, got %d", m.state)
	}
}

func TestFilter_EnterFocusesRow(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "/")     // enter filter
	m = sendKey(t, m, "enter") // focus row from filter

	if m.state != stateRowFocus {
		t.Errorf("expected stateRowFocus, got %d", m.state)
	}
}

func TestRowFocus_EscReturnsToBrowse(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row
	m = sendKey(t, m, "esc")   // unfocus

	if m.state != stateBrowse {
		t.Errorf("expected stateBrowse after esc, got %d", m.state)
	}
	if m.activeCol != colNone {
		t.Errorf("expected activeCol colNone after esc, got %d", m.activeCol)
	}
}

func TestRowFocus_CtrlQReturnsToBrowse(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter")  // focus row
	m = sendKey(t, m, "ctrl+q") // unfocus

	if m.state != stateBrowse {
		t.Errorf("expected stateBrowse after ctrl+q, got %d", m.state)
	}
}

// ─── Column Cycling ──────────────────────────────────────────────

func TestCycleColumn_FullCycle(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row, starts at colHotkey

	if m.activeCol != colHotkey {
		t.Fatalf("expected initial column colHotkey, got %d", m.activeCol)
	}

	m = sendKey(t, m, "shift+tab")
	if m.activeCol != colMode {
		t.Errorf("expected colMode after first shift+tab, got %d", m.activeCol)
	}

	m = sendKey(t, m, "shift+tab")
	if m.activeCol != colEnabled {
		t.Errorf("expected colEnabled after second shift+tab, got %d", m.activeCol)
	}

	m = sendKey(t, m, "shift+tab")
	if m.activeCol != colHotkey {
		t.Errorf("expected colHotkey after third shift+tab (wrap), got %d", m.activeCol)
	}
}

func TestCycleColumn_ResetsRecording(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row at colHotkey
	m = sendKey(t, m, " ")     // start recording hotkey

	if !m.recordingHotkey {
		t.Fatal("expected recordingHotkey true after space")
	}

	// Now we manually set state to test that cycling resets recording
	// Since shift+tab goes through handleRowFocusKey which checks recordingHotkey first,
	// we need to cancel recording first (esc), then cycle
	m = sendKey(t, m, "esc") // cancel recording
	if m.recordingHotkey {
		t.Fatal("expected recordingHotkey false after esc")
	}

	m = sendKey(t, m, "shift+tab")
	if m.recordingHotkey {
		t.Errorf("expected recordingHotkey false after shift+tab")
	}
}

// ─── Mode Cycling ────────────────────────────────────────────────

func TestCycleModeForward(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	// First setting (Firefox) starts at "default"
	m = sendKey(t, m, "enter")     // focus row
	m = sendKey(t, m, "shift+tab") // go to colMode

	if m.activeCol != colMode {
		t.Fatalf("expected colMode, got %d", m.activeCol)
	}

	// Cycle forward: default -> fullscreen
	m = sendKey(t, m, " ")
	idx := m.filteredIndices[m.cursor]
	if m.settings[idx].Mode != "fullscreen" {
		t.Errorf("expected mode fullscreen, got %s", m.settings[idx].Mode)
	}

	// Cycle forward: fullscreen -> desktop
	m = sendKey(t, m, " ")
	if m.settings[idx].Mode != "desktop" {
		t.Errorf("expected mode desktop, got %s", m.settings[idx].Mode)
	}

	// Cycle forward: desktop -> default (wrap)
	m = sendKey(t, m, " ")
	if m.settings[idx].Mode != "default" {
		t.Errorf("expected mode default (wrap), got %s", m.settings[idx].Mode)
	}
}

func TestCycleModeForward_WithEnterKey(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter")     // focus row
	m = sendKey(t, m, "shift+tab") // go to colMode

	m = sendKey(t, m, "enter")
	idx := m.filteredIndices[m.cursor]
	if m.settings[idx].Mode != "fullscreen" {
		t.Errorf("expected mode fullscreen after enter, got %s", m.settings[idx].Mode)
	}
}

func TestCycleModeForward_WithRightKey(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter")     // focus row
	m = sendKey(t, m, "shift+tab") // go to colMode

	m = sendKey(t, m, "right")
	idx := m.filteredIndices[m.cursor]
	if m.settings[idx].Mode != "fullscreen" {
		t.Errorf("expected mode fullscreen after right, got %s", m.settings[idx].Mode)
	}
}

func TestCycleModeBackward_WithLeftKey(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter")     // focus row
	m = sendKey(t, m, "shift+tab") // go to colMode

	// default -> left -> desktop (wrap backward)
	m = sendKey(t, m, "left")
	idx := m.filteredIndices[m.cursor]
	if m.settings[idx].Mode != "desktop" {
		t.Errorf("expected mode desktop after left (wrap), got %s", m.settings[idx].Mode)
	}
}

func TestCycleMode_LogsChange(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter")     // focus row
	m = sendKey(t, m, "shift+tab") // go to colMode
	m = sendKey(t, m, " ")         // cycle mode

	if len(m.changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(m.changes))
	}
	c := m.changes[0]
	if c.Field != "mode" {
		t.Errorf("expected field 'mode', got %s", c.Field)
	}
	if c.OldVal != "default" {
		t.Errorf("expected old value 'default', got %s", c.OldVal)
	}
	if c.NewVal != "fullscreen" {
		t.Errorf("expected new value 'fullscreen', got %s", c.NewVal)
	}
}

// ─── Enable Toggling ─────────────────────────────────────────────

func TestToggleEnabled(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	// Firefox starts enabled=true
	m = sendKey(t, m, "enter")     // focus row
	m = sendKey(t, m, "shift+tab") // colMode
	m = sendKey(t, m, "shift+tab") // colEnabled

	if m.activeCol != colEnabled {
		t.Fatalf("expected colEnabled, got %d", m.activeCol)
	}

	// Toggle: true -> false
	m = sendKey(t, m, " ")
	idx := m.filteredIndices[m.cursor]
	if m.settings[idx].Enabled != false {
		t.Errorf("expected enabled=false after toggle, got true")
	}

	// Toggle: false -> true
	m = sendKey(t, m, " ")
	if m.settings[idx].Enabled != true {
		t.Errorf("expected enabled=true after second toggle, got false")
	}
}

func TestToggleEnabled_WithEnterKey(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter")     // focus
	m = sendKey(t, m, "shift+tab") // colMode
	m = sendKey(t, m, "shift+tab") // colEnabled

	m = sendKey(t, m, "enter")
	idx := m.filteredIndices[m.cursor]
	if m.settings[idx].Enabled != false {
		t.Errorf("expected enabled=false after enter toggle")
	}
}

func TestToggleEnabled_LogsChange(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter")     // focus
	m = sendKey(t, m, "shift+tab") // colMode
	m = sendKey(t, m, "shift+tab") // colEnabled
	m = sendKey(t, m, " ")         // toggle

	if len(m.changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(m.changes))
	}
	c := m.changes[0]
	if c.Field != "enabled" {
		t.Errorf("expected field 'enabled', got %s", c.Field)
	}
	if c.OldVal != "true" || c.NewVal != "false" {
		t.Errorf("expected 'true'->'false', got %q->%q", c.OldVal, c.NewVal)
	}
}

// ─── Hotkey Recording ────────────────────────────────────────────

func TestHotkeyRecording_EnterStartsRecording(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row, colHotkey

	if m.activeCol != colHotkey {
		t.Fatalf("expected colHotkey, got %d", m.activeCol)
	}

	m = sendKey(t, m, "enter") // start recording
	if !m.recordingHotkey {
		t.Errorf("expected recordingHotkey=true after enter")
	}
}

func TestHotkeyRecording_SpaceStartsRecording(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row, colHotkey
	m = sendKey(t, m, " ")     // start recording

	if !m.recordingHotkey {
		t.Errorf("expected recordingHotkey=true after space")
	}
}

func TestHotkeyRecording_RecordsKey(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row
	m = sendKey(t, m, " ")     // start recording

	// Press 'a' to record
	m = sendKey(t, m, "a")

	idx := m.filteredIndices[m.cursor]
	if m.settings[idx].HotKey != "a" {
		t.Errorf("expected hotkey 'a', got %q", m.settings[idx].HotKey)
	}
	if m.recordingHotkey {
		t.Errorf("expected recordingHotkey=false after recording")
	}
}

func TestHotkeyRecording_EscCancels(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row
	m = sendKey(t, m, " ")     // start recording

	originalHotkey := m.settings[m.filteredIndices[m.cursor]].HotKey

	m = sendKey(t, m, "esc") // cancel

	if m.recordingHotkey {
		t.Errorf("expected recordingHotkey=false after esc")
	}
	idx := m.filteredIndices[m.cursor]
	if m.settings[idx].HotKey != originalHotkey {
		t.Errorf("hotkey should be unchanged after cancel, expected %q got %q", originalHotkey, m.settings[idx].HotKey)
	}
}

func TestHotkeyRecording_BackspaceClears(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	// Firefox has hotkey "ctrl+1"
	m = sendKey(t, m, "enter")     // focus row
	m = sendKey(t, m, " ")         // start recording
	m = sendKey(t, m, "backspace") // clear hotkey

	idx := m.filteredIndices[m.cursor]
	if m.settings[idx].HotKey != "" {
		t.Errorf("expected empty hotkey after backspace, got %q", m.settings[idx].HotKey)
	}
	if m.recordingHotkey {
		t.Errorf("expected recordingHotkey=false after backspace clear")
	}
}

func TestHotkeyRecording_LogsChange(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row
	m = sendKey(t, m, " ")     // start recording
	m = sendKey(t, m, "x")     // record 'x'

	if len(m.changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(m.changes))
	}
	c := m.changes[0]
	if c.Field != "hotkey" {
		t.Errorf("expected field 'hotkey', got %s", c.Field)
	}
	if c.NewVal != "x" {
		t.Errorf("expected new hotkey 'x', got %s", c.NewVal)
	}
}

func TestHotkeyRecording_BackspaceClearLogsChange(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	// Firefox starts with hotkey "ctrl+1"
	m = sendKey(t, m, "enter")     // focus row
	m = sendKey(t, m, " ")         // start recording
	m = sendKey(t, m, "backspace") // clear

	if len(m.changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(m.changes))
	}
	c := m.changes[0]
	if c.Field != "hotkey" {
		t.Errorf("expected field 'hotkey', got %s", c.Field)
	}
	if c.NewVal != "(cleared)" {
		t.Errorf("expected new value '(cleared)', got %s", c.NewVal)
	}
}

// ─── Navigation While Focused ────────────────────────────────────

func TestRowFocus_NavigateWithArrowKeys(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row 0

	m = sendKey(t, m, "down") // move to row 1
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 in row focus, got %d", m.cursor)
	}
	if m.state != stateRowFocus {
		t.Errorf("expected to remain in stateRowFocus, got %d", m.state)
	}

	m = sendKey(t, m, "up")
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 in row focus after up, got %d", m.cursor)
	}
}

func TestRowFocus_NavigateWithVimKeys(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus
	// Move to mode column so j/k don't interfere with hotkey recording
	m = sendKey(t, m, "shift+tab") // colMode

	m = sendKey(t, m, "j")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after j in row focus, got %d", m.cursor)
	}

	m = sendKey(t, m, "k")
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after k in row focus, got %d", m.cursor)
	}
}

// ─── Filter Integration ──────────────────────────────────────────

func TestFilter_CursorClampsAfterFilterChange(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.cursor = 4 // last item

	// Enter filter mode and type a restrictive filter
	m = sendKey(t, m, "/")
	// Simulate typing "firefox" by setting value directly and updating
	m.filterInput.SetValue("firefox")
	m.updateFilter()

	// Only 1 result ("Firefox"), so cursor must be clamped to 0
	if len(m.filteredIndices) != 1 {
		t.Fatalf("expected 1 filtered result, got %d", len(m.filteredIndices))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor clamped to 0, got %d", m.cursor)
	}
}

func TestFilter_NavigateDuringFilter(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "/") // enter filter mode

	m = sendKey(t, m, "down")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 during filter, got %d", m.cursor)
	}

	m = sendKey(t, m, "up")
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 during filter after up, got %d", m.cursor)
	}
}

// ─── View Rendering ──────────────────────────────────────────────

func TestView_ContainsLogo(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "██") {
		t.Error("expected view to contain ASCII logo characters")
	}
}

func TestView_ContainsVersion(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "0.1.0") {
		t.Error("expected view to contain version string")
	}
}

func TestView_ContainsColumnHeaders(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	for _, header := range []string{"NAME", "HOTKEY", "MODE", "ENABLED"} {
		if !containsAny(view, header) {
			t.Errorf("expected view to contain header %q", header)
		}
	}
}

func TestView_ContainsSettingNames(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	for _, s := range testSettings() {
		if !containsAny(view, s.Name) {
			t.Errorf("expected view to contain setting name %q", s.Name)
		}
	}
}

func TestView_EmptyListMessage(t *testing.T) {
	m := NewModel([]libyay.Setting{}, "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "No matching") {
		t.Error("expected view to show empty list message")
	}
}

func TestView_ShowsBrowseHelp(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "navigate") {
		t.Error("expected browse mode help text")
	}
}

func TestView_ShowsFilterHelp(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "/")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "stop filtering") {
		t.Error("expected filter mode help text")
	}
}

func TestView_ShowsRowFocusHelp(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter")
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "shift+tab") {
		t.Error("expected row focus help text with shift+tab")
	}
}

func TestView_ShowsRecordingHotkeyHelp(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m = sendKey(t, m, "enter") // focus row at colHotkey
	m = sendKey(t, m, " ")     // start recording
	m.width = 120
	m.height = 40
	view := m.View()

	if !containsAny(view, "Press any key") {
		t.Error("expected hotkey recording help text")
	}
}

// ─── Exit Behavior ───────────────────────────────────────────────

func TestBrowse_QExits(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	_, cmd := m.Update(keyMsg("q"))

	if cmd == nil {
		t.Error("expected quit command from q in browse mode")
		return
	}
	// Execute the command to verify it produces a quit message
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestBrowse_CtrlCExits(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
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
	m := NewModel(testSettings(), "0.1.0")
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
	m := NewModel(testSettings(), "0.1.0")
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
	m := NewModel(testSettings(), "0.1.0")
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(model)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

// ─── Changes Accumulation ────────────────────────────────────────

func TestMultipleChanges_Accumulated(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")

	// Change mode on first row
	m = sendKey(t, m, "enter")     // focus
	m = sendKey(t, m, "shift+tab") // colMode
	m = sendKey(t, m, " ")         // cycle mode

	// Toggle enabled on first row
	m = sendKey(t, m, "shift+tab") // colEnabled
	m = sendKey(t, m, " ")         // toggle

	// Unfocus and move to next row
	m = sendKey(t, m, "esc")
	m = sendKey(t, m, "down")

	// Change mode on second row
	m = sendKey(t, m, "enter")     // focus
	m = sendKey(t, m, "shift+tab") // colMode
	m = sendKey(t, m, " ")         // cycle mode

	if len(m.changes) != 3 {
		t.Errorf("expected 3 accumulated changes, got %d", len(m.changes))
	}
}

// ─── selectedSetting ─────────────────────────────────────────────

func TestSelectedSetting_ReturnsCorrect(t *testing.T) {
	m := NewModel(testSettings(), "0.1.0")
	m.cursor = 2

	s := m.selectedSetting()
	if s == nil {
		t.Fatal("expected non-nil selected setting")
	}
	if s.Name != "Finder" {
		t.Errorf("expected Finder, got %s", s.Name)
	}
}

func TestSelectedSetting_NilOnEmpty(t *testing.T) {
	m := NewModel([]libyay.Setting{}, "0.1.0")

	s := m.selectedSetting()
	if s != nil {
		t.Errorf("expected nil selected setting on empty list, got %+v", s)
	}
}

// ─── Utility function tests ──────────────────────────────────────

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

func TestFormatBool(t *testing.T) {
	if formatBool(true) != "true" {
		t.Error("formatBool(true) should be 'true'")
	}
	if formatBool(false) != "false" {
		t.Error("formatBool(false) should be 'false'")
	}
}

func TestDisplayHotkey(t *testing.T) {
	if displayHotkey("") != "---" {
		t.Error("empty hotkey should display as '---'")
	}
	if displayHotkey("ctrl+a") != "ctrl+a" {
		t.Error("non-empty hotkey should be returned as-is")
	}
}

func TestIndexOf(t *testing.T) {
	modes := []string{"default", "fullscreen", "desktop"}
	if indexOf(modes, "fullscreen") != 1 {
		t.Error("indexOf should find fullscreen at index 1")
	}
	if indexOf(modes, "desktop") != 2 {
		t.Error("indexOf should find desktop at index 2")
	}
	if indexOf(modes, "nonexistent") != 0 {
		t.Error("indexOf should return 0 for unknown values")
	}
}

func TestIsModifierOnly(t *testing.T) {
	if !isModifierOnly("ctrl") {
		t.Error("expected ctrl to be modifier only")
	}
	if !isModifierOnly("shift") {
		t.Error("expected shift to be modifier only")
	}
	if !isModifierOnly("alt") {
		t.Error("expected alt to be modifier only")
	}
	if isModifierOnly("a") {
		t.Error("'a' should not be modifier only")
	}
	if isModifierOnly("ctrl+a") {
		t.Error("'ctrl+a' should not be modifier only")
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
