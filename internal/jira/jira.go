package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

type JiraTicketsMsg struct {
	Key     string
	Summary string
	Type    string
	Status  string
	Created string
}

func GetJiraTickets(token *oauth2.Token) ([]JiraTicketsMsg, error) {

	storeTokens("jira-cli", "jira-cli", TokenPair{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry.Unix(),
		TokenType:    token.TokenType,
	})

	cloudId, err := getCloudId(token)
	if err != nil {
		return []JiraTicketsMsg{}, err
	}

	url := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/search", cloudId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []JiraTicketsMsg{}, err
	}

	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add("jql", "assignee = currentUser() AND status != Done order by createdDate")
	q.Add("fields", "summary,status,issuetype,assignee,created")
	q.Add("maxResults", "100")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []JiraTicketsMsg{}, err
	}
	defer resp.Body.Close()

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
