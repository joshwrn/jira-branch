package gui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type HelpItem struct {
	Key  string
	Desc string
}

func CreateHelpItems(items []HelpItem) string {
	b := strings.Builder{}

	for index, item := range items {
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Render(item.Key))
		b.WriteString(" ")
		b.WriteString(FaintWhiteText.Render(item.Desc))
		if index != len(items)-1 {
			b.WriteString(FaintWhiteText.Render(" • "))
		}
	}

	return lipgloss.NewStyle().
		PaddingLeft(1).
		Render(b.String())
}
