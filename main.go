package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	git_utils "jira_cli/internal/git"
	"jira_cli/internal/gui"
	"jira_cli/internal/jira"

	"github.com/charmbracelet/lipgloss"
)

type item struct {
	key     string
	summary string
}

func (i item) FilterValue() string { return i.key + " " + i.summary }
func (i item) Title() string       { return i.key }
func (i item) Description() string { return i.summary }

type model struct {
	list      list.Model
	isLoading bool
	spinner   spinner.Model
	err       error
	width     int
	height    int
	view      string
	input     textinput.Model
}

type errMsg error

func returnChoices() tea.Cmd {
	return func() tea.Msg {
		j, err := jira.GetJiraTickets()
		if err != nil {
			return errMsg(err)
		}

		newChoices := []jira.JiraTicketsMsg{}
		for _, issue := range j.Issues {
			newChoices = append(newChoices, jira.JiraTicketsMsg{
				Key:     issue.Key,
				Summary: issue.Fields.Summary,
			})
		}

		return newChoices
	}
}

func (m *model) updateListSize() {
	if m.width > 0 && m.height > 0 {
		m.list.SetWidth(m.width)
		m.list.SetHeight(m.height)
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
		returnChoices(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.isLoading && m.err == nil {
			m.updateListSize()
		}

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.view {
		case "list":
			switch msg.String() {

			case "esc":
				if m.list.FilterState() == list.Filtering {
					m.list.SetFilterState(list.Unfiltered)
					m.list.SetFilterText("")
					return m, nil
				}
				return m, nil

			case "enter":
				if m.view == "list" {
					if m.list.FilterState() == list.Filtering {
						m.list.SetFilterState(list.FilterApplied)
						return m, nil
					}

					selectedItem := m.list.SelectedItem()
					if selectedItem != nil {
						if i, ok := selectedItem.(item); ok {
							selected_branch := git_utils.FormatBranchName(jira.JiraTicketsMsg{
								Key:     i.key,
								Summary: i.summary,
							})
							m.view = "input"

							ti := textinput.New()
							ti.Prompt = "Confirm branch name: \n\n"
							ti.SetValue(selected_branch)
							ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
							ti.Focus()
							ti.CharLimit = 200
							ti.Width = m.width
							m.input = ti

							return m, nil
						}
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
		items := []list.Item{}
		for _, choice := range msg {
			items = append(items, item{
				key:     choice.Key,
				summary: choice.Summary,
			})
		}
		delegate := gui.NewCustomDelegate()
		l := list.New(items, delegate, 0, 0)

		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(true)
		l.SetShowHelp(false)
		l.SetShowPagination(false)

		l.Title = strings.Repeat("─", 10) + " Select a ticket " + strings.Repeat("─", 10)

		l.Styles.Title = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

		l.Styles.TitleBar = lipgloss.NewStyle().
			Padding(0, 0)

		m.list = l
		m.isLoading = false
		m.updateListSize()
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

	if !m.isLoading {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.isLoading {
		loadingText := fmt.Sprintf("%s %s", m.spinner.View(), "Loading Jira tickets...")

		return fmt.Sprintf("\n%s\n\nPress q to quit", loadingText)
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

	if m.view == "input" {
		return m.input.View()
	}

	return m.list.View()
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found. Make sure you have set the required environment variables.")
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	emptyList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)

	m := model{
		list:      emptyList,
		spinner:   s,
		isLoading: true,
		view:      "list",
		input:     textinput.New(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
