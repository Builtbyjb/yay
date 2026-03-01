package tui

import (
	lib "github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	state           focusState
	settings        []ModelSetting
	searchedIndices []int
	searchInput     textinput.Model
	cursor          int // position within filteredIndices
	activeCol       columnID
	version         string
	width           int
	height          int
	changes         []int // Stores indices of settings that have been changed but not yet saved
	recordingHotkey bool  // true when waiting for the next key press for hotkey
}

func NewModel(settings []ModelSetting, version string) model {
	ti := textinput.New()
	ti.Placeholder = "Type to Search..."
	ti.CharLimit = 64
	ti.Width = 40

	m := model{
		state:       stateBrowse,
		settings:    settings,
		searchInput: ti,
		cursor:      0,
		activeCol:   colNone,
		version:     version,
		changes:     []int{},
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
			// Check if there is any values in the changes list
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
	if len(m.searchedIndices) == 0 || m.cursor >= len(m.searchedIndices) {
		return nil
	}
	idx := m.searchedIndices[m.cursor]
	return &m.settings[idx]
}

// Starts the TUI
func Run(settings []lib.Setting, version string) error {
	modelSettings := mapToModelSetting(settings)
	m := NewModel(modelSettings, version)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err := p.Run()
	if err != nil {
		return err
	}

	return nil
}
