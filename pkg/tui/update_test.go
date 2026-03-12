package tui

import (
	"testing"

	"github.com/Builtbyjb/yay/pkg/lib/core"
)

func TestUpdateFilter_MatchesSubstring(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.searchInput.SetValue("fire")
	m.updateFilter()

	if len(m.searchedIndices) != 1 {
		t.Fatalf("expected 1 match for 'fire', got %d", len(m.searchedIndices))
	}
	if m.settings[m.searchedIndices[0]].Name != "Firefox" {
		t.Errorf("expected Firefox to match, got %s", m.settings[m.searchedIndices[0]].Name)
	}
}

func TestUpdateFilter_CaseInsensitive(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.searchInput.SetValue("TERMINAL")
	m.updateFilter()

	if len(m.searchedIndices) != 1 {
		t.Fatalf("expected 1 match for 'TERMINAL', got %d", len(m.searchedIndices))
	}
	if m.settings[m.searchedIndices[0]].Name != "Terminal" {
		t.Errorf("expected Terminal, got %s", m.settings[m.searchedIndices[0]].Name)
	}
}

func TestUpdateFilter_EmptyQueryShowsAll(t *testing.T) {
	database := setupTestDatabase(t)
	settings := testSettings(t, database)
	m := NewModel(nil, settings, "0.1.0")
	m.searchInput.SetValue("")
	m.updateFilter()

	if len(m.searchedIndices) != len(settings) {
		t.Errorf("expected all %d settings visible, got %d", len(settings), len(m.searchedIndices))
	}
}

func TestUpdateFilter_NoMatch(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.searchInput.SetValue("zzzznotanapp")
	m.updateFilter()

	if len(m.searchedIndices) != 0 {
		t.Errorf("expected 0 matches, got %d", len(m.searchedIndices))
	}
}

func TestUpdateFilter_MultipleMatches(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	// Both "Firefox" and "Finder" contain "fi" (case-insensitive)
	m.searchInput.SetValue("fi")
	m.updateFilter()

	if len(m.searchedIndices) != 2 {
		t.Fatalf("expected 2 matches for 'fi', got %d", len(m.searchedIndices))
	}
}

func TestMoveCursor_Down(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")

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
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.cursor = 3

	m = sendKey(t, m, "up")
	if m.cursor != 2 {
		t.Errorf("expected cursor 2 after up, got %d", m.cursor)
	}
}

func TestMoveCursor_VimKeys(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")

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
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.cursor = 0

	m = sendKey(t, m, "up")
	if m.cursor != 0 {
		t.Errorf("expected cursor clamped at 0, got %d", m.cursor)
	}
}

func TestMoveCursor_ClampAtBottom(t *testing.T) {
	database := setupTestDatabase(t)
	settings := testSettings(t, database)
	m := NewModel(nil, settings, "0.1.0")
	m.cursor = len(settings) - 1

	m = sendKey(t, m, "down")
	if m.cursor != len(settings)-1 {
		t.Errorf("expected cursor clamped at %d, got %d", len(settings)-1, m.cursor)
	}
}

func TestMoveCursor_EmptyList(t *testing.T) {
	m := NewModel(nil, []core.Setting{}, "0.1.0")

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
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "/")

	if m.state != stateFilter {
		t.Errorf("expected stateFilter, got %d", m.state)
	}
}

func TestFilter_EscReturnsToBrowse(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "/")   // enter filter
	m = sendKey(t, m, "esc") // leave filter

	if m.state != stateBrowse {
		t.Errorf("expected stateBrowse after esc, got %d", m.state)
	}
}

func TestBrowse_EnterFocusesRow(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter")

	if m.state != stateRowFocus {
		t.Errorf("expected stateRowFocus, got %d", m.state)
	}
	if m.activeCol != colKey {
		t.Errorf("expected activeCol colKey(1) on enter, got %d", m.activeCol)
	}
}

func TestBrowse_EnterNoopOnEmptyList(t *testing.T) {
	m := NewModel(nil, []core.Setting{}, "0.1.0")
	m = sendKey(t, m, "enter")

	if m.state != stateBrowse {
		t.Errorf("expected stateBrowse on enter with empty list, got %d", m.state)
	}
}

func TestFilter_EnterFocusesRow(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "/")     // enter filter
	m = sendKey(t, m, "enter") // focus row from filter

	if m.state != stateRowFocus {
		t.Errorf("expected stateRowFocus, got %d", m.state)
	}
}

func TestRowFocus_EscReturnsToBrowse(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter") // focus row
	m = sendKey(t, m, "esc")   // un-focus

	if m.state != stateBrowse {
		t.Errorf("expected stateBrowse after esc, got %d", m.state)
	}
	if m.activeCol != colNone {
		t.Errorf("expected activeCol colNone after esc, got %d", m.activeCol)
	}
}

func TestRowFocus_CtrlQReturnsToBrowse(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter")    // focus row
	m = sendKey(t, m, CANCEL_KEY) // un-focus

	if m.state != stateBrowse {
		t.Errorf("expected stateBrowse after ctrl+q, got %d", m.state)
	}
}

// ─── Column Cycling ──────────────────────────────────────────────

func TestCycleColumn_FullCycle(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter") // focus row, starts at colHotkey

	if m.activeCol != colKey {
		t.Errorf("expected colKey(1) first %s, got %d", SWITCH_COLUMN_KEY, m.activeCol)
	}

	m = sendKey(t, m, SWITCH_COLUMN_KEY)
	if m.activeCol != colMode {
		t.Errorf("expected colMode(2) %s, got %d", SWITCH_COLUMN_KEY, m.activeCol)
	}

	m = sendKey(t, m, SWITCH_COLUMN_KEY)
	if m.activeCol != colEnabled {
		t.Errorf("expected colEnabled(3) %s (wrap), got %d", SWITCH_COLUMN_KEY, m.activeCol)
	}

	// Full cycle
	m = sendKey(t, m, SWITCH_COLUMN_KEY)
	if m.activeCol != colKey {
		t.Errorf("expected colKey(1) after full cycle %s (wrap), got %d", SWITCH_COLUMN_KEY, m.activeCol)
	}
}

func TestCycleColumn_ResetsRecording(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter") // focus row at colKey
	m = sendKey(t, m, "enter") // Enter to start recording hotkey

	if !m.recordingHotkey {
		t.Fatal("expected recordingHotkey true after space")
	}

	// Now we manually set state to test that cycling resets recording
	// Since SWITCH_COLUMN_KEY goes through handleRowFocusKey which checks recordingHotkey first,
	// we need to cancel recording first (esc), then cycle
	m = sendKey(t, m, "esc") // cancel recording
	if m.recordingHotkey {
		t.Fatal("expected recordingHotkey false after esc")
	}

	m = sendKey(t, m, SWITCH_COLUMN_KEY) // cycle column
	if m.recordingHotkey {
		t.Errorf("expected recordingHotkey false after %s", SWITCH_COLUMN_KEY)
	}
}

// ─── Mode Cycling ────────────────────────────────────────────────

func TestCycleMode(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(database, testSettings(t, database), "0.1.0")
	// First setting (Finder) starts at "desktop"
	m = sendKey(t, m, "enter")           // focus row
	m = sendKey(t, m, SWITCH_COLUMN_KEY) // go to colMode

	if m.activeCol != colMode {
		t.Fatalf("expected colMode, got %d", m.activeCol)
	}

	idx := m.searchedIndices[m.cursor]

	// Cycle forward: desktop -> default
	m = sendKey(t, m, " ")
	if m.settings[idx].Mode != "default" {
		t.Errorf("expected mode default, got %s", m.settings[idx].Mode)
	}

	// Cycle forward: default -> desktop (wrap)
	m = sendKey(t, m, " ")
	if m.settings[idx].Mode != "desktop" {
		t.Errorf("expected mode desktop (wrap), got %s", m.settings[idx].Mode)
	}
}

func TestCycleMode_WithEnterKey(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(database, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter")           // focus row
	m = sendKey(t, m, SWITCH_COLUMN_KEY) // go to colMode

	m = sendKey(t, m, "enter")
	idx := m.searchedIndices[m.cursor]
	if m.settings[idx].Mode != "default" {
		t.Errorf("expected mode default after enter, got %s", m.settings[idx].Mode)
	}
}

// ─── Enable Toggling ─────────────────────────────────────────────

func TestToggleEnabled(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(database, testSettings(t, database), "0.1.0")
	// Finder starts enabled=false
	m = sendKey(t, m, "enter")           // focus row
	m = sendKey(t, m, SWITCH_COLUMN_KEY) // colMode
	m = sendKey(t, m, SWITCH_COLUMN_KEY) // colEnabled

	if m.activeCol != colEnabled {
		t.Fatalf("expected colEnabled, got %d", m.activeCol)
	}

	// Toggle: false -> true
	m = sendKey(t, m, " ")
	idx := m.searchedIndices[m.cursor]
	if m.settings[idx].Enabled != true {
		t.Errorf("expected enabled=true after toggle, got false")
	}

	// Toggle: true -> false
	m = sendKey(t, m, " ")
	if m.settings[idx].Enabled != false {
		t.Errorf("expected enabled=false after second toggle, got true")
	}
}

func TestToggleEnabled_WithEnterKey(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(database, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter")           // focus
	m = sendKey(t, m, SWITCH_COLUMN_KEY) // colMode
	m = sendKey(t, m, SWITCH_COLUMN_KEY) // colEnabled

	m = sendKey(t, m, "enter")
	idx := m.searchedIndices[m.cursor]
	if m.settings[idx].Enabled != true {
		t.Errorf("expected enabled=true after enter toggle")
	}
}

// ─── Hotkey Recording ────────────────────────────────────────────

func TestHotkeyRecording_EnterStartsRecording(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(database, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter") // focus row,

	if m.activeCol != colKey {
		t.Fatalf("expected colHotkey, got %d", m.activeCol)
	}

	m = sendKey(t, m, "enter") // start recording
	if !m.recordingHotkey {
		t.Errorf("expected recordingHotkey=true after enter")
	}
}

func TestHotkeyRecording_SpaceStartsRecording(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(database, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter") // focus row

	m = sendKey(t, m, " ") // start recording
	if !m.recordingHotkey {
		t.Errorf("expected recordingHotkey=true after space")
	}
}

// func TestHotkeyRecording_RecordsKey(t *testing.T) {
// 	database := setupTestDatabase(t)
// 	m := NewModel(database, testSettings(t, database), "0.1.0")
// 	m = sendKey(t, m, "enter") // focus row
// 	m = sendKey(t, m, " ")     // start recording

// 	// Press 'a' to record
// 	m = sendKey(t, m, "a")

// 	idx := m.searchedIndices[m.cursor]
// 	if m.settings[idx].HotKey != (sql.NullString{String: "a", Valid: true}) {
// 		t.Errorf("expected hotkey 'a', got %q", m.settings[idx].HotKey.String)
// 	}
// 	if m.recordingHotkey {
// 		t.Errorf("expected recordingHotkey=false after recording")
// 	}
// }

// func TestHotkeyRecording_EscCancels(t *testing.T) {
// 	database := setupTestDatabase(t)
// 	m := NewModel(nil, testSettings(t, database), "0.1.0")
// 	m = sendKey(t, m, "enter") // focus row
// 	m = sendKey(t, m, " ")     // start recording

// 	originalHotkey := m.settings[m.searchedIndices[m.cursor]].HotKey

// 	m = sendKey(t, m, CANCEL_KEY) // cancel

// 	if m.recordingHotkey {
// 		t.Errorf("expected recordingHotkey=false after esc")
// 	}
// 	idx := m.searchedIndices[m.cursor]
// 	if m.settings[idx].HotKey != originalHotkey {
// 		t.Errorf("hotkey should be unchanged after cancel, expected %q got %q", originalHotkey.String, m.settings[idx].HotKey.String)
// 	}
// }

// func TestHotkeyRecording_BackspaceClears(t *testing.T) {
// 	database := setupTestDatabase(t)
// 	m := NewModel(nil, testSettings(t, database), "0.1.0")
// 	// Firefox has hotkey "ctrl+1"
// 	m = sendKey(t, m, "enter")     // focus row
// 	m = sendKey(t, m, " ")         // start recording
// 	m = sendKey(t, m, "backspace") // clear hotkey

// 	idx := m.searchedIndices[m.cursor]
// 	if m.settings[idx].HotKey != (sql.NullString{String: "", Valid: false}) {
// 		t.Errorf("expected empty hotkey after backspace, got %q", m.settings[idx].HotKey.String)
// 	}
// 	if m.recordingHotkey {
// 		t.Errorf("expected recordingHotkey=false after backspace clear")
// 	}
// }

// ─── Navigation While Focused ────────────────────────────────────

func TestRowFocus_NavigateWithArrowKeys(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
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
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m = sendKey(t, m, "enter") // focus
	// Move to mode column so j/k don't interfere with hotkey recording
	m = sendKey(t, m, SWITCH_COLUMN_KEY) // colMode

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
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
	m.cursor = 4 // last item

	// Enter filter mode and type a restrictive filter
	m = sendKey(t, m, "/")
	// Simulate typing "firefox" by setting value directly and updating
	m.searchInput.SetValue("firefox")
	m.updateFilter()

	// Only 1 result ("Firefox"), so cursor must be clamped to 0
	if len(m.searchedIndices) != 1 {
		t.Fatalf("expected 1 filtered result, got %d", len(m.searchedIndices))
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor clamped to 0, got %d", m.cursor)
	}
}

func TestFilter_NavigateDuringFilter(t *testing.T) {
	database := setupTestDatabase(t)
	m := NewModel(nil, testSettings(t, database), "0.1.0")
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
