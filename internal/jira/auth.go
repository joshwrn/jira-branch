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

func ValidateCredentials(credentials Credentials) error {
	url := fmt.Sprintf("%s/rest/api/3/myself", credentials.JiraURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(credentials.Email + ":" + credentials.APIToken))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Jira: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid credentials: check your email and API token")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response from Jira API: %d", resp.StatusCode)
	}

	return nil
}

func createAuthHeader(credentials Credentials) string {
	return base64.StdEncoding.EncodeToString([]byte(credentials.Email + ":" + credentials.APIToken))
}
