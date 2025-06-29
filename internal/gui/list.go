package gui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func newCustomDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("5")).
		Bold(true)

	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color("5")).
		Faint(true)

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

func NewCustomList(items []list.Item) list.Model {
	delegate := newCustomDelegate()
	l := list.New(items, delegate, 0, 0)

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowTitle(false)
	return l
}
