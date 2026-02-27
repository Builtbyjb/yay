package tui

import (
	"fmt"
	"strings"

	lib "github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// model is the Bubble Tea model for the YAY TUI
type model struct {
	state           focusState
	settings        []lib.Setting
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

// NewModel creates and initialises a new TUI model
func NewModel(settings []lib.Setting, version string) model {
	ti := textinput.New()
	ti.Placeholder = "Type to Search..."
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
		switch m.state {
		case stateBrowse:
			return m.HandleBrowseKey(msg)
		case stateFilter:
			return m.SearchUpdate(msg)
		case stateRowFocus:
			return m.handleRowFocusKey(msg)
		}
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	contents := []string{}

	contents = append(contents, m.HeaderView())
	contents = append(contents, m.SearchView())
	contents = append(contents, m.TableView())
	contents = append(contents, m.StatusLineView())
	contents = append(contents, m.HelpView())

	return lipgloss.JoinVertical(lipgloss.Left, contents...)
}

func (m model) HandleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
			m.activeCol = colKey
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

func (m model) handleRowFocusKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If recording a hotkey, capture the key
	if m.recordingHotkey {
		return m.handleHotkeyRecording(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.state = stateBrowse
		m.activeCol = colNone
		m.recordingHotkey = false
		m.pendingMods = ""
		return m, nil

	case "tab":
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
	case colKey:
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
	hotkey := msg.String()
	if hotkey == "" {
		return m, nil
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
		m.activeCol = colKey
	case colKey:
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
func (m *model) selectedSetting() *lib.Setting {
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

// Starts the TUI
func Run(settings []lib.Setting, version string) ([]changeEntry, error) {
	m := NewModel(settings, version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := finalModel.(model)
	return fm.changes, nil
}
