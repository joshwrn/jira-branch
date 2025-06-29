package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetJiraTickets() (JiraSearchResult, error) {
	apiToken := os.Getenv("JIRA_API_TOKEN")
	username := os.Getenv("JIRA_USERNAME")
	jiraUrl := os.Getenv("JIRA_URL")

	if apiToken == "" {
		return JiraSearchResult{}, fmt.Errorf("JIRA_API_TOKEN environment variable is required")
	}
	if username == "" {
		return JiraSearchResult{}, fmt.Errorf("JIRA_USERNAME environment variable is required")
	}
	if jiraUrl == "" {
		return JiraSearchResult{}, fmt.Errorf("JIRA_URL environment variable is required")
	}

	url := jiraUrl + "/rest/api/3/search"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return JiraSearchResult{}, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + apiToken))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	q.Add("jql", "assignee = currentUser() order by createdDate")
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
	Fields Fields `json:"fields"`
}

type Fields struct {
	Summary string `json:"summary"`
	Status  Status `json:"status"`
}

type Status struct {
	Name string `json:"name"`
}
