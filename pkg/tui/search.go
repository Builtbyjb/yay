package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) SearchUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case EXIT_KEY:
		return m, tea.Quit

	case CANCEL_KEY:
		m.state = stateBrowse
		m.searchInput.Blur()
		return m, nil

	case "up":
		m.moveCursor(-1)
		return m, nil

	case "down":
		m.moveCursor(1)
		return m, nil

	case "enter":
		if len(m.searchedIndices) > 0 {
			m.state = stateRowFocus
			m.activeCol = colKey
			m.recordingHotkey = false
			m.searchInput.Blur()
		}
		return m, nil
	}

	// Pass key to text input for filtering
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.updateFilter()
	// Ensure cursor is in bounds after filter changes
	if m.cursor >= len(m.searchedIndices) {
		if len(m.searchedIndices) > 0 {
			m.cursor = len(m.searchedIndices) - 1
		} else {
			m.cursor = 0
		}
	}
	return m, cmd
}
func (m model) SearchView() string {

	contents := []string{}

	contents = append(contents, lipgloss.JoinVertical(
		lipgloss.Left,
		FilterLabelStyle.Render("Search: "),
	))

	if m.state == stateFilter {
		contents = append(contents, m.searchInput.View())
	} else {
		// Show filter text but not focused
		filterText := m.searchInput.Value()
		if filterText == "" {
			filterText = DimStyle.Render("(press / to Search)")
		} else {
			filterText = NormalRowStyle.Render(filterText)
		}

		contents = append(contents, filterText)
	}

	contents = append(contents, "\n")

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		contents...,
	)
}

func (m *model) updateFilter() {
	query := strings.ToLower(m.searchInput.Value())
	m.searchedIndices = make([]int, 0, len(m.settings))
	for i, s := range m.settings {
		if query == "" || strings.Contains(strings.ToLower(s.Name), query) {
			m.searchedIndices = append(m.searchedIndices, i)
		}
	}
	// Clamp cursor to stay within the new filtered list
	if len(m.searchedIndices) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.searchedIndices) {
		m.cursor = len(m.searchedIndices) - 1
	}
}
