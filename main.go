package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"jira_cli/internal/git_utils"
	"jira_cli/internal/gui"
	"jira_cli/internal/jira"
	"jira_cli/internal/utils"

	"github.com/charmbracelet/lipgloss"
)

type model struct {
	table     table.Model
	isLoading bool
	spinner   spinner.Model
	err       error
	width     int
	height    int
	view      string
	input     textinput.Model
	tickets   []jira.JiraTicketsMsg
}

type errMsg error

func returnChoices(m *model) tea.Cmd {
	return func() tea.Msg {
		j, err := jira.GetJiraTickets()
		if err != nil {
			return errMsg(err)
		}

		newChoices := []jira.JiraTicketsMsg{}
		for _, issue := range j.Issues {
			newChoices = append(newChoices, jira.JiraTicketsMsg{
				Key:     issue.Key,
				Type:    issue.Fields.IssueType.Name,
				Summary: issue.Fields.Summary,
				Status:  issue.Fields.Status.Name,
				Created: issue.Fields.Created,
			})
		}

		return newChoices
	}
}

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		textinput.Blink,
		returnChoices(&m),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.isLoading && m.err == nil {
			m.updateTableSize()
		}

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.view {
		case "list":
			switch msg.String() {
			case "enter":
				if m.view == "list" && len(m.tickets) > 0 {
					selectedRow := m.table.Cursor()
					if selectedRow < len(m.tickets) {
						selectedTicket := m.tickets[selectedRow]
						selected_branch := git_utils.FormatBranchName(selectedTicket)
						m.view = "input"
						m.input = gui.CreateBranchInput(selected_branch, m.width)
						return m, nil
					}
				}
			}

		case "input":
			switch msg.String() {
			case "enter":
				branchName := m.input.Value()
				return m, git_utils.CheckoutBranch(branchName)
			case "esc":
				m.view = "list"
				return m, nil
			}

			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

	case []jira.JiraTicketsMsg:
		m.tickets = msg

		columns := []table.Column{
			{Title: "Key", Width: 0},
			{Title: "Type", Width: 0},
			{Title: "Summary", Width: 0},
			{Title: "Status", Width: 0},
			{Title: "Created", Width: 0},
		}

		rows := []table.Row{}
		for _, ticket := range msg {
			rows = append(rows, table.Row{ticket.Key, ticket.Type, ticket.Summary, ticket.Status, utils.FormatRelativeTime(ticket.Created)})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			BorderBottom(true).
			Bold(false)
		s.Selected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("57")).
			Background(lipgloss.Color("0")).
			Bold(false)

		t.SetStyles(s)

		m.table = t
		m.isLoading = false
		m.updateTableSize()
		return m, nil

	case errMsg:
		m.err = msg
		m.isLoading = false
		return m, nil

	case spinner.TickMsg:
		if m.isLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if !m.isLoading && m.view == "list" {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.isLoading {
		_, authURL := jira.GetAuthUrlAndConfig()

		line2 := lipgloss.NewStyle().Width(m.width / 2).
			Foreground(lipgloss.Color("7")).
			Faint(true).
			Render(fmt.Sprintf(
				"or open the following URL in your browser: \n\n%s",
				authURL))

		line1 := fmt.Sprintf("%s Opening browser for Atlassian authorization...",
			m.spinner.View())

		return fmt.Sprintf("\n%s\n\n%s",
			line1,
			line2)
	}

	if m.err != nil {
		errorText := fmt.Sprintf("Error loading data: %v", m.err)
		helpText := "Press 'r' to retry or 'q' to quit"

		if m.width > 0 {
			errorPadding := max(0, (m.width-len(errorText))/2)
			helpPadding := max(0, (m.width-len(helpText))/2)
			errorText = strings.Repeat(" ", errorPadding) + errorText
			helpText = strings.Repeat(" ", helpPadding) + helpText
		}

		return fmt.Sprintf("\n%s\n\n%s\n", errorText, helpText)
	}

	if m.isLoading {
		loadingText := fmt.Sprintf("%s %s", m.spinner.View(), "Loading Jira tickets...")

		return fmt.Sprintf("\n%s\n\nPress q to quit", loadingText)
	}

	if m.view == "input" {
		faintStyle := lipgloss.NewStyle().Faint(true)
		return fmt.Sprintf(
			"%s\n\nenter %s %s esc %s %s q/ctrl+c %s",
			m.input.View(),
			faintStyle.Render("checkout branch"),
			faintStyle.Render("•"),
			faintStyle.Render("go back"),
			faintStyle.Render("•"),
			faintStyle.Render("quit"),
		)
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Render(m.table.View())
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found. Make sure you have set the required environment variables.")
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

	m := model{
		table:     table.New(),
		spinner:   s,
		isLoading: true,
		view:      "list",
		input:     textinput.New(),
		tickets:   []jira.JiraTicketsMsg{},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
