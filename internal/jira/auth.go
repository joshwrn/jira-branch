package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zalando/go-keyring"
)

type Credentials struct {
	JiraURL  string `json:"jira_url"`
	Email    string `json:"email"`
	APIToken string `json:"api_token"`
}

func StoreCredentials(credentials Credentials) error {
	data, err := json.Marshal(credentials)
	if err != nil {
		return err
	}
	return keyring.Set("jira-cli", "credentials", string(data))
}

func LoadCredentials() (Credentials, error) {
	var credentials Credentials
	data, err := keyring.Get("jira-cli", "credentials")
	if err != nil {
		return credentials, err
	}
	err = json.Unmarshal([]byte(data), &credentials)
	return credentials, err
}

func ClearCredentials() error {
	return keyring.Delete("jira-cli", "credentials")
}

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-myself/#api-rest-api-3-myself-get
func ValidateCredentials(credentials Credentials) error {
	client := newClient()
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/api/3/myself", credentials.JiraURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Basic "+createAuthHeader(credentials))
	req.Header.Add("Accept", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Jira: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid credentials: check your email and API token")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from Jira API: %d", resp.StatusCode)
	}

	return nil
}

func createAuthHeader(credentials Credentials) string {
	return base64.StdEncoding.EncodeToString([]byte(credentials.Email + ":" + credentials.APIToken))
}
