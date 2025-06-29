package git_utils

import (
	"fmt"
	"os/exec"
	"strings"

	"jira_cli/internal/jira"

	tea "github.com/charmbracelet/bubbletea"
)

type errMsg error

func CheckoutBranch(ticket jira.JiraTicketsMsg) tea.Cmd {
	return func() tea.Msg {
		branchName := strings.ToLower("feature/" + ticket.Key + "-" + ticket.Summary)
		branchName = strings.ReplaceAll(branchName, " ", "_")

		maxLength := 80
		if len(branchName) > maxLength {
			truncated := branchName[:maxLength]
			if lastUnderscore := strings.LastIndex(truncated, "_"); lastUnderscore != -1 {
				branchName = truncated[:lastUnderscore] + "-"
			} else {
				branchName = truncated + "-"
			}
		}

		fmt.Println("branchName", branchName)

		checkCmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
		branchExists := checkCmd.Run() == nil

		var cmd *exec.Cmd

		if branchExists {
			cmd = exec.Command("git", "checkout", branchName)
		} else {
			cmd = exec.Command("git", "checkout", "-b", branchName)
		}

		err := cmd.Run()

		if err != nil {
			return errMsg(fmt.Errorf("failed to checkout branch %s: %v", branchName, err))
		}

		return tea.Quit()
	}
}
