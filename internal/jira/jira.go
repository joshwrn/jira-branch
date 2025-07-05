package jira

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func createJiraUrl(endpoint string) string {
	credentials, err := LoadCredentials()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s/rest/api/3/%s", credentials.JiraURL, endpoint)
}

type Transition struct {
	ID string `json:"id"`
}
type TransitionIssueBody struct {
	Transition Transition `json:"transition"`
}

func MarkAsInProgress(credentials Credentials, issueKey string) error {
	url := createJiraUrl(fmt.Sprintf("issue/%s/transitions", issueKey))

	body, err := json.Marshal(TransitionIssueBody{
		Transition: Transition{ID: "21"},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(credentials.Email + ":" + credentials.APIToken))

	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("authentication failed: check your credentials")
	}

	if resp.StatusCode != 204 {
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

func GetJiraTickets(credentials Credentials) ([]JiraTicketsMsg, error) {
	url := createJiraUrl("search")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []JiraTicketsMsg{}, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(credentials.Email + ":" + credentials.APIToken))

	req.Header.Add("Authorization", "Basic "+auth)
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

	if resp.StatusCode == 401 {
		return []JiraTicketsMsg{}, fmt.Errorf("authentication failed: check your credentials")
	}

	if resp.StatusCode != 200 {
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
