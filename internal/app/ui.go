package app

import "github.com/charmbracelet/bubbles/table"

func (m *model) updateTableSize() {
	if m.width > 0 && m.height > 0 {
		keyWidth := 10
		typeWidth := 10
		statusWidth := 30
		createdWidth := 15
		summaryWidth := max(20, m.width-keyWidth-typeWidth-statusWidth-createdWidth-12)

		columns := []table.Column{
			{Title: "Key", Width: keyWidth},
			{Title: "Type", Width: typeWidth},
			{Title: "Summary", Width: summaryWidth},
			{Title: "Status", Width: statusWidth},
			{Title: "Created", Width: createdWidth},
		}
		m.table.SetColumns(columns)
		m.table.SetWidth(m.width - 2)
		m.table.SetHeight(m.height - 2)
	}
}
