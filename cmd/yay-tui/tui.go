package main

import (
	"fmt"
	"strings"

	"github.com/Builtbyjb/yay/pkg/libyay"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
	colHotkey           // listening for hotkey input
	colMode             // cycling through modes
	colEnabled          // toggling true/false
)

// changeEntry records a setting change for logging after exit
type changeEntry struct {
	Name   string
	Field  string
	OldVal string
	NewVal string
}

// model is the Bubble Tea model for the YAY TUI
type model struct {
	state           focusState
	settings        []libyay.Setting
	filteredIndices []int
	filterInput     textinput.Model
	cursor          int // position within filteredIndices
	activeCol       columnID
	version         string
	width           int
	height          int
	changes         []changeEntry
	recordingHotkey bool   // true when waiting for the next key press for hotkey
	pendingMods     string // accumulated modifier keys for hotkey recording
}

// ----- Styles (shades of blue) -----

var (
	// Deep navy for the logo
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00b4d8")).
			Bold(true)

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5dade2"))

	// Filter input label
	filterLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#5dade2")).
				Bold(true)

	// Column header style
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a8d8ea")).
			Bold(true).
			Underline(true)

	// Normal row
	normalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cce5ff"))

	// Cursor row (highlighted in browse/filter mode)
	cursorRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#0f3460"))

	// Focused row (in row-focus mode)
	focusedRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color("#1a3a5c"))

	// Active column cell (the column currently being edited)
	activeCellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0a1128")).
			Background(lipgloss.Color("#00b4d8")).
			Bold(true)

	// Help bar at the bottom
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5b8fb9"))

	// Dimmed text
	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2c3e6b"))

	// Status indicator style
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00b4d8")).
			Bold(true)
)

// Column widths
const (
	colWidthName    = 28
	colWidthHotkey  = 18
	colWidthMode    = 14
	colWidthEnabled = 10
)

// ASCII logo for YAY
const asciiLogo = `
██    ██  █████  ██    ██
 ██  ██  ██   ██  ██  ██
  ████   ███████   ████
   ██    ██   ██    ██
   ██    ██   ██    ██`

// NewModel creates and initialises a new TUI model
func NewModel(settings []libyay.Setting, version string) model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 64
	ti.Width = 40

	m := model{
		state:       stateBrowse,
		settings:    settings,
		filterInput: ti,
		cursor:      0,
		activeCol:   colNone,
		version:     version,
		changes:     []changeEntry{},
	}
	m.updateFilter()
	return m
}

// RunTUI starts the TUI program. Call this from main.
func RunTUI(settings []libyay.Setting, version string) ([]changeEntry, error) {
	m := NewModel(settings, version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := finalModel.(model)
	return fm.changes, nil
}

// ----- Bubble Tea interface -----

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	// Logo + version
	b.WriteString(logoStyle.Render(asciiLogo))
	b.WriteString("\n")
	b.WriteString(versionStyle.Render("  v" + m.version))
	b.WriteString("\n\n")

	// Filter input
	filterPrefix := filterLabelStyle.Render("  Filter: ")
	if m.state == stateFilter {
		b.WriteString(filterPrefix + m.filterInput.View())
	} else {
		// Show filter text but not focused
		filterText := m.filterInput.Value()
		if filterText == "" {
			filterText = dimStyle.Render("(press / to filter)")
		} else {
			filterText = normalRowStyle.Render(filterText)
		}
		b.WriteString(filterPrefix + filterText)
	}
	b.WriteString("\n\n")

	// Column headers
	b.WriteString("  ")
	b.WriteString(headerStyle.Render(padRight("NAME", colWidthName)))
	b.WriteString(headerStyle.Render(padRight("HOTKEY", colWidthHotkey)))
	b.WriteString(headerStyle.Render(padRight("MODE", colWidthMode)))
	b.WriteString(headerStyle.Render(padRight("ENABLED", colWidthEnabled)))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", colWidthName+colWidthHotkey+colWidthMode+colWidthEnabled)))
	b.WriteString("\n")

	// Determine how many rows we can show
	// Reserve lines: logo(~6) + version(1) + blank(1) + filter(1) + blank(1) + header(1) + separator(1) + blank(1) + help(~3) = ~16
	maxRows := m.height - 16
	if maxRows < 3 {
		maxRows = 3
	}

	// Calculate scroll window
	startIdx := 0
	if len(m.filteredIndices) > maxRows {
		if m.cursor >= maxRows {
			startIdx = m.cursor - maxRows + 1
		}
		if startIdx+maxRows > len(m.filteredIndices) {
			startIdx = len(m.filteredIndices) - maxRows
		}
	}

	endIdx := startIdx + maxRows
	if endIdx > len(m.filteredIndices) {
		endIdx = len(m.filteredIndices)
	}

	if len(m.filteredIndices) == 0 {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  No matching applications."))
		b.WriteString("\n")
	} else {
		for i := startIdx; i < endIdx; i++ {
			idx := m.filteredIndices[i]
			s := m.settings[idx]
			isCursor := i == m.cursor
			isRowFocused := isCursor && m.state == stateRowFocus

			prefix := "  "
			if isCursor {
				prefix = "> "
			}

			nameCell := padRight(truncate(s.Name, colWidthName-1), colWidthName)
			hotkeyCell := padRight(displayHotkey(s.HotKey), colWidthHotkey)
			modeCell := padRight(s.Mode, colWidthMode)
			enabledCell := padRight(formatBool(s.Enabled), colWidthEnabled)

			if isRowFocused {
				// Render each cell; highlight the active column
				nameCell = focusedRowStyle.Render(nameCell)
				if m.activeCol == colHotkey {
					if m.recordingHotkey {
						displayText := "recording..."
						if m.pendingMods != "" {
							displayText = m.pendingMods + "+..."
						}
						hotkeyCell = activeCellStyle.Render(padRight(displayText, colWidthHotkey))
					} else {
						hotkeyCell = activeCellStyle.Render(padRight(displayHotkey(s.HotKey), colWidthHotkey))
					}
				} else {
					hotkeyCell = focusedRowStyle.Render(hotkeyCell)
				}
				if m.activeCol == colMode {
					modeCell = activeCellStyle.Render(padRight(s.Mode, colWidthMode))
				} else {
					modeCell = focusedRowStyle.Render(modeCell)
				}
				if m.activeCol == colEnabled {
					enabledCell = activeCellStyle.Render(padRight(formatBool(s.Enabled), colWidthEnabled))
				} else {
					enabledCell = focusedRowStyle.Render(enabledCell)
				}

				b.WriteString(focusedRowStyle.Render(prefix) + nameCell + hotkeyCell + modeCell + enabledCell)
			} else if isCursor {
				row := prefix + nameCell + hotkeyCell + modeCell + enabledCell
				b.WriteString(cursorRowStyle.Render(row))
			} else {
				row := prefix + nameCell + hotkeyCell + modeCell + enabledCell
				b.WriteString(normalRowStyle.Render(row))
			}
			b.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(m.filteredIndices) > maxRows {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  showing %d-%d of %d", startIdx+1, endIdx, len(m.filteredIndices))))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Status line
	switch m.state {
	case stateRowFocus:
		colName := "none"
		switch m.activeCol {
		case colHotkey:
			if m.recordingHotkey {
				colName = "hotkey (recording)"
			} else {
				colName = "hotkey"
			}
		case colMode:
			colName = "mode"
		case colEnabled:
			colName = "enabled"
		}
		b.WriteString(statusStyle.Render(fmt.Sprintf("  EDITING ROW  |  column: %s", colName)))
		b.WriteString("\n")
	case stateFilter:
		b.WriteString(statusStyle.Render("  FILTER MODE"))
		b.WriteString("\n")
	default:
		b.WriteString(statusStyle.Render("  BROWSE"))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	switch m.state {
	case stateBrowse:
		b.WriteString(helpStyle.Render("  ↑/↓/j/k: navigate  enter: edit row  /: filter  q/ctrl+c: quit"))
	case stateFilter:
		b.WriteString(helpStyle.Render("  ↑/↓: navigate  enter: edit row  esc: stop filtering  ctrl+c: quit"))
	case stateRowFocus:
		switch m.activeCol {
		case colHotkey:
			if m.recordingHotkey {
				b.WriteString(helpStyle.Render("  Press any key to set hotkey  backspace: clear  esc: cancel"))
			} else {
				b.WriteString(helpStyle.Render("  shift+tab: next column  enter: record hotkey  esc/ctrl+q: unfocus"))
			}
		case colMode:
			b.WriteString(helpStyle.Render("  space/enter/←/→: cycle mode  shift+tab: next column  esc/ctrl+q: unfocus"))
		case colEnabled:
			b.WriteString(helpStyle.Render("  space/enter: toggle  shift+tab: next column  esc/ctrl+q: unfocus"))
		default:
			b.WriteString(helpStyle.Render("  shift+tab: select column  esc/ctrl+q: unfocus"))
		}
	}

	return b.String()
}

// ----- Key handling -----

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateBrowse:
		return m.handleBrowseKey(msg)
	case stateFilter:
		return m.handleFilterKey(msg)
	case stateRowFocus:
		return m.handleRowFocusKey(msg)
	}
	return m, nil
}

func (m model) handleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		m.moveCursor(-1)
		return m, nil

	case "down", "j":
		m.moveCursor(1)
		return m, nil

	case "home", "g":
		m.cursor = 0
		return m, nil

	case "end", "G":
		if len(m.filteredIndices) > 0 {
			m.cursor = len(m.filteredIndices) - 1
		}
		return m, nil

	case "enter":
		if len(m.filteredIndices) > 0 {
			m.state = stateRowFocus
			m.activeCol = colHotkey
			m.recordingHotkey = false
			m.pendingMods = ""
		}
		return m, nil

	case "/":
		m.state = stateFilter
		m.filterInput.Focus()
		return m, nil
	}

	return m, nil
}

func (m model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.state = stateBrowse
		m.filterInput.Blur()
		return m, nil

	case "up":
		m.moveCursor(-1)
		return m, nil

	case "down":
		m.moveCursor(1)
		return m, nil

	case "enter":
		if len(m.filteredIndices) > 0 {
			m.state = stateRowFocus
			m.activeCol = colHotkey
			m.recordingHotkey = false
			m.pendingMods = ""
			m.filterInput.Blur()
		}
		return m, nil
	}

	// Pass key to text input for filtering
	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.updateFilter()
	// Ensure cursor is in bounds after filter changes
	if m.cursor >= len(m.filteredIndices) {
		if len(m.filteredIndices) > 0 {
			m.cursor = len(m.filteredIndices) - 1
		} else {
			m.cursor = 0
		}
	}
	return m, cmd
}

func (m model) handleRowFocusKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If recording a hotkey, capture the key
	if m.recordingHotkey {
		return m.handleHotkeyRecording(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "ctrl+q":
		m.state = stateBrowse
		m.activeCol = colNone
		m.recordingHotkey = false
		m.pendingMods = ""
		return m, nil

	case "shift+tab":
		m.cycleColumn()
		return m, nil

	case "up", "k":
		m.moveCursor(-1)
		return m, nil

	case "down", "j":
		m.moveCursor(1)
		return m, nil
	}

	// Column-specific actions
	switch m.activeCol {
	case colHotkey:
		if msg.String() == "enter" || msg.String() == " " {
			m.recordingHotkey = true
			m.pendingMods = ""
			return m, nil
		}

	case colMode:
		switch msg.String() {
		case "enter", " ", "right", "l":
			m.cycleModeForward()
			return m, nil
		case "left", "h":
			m.cycleModeBackward()
			return m, nil
		}

	case colEnabled:
		if msg.String() == "enter" || msg.String() == " " {
			m.toggleEnabled()
			return m, nil
		}
	}

	return m, nil
}

func (m model) handleHotkeyRecording(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Esc cancels recording
	if key == "esc" {
		m.recordingHotkey = false
		m.pendingMods = ""
		return m, nil
	}

	// Backspace clears the hotkey
	if key == "backspace" {
		if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
			idx := m.filteredIndices[m.cursor]
			old := m.settings[idx].HotKey
			m.settings[idx].HotKey = ""
			m.recordingHotkey = false
			m.pendingMods = ""
			if old != "" {
				m.changes = append(m.changes, changeEntry{
					Name:   m.settings[idx].Name,
					Field:  "hotkey",
					OldVal: old,
					NewVal: "(cleared)",
				})
			}
		}
		return m, nil
	}

	// Build the hotkey string from the tea.KeyMsg
	hotkey := buildHotkeyString(msg)
	if hotkey == "" {
		return m, nil
	}

	// Check if this is a standalone modifier key press - accumulate it
	if isModifierOnly(hotkey) {
		if m.pendingMods == "" {
			m.pendingMods = hotkey
		} else {
			m.pendingMods = m.pendingMods + "+" + hotkey
		}
		return m, nil
	}

	// If we have pending mods, prepend them
	if m.pendingMods != "" {
		hotkey = m.pendingMods + "+" + hotkey
		m.pendingMods = ""
	}

	if len(m.filteredIndices) > 0 && m.cursor < len(m.filteredIndices) {
		idx := m.filteredIndices[m.cursor]
		old := m.settings[idx].HotKey
		m.settings[idx].HotKey = hotkey
		m.recordingHotkey = false
		m.changes = append(m.changes, changeEntry{
			Name:   m.settings[idx].Name,
			Field:  "hotkey",
			OldVal: displayHotkey(old),
			NewVal: hotkey,
		})
	}
	return m, nil
}

// ----- Model helpers -----

func (m *model) moveCursor(delta int) {
	if len(m.filteredIndices) == 0 {
		m.cursor = 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.filteredIndices) {
		m.cursor = len(m.filteredIndices) - 1
	}
}

func (m *model) updateFilter() {
	query := strings.ToLower(m.filterInput.Value())
	m.filteredIndices = make([]int, 0, len(m.settings))
	for i, s := range m.settings {
		if query == "" || strings.Contains(strings.ToLower(s.Name), query) {
			m.filteredIndices = append(m.filteredIndices, i)
		}
	}
	// Clamp cursor to stay within the new filtered list
	if len(m.filteredIndices) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.filteredIndices) {
		m.cursor = len(m.filteredIndices) - 1
	}
}

func (m *model) cycleColumn() {
	switch m.activeCol {
	case colNone, colEnabled:
		m.activeCol = colHotkey
	case colHotkey:
		m.activeCol = colMode
	case colMode:
		m.activeCol = colEnabled
	}
	m.recordingHotkey = false
	m.pendingMods = ""
}

func (m *model) cycleModeForward() {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return
	}
	idx := m.filteredIndices[m.cursor]
	old := m.settings[idx].Mode
	current := indexOf(AvailableModes, old)
	next := (current + 1) % len(AvailableModes)
	m.settings[idx].Mode = AvailableModes[next]
	m.changes = append(m.changes, changeEntry{
		Name:   m.settings[idx].Name,
		Field:  "mode",
		OldVal: old,
		NewVal: m.settings[idx].Mode,
	})
}

func (m *model) cycleModeBackward() {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return
	}
	idx := m.filteredIndices[m.cursor]
	old := m.settings[idx].Mode
	current := indexOf(AvailableModes, old)
	next := current - 1
	if next < 0 {
		next = len(AvailableModes) - 1
	}
	m.settings[idx].Mode = AvailableModes[next]
	m.changes = append(m.changes, changeEntry{
		Name:   m.settings[idx].Name,
		Field:  "mode",
		OldVal: old,
		NewVal: m.settings[idx].Mode,
	})
}

func (m *model) toggleEnabled() {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return
	}
	idx := m.filteredIndices[m.cursor]
	old := m.settings[idx].Enabled
	m.settings[idx].Enabled = !old
	m.changes = append(m.changes, changeEntry{
		Name:   m.settings[idx].Name,
		Field:  "enabled",
		OldVal: formatBool(old),
		NewVal: formatBool(m.settings[idx].Enabled),
	})
}

// selectedSetting returns the currently selected setting, or nil if none.
func (m *model) selectedSetting() *libyay.Setting {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return nil
	}
	idx := m.filteredIndices[m.cursor]
	return &m.settings[idx]
}

// ----- Utility functions -----

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return s + strings.Repeat(" ", width-len(s))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func formatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func displayHotkey(h string) string {
	if h == "" {
		return "---"
	}
	return h
}

func indexOf(slice []string, val string) int {
	for i, s := range slice {
		if s == val {
			return i
		}
	}
	return 0
}

// buildHotkeyString converts a tea.KeyMsg into a human-readable hotkey string
// using the format <modifier>+<key>
func buildHotkeyString(msg tea.KeyMsg) string {
	raw := msg.String()
	if raw == "" {
		return ""
	}
	return raw
}

// isModifierOnly returns true if the key string is just a modifier key name
func isModifierOnly(key string) bool {
	switch strings.ToLower(key) {
	case "ctrl", "alt", "shift", "super", "meta":
		return true
	}
	return false
}

// PrintChanges outputs all recorded changes to stdout
func PrintChanges(changes []changeEntry) {
	if len(changes) == 0 {
		fmt.Println("No changes were made.")
		return
	}
	fmt.Println("\n--- YAY Settings Changes ---")
	for i, c := range changes {
		fmt.Printf("  %d. [%s] %s: %q -> %q\n", i+1, c.Name, c.Field, c.OldVal, c.NewVal)
	}
	fmt.Printf("Total changes: %d\n", len(changes))
}
