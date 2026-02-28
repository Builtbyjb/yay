package tui

import "github.com/charmbracelet/lipgloss"

var (
	LogoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PRIMARY_COLOR))

	VersionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PRIMARY_COLOR)).
			Bold(true)

	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(SECONDARY_COLOR))

	// Filter input label
	FilterLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(PRIMARY_COLOR)).
				Bold(true)

	// Normal row
	NormalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PRIMARY_COLOR))

	// Cursor row (highlighted in browse/filter mode)
	CursorRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PRIMARY_COLOR)).
			Background(lipgloss.Color(PRIMARY_ACCENT_COLOR))

	// Focused row (in row-focus mode)
	FocusedRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PRIMARY_COLOR)).
			Background(lipgloss.Color(SECONDARY_ACCENT_COLOR))

	// Active column cell (the column currently being edited)
	ActiveCellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PRIMARY_COLOR)).
			Background(lipgloss.Color(ACTIVE_COLOR)).
			Bold(true)

	// Help bar at the bottom
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PRIMARY_COLOR))

	// Status indicator style
	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(PRIMARY_COLOR)).
			Bold(true)
)
