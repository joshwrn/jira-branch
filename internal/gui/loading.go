package gui

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

type LoadingView struct {
	Spinner spinner.Model
	Text    string
	Width   int
	Height  int
}

func CreateLoadingView(m *LoadingView) string {
	b := strings.Builder{}
	bw := b.WriteString
	bw(m.Spinner.View())
	bw(" ")
	bw(m.Text)
	bw("\n\n")
	b.WriteString(CreateHelpItems([]HelpItem{
		{Key: "q/ctrl+c", Desc: "Quit"},
	}))
	return lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(b.String())
}
