package gui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func CreateBranchInput(selected_branch string, width int) textinput.Model {
	ti := textinput.New()
	ti.Prompt = "Confirm or edit branch name: \n\n"
	ti.SetValue(selected_branch)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = width

	return ti
}

var FaintWhiteText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("7")).
	Faint(true)
