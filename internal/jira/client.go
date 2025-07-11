package jira

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func newClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) createRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	credentials, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	url, err := createJiraUrl(endpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Basic "+createAuthHeader(credentials))
	req.Header.Add("Accept", "application/json")

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := c.createRequest(method, endpoint, body)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

func createJiraUrl(endpoint string) (string, error) {
	credentials, err := LoadCredentials()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/rest/api/3/%s", credentials.JiraURL, endpoint), nil
}
