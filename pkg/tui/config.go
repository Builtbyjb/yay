package tui

// Available modes for the mode column
var AvailableModes = []string{"default", "fullscreen", "desktop"}

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

// changeEntry records a setting change for logging after exit
type changeEntry struct {
	Name   string
	Field  string
	OldVal string
	NewVal string
}

// Column widths
const (
	colWidthName    = 28
	colWidthHotkey  = 18
	colWidthMode    = 14
	colWidthEnabled = 10
)
