package tui

var AvailableModes = []string{"default", "fullscreen", "desktop"}

var AvailableModifiersMacos = []string{"⇧ Shift", "⌥ Option", "⌃ Control", "⌘ Command"}
var AvailableModifiersWindows = []string{"⭡ Shift", "⎇ Alt", "⌃ Ctrl", "⊞ Win"}

// Focus states for the TUI
type focusState int

const (
	stateBrowse   focusState = iota // navigating the list, q exits
	stateFilter                     // typing in the filter input
	stateRowFocus                   // a row is focused for editing
)

// Column focus within a focused row
type columnID int

const (
	colNone    columnID = iota
	colMod              // cycle modifier keys
	colKey              // listening for hotkey input
	colMode             // cycle through modes
	colEnabled          // toggle true/false
)

const (
	colWidthName    = 28
	colWidthHotkey  = 18
	colWidthMode    = 14
	colWidthEnabled = 10
)

/* Colors */
const PRIMARY_COLOR = "#fffff"
const SECONDARY_COLOR = "#b6b8ba"
const PRIMARY_ACCENT_COLOR = "#1a3a5c"
const SECONDARY_ACCENT_COLOR = "#0f3460"
const ACTIVE_COLOR = "#00b4d8"

/* Keys */
const SWITCH_COLUMN_KEY = "tab"
const SEARCH_KEY = "/"
const CANCEL_KEY = "esc"
const EXIT_KEY = "ctrl+c"
