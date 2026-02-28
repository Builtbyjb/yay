package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) HeaderView() string {
	var ascii_logo = `
██    ██  █████  ██    ██
 ██  ██  ██   ██  ██  ██
  ████   ███████   ████
   ██    ██   ██    ██
   ██    ██   ██    ██` + "	v" + m.version

	logo := lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Split(LogoStyle.Render(ascii_logo), "\n")...,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		logo,
		"\n",
	)
}
