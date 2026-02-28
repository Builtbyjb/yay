package tui

import (
	"fmt"

	lib "github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	state           focusState
	settings        []ModelSetting
	filteredIndices []int
	filterInput     textinput.Model
	cursor          int // position within filteredIndices
	activeCol       columnID
	version         string
	width           int
	height          int
	changes         []changeEntry
	recordingHotkey bool // true when waiting for the next key press for hotkey
}

func NewModel(settings []ModelSetting, version string) model {
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

// selectedSetting returns the currently selected setting, or nil if none.
func (m *model) selectedSetting() *ModelSetting {
	if len(m.filteredIndices) == 0 || m.cursor >= len(m.filteredIndices) {
		return nil
	}
	idx := m.filteredIndices[m.cursor]
	return &m.settings[idx]
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
	modelSettings := mapToModelSetting(settings)
	m := NewModel(modelSettings, version)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm := finalModel.(model)
	return fm.changes, nil
}
