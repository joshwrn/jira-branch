package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/joshwrn/jira-branch/internal/gui"
)

func viewList(m model) string {
	helper := gui.CreateHelpItems([]gui.HelpItem{
		{Key: "j/k", Desc: "↓/↑"},
		{Key: "enter", Desc: "Select ticket"},
		{Key: "/", Desc: "Search"},
		{Key: "r", Desc: "Refresh"},
		{Key: "S", Desc: "Sign out"},
		{Key: "q/ctrl+c", Desc: "Quit"},
	})

	searchView := strings.Builder{}
	bw := searchView.WriteString

	if m.showSearch {
		bw(m.searchInput.View())
		bw("\n")
		helper = gui.CreateHelpItems([]gui.HelpItem{
			{Key: "esc", Desc: "Clear"},
			{Key: "enter", Desc: "Confirm"},
		})
	}

	if m.search != "" && !m.showSearch {
		helperWidth := lipgloss.Width(helper)
		availableWidth := m.width - helperWidth
		searchText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Render(fmt.Sprintf("/%s", m.search))

		helper = helper +
			lipgloss.NewStyle().
				Width(availableWidth-1).
				Align(lipgloss.Right).
				Render(searchText)
	}

	return searchView.String() + lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Render(m.table.View()) + "\n" + helper
}
