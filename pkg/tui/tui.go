package tui

import (
	"github.com/Builtbyjb/yay/pkg/lib"
	"github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/Builtbyjb/yay/pkg/lib/darwin"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	db              *core.Database
	state           focusState
	settings        []core.Setting
	searchedIndices []int
	searchInput     textinput.Model
	cursor          int // position within SearchedIndices
	activeCol       columnID
	version         string
	width           int
	height          int
	keys            []uint16
	mod             string
	key             uint16
	recordingHotkey bool // true when waiting for the next key press for hotkey
	errors          []string
	debug           []int
}

func NewModel(db *core.Database, settings []core.Setting, version string) model {
	ti := textinput.New()
	ti.Placeholder = "Type to Search..."
	ti.CharLimit = 64
	ti.Width = 40

	m := model{
		db:          db,
		state:       stateBrowse,
		settings:    settings,
		searchInput: ti,
		cursor:      0,
		activeCol:   colNone,
		version:     version,
		keys:        []uint16{},
		errors:      []string{},
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
	case lib.CKeyMsg:
		return m.RecordKey(msg)
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
func Run(db *core.Database, settings []core.Setting, version string) error {
	m := NewModel(db, settings, version)
	p := tea.NewProgram(m, tea.WithAltScreen())

	go lib.Listener(db, func(event darwin.KeyEvent) {
		p.Send(lib.CKeyMsg{Event: event})
	})

	_, err := p.Run()
	if err != nil {
		return err
	}

	// fmt.Println(fModel.(model).errors)
	// fmt.Println(fModel.(model).debug)

	return nil
}
