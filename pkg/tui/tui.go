package tui

import (
	"fmt"

	"github.com/Builtbyjb/yay/pkg/lib"
	"github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	hook "github.com/robotn/gohook"
)

type model struct {
	state           focusState
	settings        []ModelSetting
	searchedIndices []int
	searchInput     textinput.Model
	cursor          int // position within SearchedIndices
	activeCol       columnID
	version         string
	width           int
	height          int
	changes         []int // Stores indices of settings that have been changed but not yet saved
	debug           []int
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
		searchInput: ti,
		cursor:      0,
		activeCol:   colNone,
		version:     version,
		changes:     []int{},
		debug:       []int{},
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
	case lib.CustomKeyMsg:
		if m.recordingHotkey && msg.Event.Kind == hook.KeyDown {
			m.debug = append(m.debug, int(msg.Event.Rawcode))
		}
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

// Starts the TUI
func Run(settings []core.Setting, version string) error {
	modelSettings := mapToModelSetting(settings)
	m := NewModel(modelSettings, version)
	p := tea.NewProgram(m, tea.WithAltScreen())

	go func() {
		eventChan := hook.Start()
		defer hook.End()

		for event := range eventChan {
			p.Send(lib.CustomKeyMsg{Event: event})
		}
	}()

	fModel, err := p.Run()
	if err != nil {
		return err
	}

	fmt.Println(fModel.(model).debug)

	return nil
}
