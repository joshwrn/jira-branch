package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type JiraTicketsMsg struct {
	Key     string
	Summary string
	Type    string
	Status  string
	Created string
}

func GetJiraTickets() (JiraSearchResult, error) {
	token, err := getToken()
	if err != nil {
		fmt.Println("err-getToken", err)
		return JiraSearchResult{}, err
	}
	storeTokens("jira-cli", "jira-cli", TokenPair{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry.Unix(),
		TokenType:    token.TokenType,
	})

	cloudId, err := getCloudId(token)
	if err != nil {
		return JiraSearchResult{}, err
	}

	url := fmt.Sprintf("https://api.atlassian.com/ex/jira/%s/rest/api/3/search", cloudId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return JiraSearchResult{}, err
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
		return JiraSearchResult{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return JiraSearchResult{}, err
	}

	var result JiraSearchResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return JiraSearchResult{}, err
	}

	return result, nil
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
