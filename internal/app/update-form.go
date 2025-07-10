package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/joshwrn/jira-branch/internal/git_utils"
	"github.com/joshwrn/jira-branch/internal/jira"
)

func updateForm(m model, msg tea.Msg) (model, tea.Cmd) {
	if m.isSubmittingForm {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = "list"
			return m, nil
		}
	}
	form, formCmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State != huh.StateCompleted {
			return m, formCmd
		}
		m.isSubmittingForm = true
		return m, tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				if *m.formShouldMarkAsInProgress {
					err := jira.MarkAsInProgress(
						m.credentials,
						m.list.SelectedRow()[0],
					)
					if err != nil {
						return errMsg(err)
					}
				}
				checkCmd := git_utils.CheckoutBranch(*m.formBranchName)
				return checkCmd()
			},
		)
	}
	return m, nil
}
