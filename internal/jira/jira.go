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

	// JIRA REST API endpoint
	url := jiraUrl + "/rest/api/3/search"

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return JiraSearchResult{}, err
	}

	// Add authentication (Basic Auth with API token)
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + apiToken))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	// Add query parameters
	q := req.URL.Query()
	q.Add("jql", "assignee = currentUser()") // Get tickets assigned to you
	q.Add("maxResults", "50")
	req.URL.RawQuery = q.Encode()

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return JiraSearchResult{}, err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return JiraSearchResult{}, err
	}

	// Parse JSON
	var result JiraSearchResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return JiraSearchResult{}, err
	}

	// // Print tickets
	// for _, issue := range result.Issues {
	//     fmt.Printf("%s: %s\n", issue.Key, issue.Fields.Summary)
	// }

	return result, nil
}

// Structs for JSON response
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
