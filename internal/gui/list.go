package gui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func NewCustomDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("5")).
		Bold(true)

	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color("7"))

	d.Styles.NormalTitle = d.Styles.NormalTitle.
		Foreground(lipgloss.Color("4")).
		Bold(true)

	d.Styles.NormalDesc = d.Styles.NormalDesc.
		Foreground(lipgloss.Color("7"))

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		BorderLeft(true).
		BorderLeftForeground(lipgloss.Color("5"))

	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		BorderLeft(true).
		BorderLeftForeground(lipgloss.Color("5"))

	d.SetSpacing(0)

	return d
}
