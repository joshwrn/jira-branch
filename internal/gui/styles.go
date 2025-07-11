package gui

import (
	"github.com/charmbracelet/lipgloss"
)

var FaintWhiteText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("7")).
	Faint(true)

var ErrorText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9")).
	Bold(true)
