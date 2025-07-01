package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/joshwrn/jira-branch/internal/git_utils"
	"github.com/joshwrn/jira-branch/internal/gui"
	"github.com/joshwrn/jira-branch/internal/jira"
	"github.com/joshwrn/jira-branch/internal/utils"

	"github.com/charmbracelet/lipgloss"
)

type model struct {
	table      table.Model
	isLoading  bool
	isLoggedIn bool
	spinner    spinner.Model
	err        error
	width      int
	height     int
	view       string
	input      textinput.Model
	tickets    []jira.JiraTicketsMsg

	// Credential input fields
	credentialInputs []textinput.Model
	currentField     int
	credentials      jira.Credentials
}

type errMsg error

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

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		textinput.Blink,
		func() tea.Msg {
			credentials, err := jira.LoadCredentials()
			if err != nil {
				return credentialsNeededMsg{}
			}
			if err := jira.ValidateCredentials(credentials); err != nil {
				return credentialsNeededMsg{}
			}
			return credentials
		},
	)
}

type credentialsNeededMsg struct{}

type ticketsMsg struct {
	tickets []jira.JiraTicketsMsg
	err     error
}

func fetchTickets(credentials jira.Credentials) tea.Cmd {
	return func() tea.Msg {
		tickets, err := jira.GetJiraTickets(credentials)
		return ticketsMsg{tickets: tickets, err: err}
	}
}

func validateAndStoreCredentials(credentials jira.Credentials) tea.Cmd {
	return func() tea.Msg {
		if err := jira.ValidateCredentials(credentials); err != nil {
			return errMsg(err)
		}

		if err := jira.StoreCredentials(credentials); err != nil {
			return errMsg(fmt.Errorf("failed to store credentials: %v", err))
		}

		return credentials
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.isLoading && m.err == nil && m.isLoggedIn {
			m.updateTableSize()
		}

	case tea.KeyMsg:
		if m.isLoggedIn && msg.String() == "q" {
			return m, nil
		}
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.view {
		case "credentials":
			switch msg.String() {
			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()

				if s == "enter" && m.currentField == len(m.credentialInputs)-1 {
					m.credentials = jira.Credentials{
						JiraURL:  strings.TrimSpace(m.credentialInputs[0].Value()),
						Email:    strings.TrimSpace(m.credentialInputs[1].Value()),
						APIToken: strings.TrimSpace(m.credentialInputs[2].Value()),
					}

					if m.credentials.JiraURL == "" || m.credentials.Email == "" || m.credentials.APIToken == "" {
						m.err = errMsg(fmt.Errorf("all fields are required"))
						return m, nil
					}

					m.isLoading = true
					m.err = nil
					return m, validateAndStoreCredentials(m.credentials)
				}

				if s == "up" || s == "shift+tab" {
					m.currentField--
				} else {
					m.currentField++
				}

				if m.currentField > len(m.credentialInputs)-1 {
					m.currentField = 0
				} else if m.currentField < 0 {
					m.currentField = len(m.credentialInputs) - 1
				}

				for i := 0; i < len(m.credentialInputs); i++ {
					if i == m.currentField {
						m.credentialInputs[i].Focus()
					} else {
						m.credentialInputs[i].Blur()
					}
				}

				return m, textinput.Blink
			}

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
						return m, textinput.Blink
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

	case credentialsNeededMsg:
		m.view = "credentials"
		m.credentialInputs = gui.CreateCredentialInputs(m.width)
		m.currentField = 0
		return m, textinput.Blink

	case jira.Credentials:
		m.credentials = msg
		m.isLoggedIn = true
		m.isLoading = true
		m.view = "list"
		return m, fetchTickets(msg)

	case ticketsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.isLoading = false
			return m, nil
		}

		m.tickets = msg.tickets

		columns := []table.Column{
			{Title: "Key", Width: 0},
			{Title: "Type", Width: 0},
			{Title: "Summary", Width: 0},
			{Title: "Status", Width: 0},
			{Title: "Created", Width: 0},
		}

		rows := []table.Row{}
		for _, ticket := range m.tickets {
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
			Foreground(lipgloss.Color("12")).
			Background(lipgloss.Color("0")).
			Bold(false)

		t.SetStyles(s)

		m.table = t
		m.isLoading = false
		m.isLoggedIn = true
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

	if !m.isLoading && m.isLoggedIn && m.view == "list" {
		m.table, cmd = m.table.Update(msg)
	} else if m.view == "credentials" && len(m.credentialInputs) > 0 {
		m.credentialInputs[m.currentField], cmd = m.credentialInputs[m.currentField].Update(msg)
	} else if m.view == "input" {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.view == "credentials" {
		var b strings.Builder

		b.WriteString("\n")
		b.WriteString(lipgloss.
			NewStyle().
			Foreground(lipgloss.Color("7")).
			Render("Generate an API token at: "))
		b.WriteString(
			lipgloss.
				NewStyle().
				Foreground(lipgloss.Color("4")).
				Underline(true).
				Render("https://id.atlassian.com/manage-profile/security/api-tokens"))

		b.WriteString("\n\n")

		b.WriteString(gui.FaintWhiteText.
			Render(`If you choose "API token with scopes" give it the "read:jira-work" scope.`))

		b.WriteString("\n\n")

		if m.err != nil {
			b.WriteString(gui.ErrorText.Render(fmt.Sprintf("❌ %v", m.err)))
			b.WriteString("\n\n")
		}

		b.WriteString(m.credentialInputs[0].View())
		b.WriteString("\n")
		b.WriteString(m.credentialInputs[1].View())
		b.WriteString("\n")
		b.WriteString(m.credentialInputs[2].View())
		b.WriteString("\n\n")

		b.WriteString(gui.CreateHelpItems([]gui.HelpItem{
			{Key: "tab", Desc: "Navigate"},
			{Key: "enter", Desc: "Submit"},
			{Key: "ctrl+c", Desc: "Quit"},
		}))

		return b.String()
	}

	if !m.isLoggedIn && m.isLoading {
		line1 := fmt.Sprintf(
			"%s Validating credentials...",
			m.spinner.View(),
		)

		line2 := gui.FaintWhiteText.
			Render("Press ctrl+c to quit")

		return fmt.Sprintf(
			"\n%s\n\n%s",
			line1,
			line2,
		)
	}

	if m.err != nil {
		errorText := fmt.Sprintf("Error: %v", m.err)
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
		faintStyle := gui.FaintWhiteText
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
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

	m := model{
		table:            table.New(),
		spinner:          s,
		isLoading:        true,
		isLoggedIn:       false,
		view:             "list",
		input:            textinput.New(),
		tickets:          []jira.JiraTicketsMsg{},
		credentialInputs: []textinput.Model{},
		currentField:     0,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
