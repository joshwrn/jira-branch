package app

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshwrn/jira-branch/internal/gui"
)

func createSearchInput(width int) textinput.Model {
	ti := textinput.New()
	ti.CharLimit = 200
	ti.Width = width
	ti.PlaceholderStyle = gui.FaintWhiteText
	ti.Placeholder = "Search for a ticket"
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	return ti
}
