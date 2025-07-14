package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/joshwrn/jira-branch/internal/utils"
)

type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type TransitionIssueBody struct {
	Transition Transition `json:"transition"`
}

type TransitionResponse struct {
	Transitions []Transition `json:"transitions"`
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-issueidorkey-transitions-get
func GetInProgressTransition(issueKey string, credentials Credentials) (string, error) {
	client := newClient()
	resp, err := client.makeRequest(
		"GET",
		fmt.Sprintf("issue/%s/transitions", issueKey),
		nil,
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", fmt.Errorf("authentication failed: check your credentials")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("jira API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result TransitionResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	for _, transition := range result.Transitions {
		if transition.Name == "In Progress" {
			return transition.ID, nil
		}
	}

	return "", fmt.Errorf("transition not found")
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-issueidorkey-transitions-post
func MarkAsInProgress(credentials Credentials, issueKey string) error {
	transitionId, err := GetInProgressTransition(issueKey, credentials)
	if err != nil {
		return err
	}

	body, err := json.Marshal(TransitionIssueBody{
		Transition: Transition{ID: transitionId},
	})
	if err != nil {
		return err
	}

	client := newClient()
	resp, err := client.makeRequest(
		"POST",
		fmt.Sprintf("issue/%s/transitions", issueKey),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: check your credentials")
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("jira API error: %d", resp.StatusCode)
	}

	return nil
}

type JiraTicketsMsg struct {
	Key     string
	Summary string
	Type    string
	Status  string
	Created string
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/#api-rest-api-3-search-jql-get
func GetJiraTickets(credentials Credentials) ([]JiraTicketsMsg, error) {
	client := newClient()
	req, err := client.createRequest("GET", "search", nil)
	if err != nil {
		return []JiraTicketsMsg{}, err
	}

	config, err := utils.ReadConfigFile()
	if err != nil {
		utils.Log.Info().Err(err).Msg("Failed to read config file")
	}

	q := req.URL.Query()
	jql := ""
	if config.ProjectKey != "" {
		jql = fmt.Sprintf("project = %s AND ", config.ProjectKey)
	}
	jql = jql + "assignee = currentUser() AND status != Done order by createdDate"
	q.Add("jql", jql)
	q.Add("fields", "summary,status,issuetype,assignee,created")
	q.Add("maxResults", "100")
	req.URL.RawQuery = q.Encode()

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return []JiraTicketsMsg{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return []JiraTicketsMsg{}, fmt.Errorf("authentication failed: check your credentials")
	}

	if resp.StatusCode != http.StatusOK {
		return []JiraTicketsMsg{}, fmt.Errorf("jira API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []JiraTicketsMsg{}, err
	}

	var result JiraSearchResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return []JiraTicketsMsg{}, err
	}

	newChoices := []JiraTicketsMsg{}
	for _, issue := range result.Issues {
		newChoices = append(newChoices, JiraTicketsMsg{
			Key:     issue.Key,
			Type:    issue.Fields.IssueType.Name,
			Summary: issue.Fields.Summary,
			Status:  issue.Fields.Status.Name,
			Created: issue.Fields.Created,
		})
	}

	return newChoices, nil
}

type JiraSearchResult struct {
	Issues []Issue `json:"issues"`
	Total  int     `json:"total"`
}

type Issue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Created string `json:"created"`
	} `json:"fields"`
}

type Fields struct {
	Summary string `json:"summary"`
	Status  Status `json:"status"`
}

type Status struct {
	Name string `json:"name"`
}
