package tui

import (
	"github.com/Builtbyjb/yay/pkg/lib"
	"github.com/Builtbyjb/yay/pkg/lib/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

// Starts the TUI
func Run(db *core.Database, settings []core.Setting, version string) error {
	m := NewModel(db, settings, version)
	p := tea.NewProgram(m, tea.WithAltScreen())

	go lib.KeyEventListener(db, func(event lib.KeyEvent) {
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
