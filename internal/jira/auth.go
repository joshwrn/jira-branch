package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
}

func getCloudId(token *oauth2.Token) (string, error) {
	req, err := http.NewRequest("GET", "https://api.atlassian.com/oauth/token/accessible-resources", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+token.AccessToken)
	req.Header.Add("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var resources []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&resources); err != nil {
		return "", err
	}

	if len(resources) == 0 {
		return "", fmt.Errorf("no accessible resources found")
	}

	return resources[0].ID, nil
}

func GetToken() (*oauth2.Token, error) {
	config, authURL := GetAuthUrlAndConfig()

	storedTokens, err := loadTokens("jira-cli", "jira-cli")
	if err == nil {
		token := &oauth2.Token{
			AccessToken:  storedTokens.AccessToken,
			RefreshToken: storedTokens.RefreshToken,
			TokenType:    storedTokens.TokenType,
			Expiry:       time.Unix(storedTokens.ExpiresAt, 0),
		}

		tokenSource := config.TokenSource(context.Background(), token)

		newToken, err := tokenSource.Token()

		if err == nil {
			return newToken, nil
		}

		fmt.Println("failed to get/refresh token: %w", err)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			go func() {
				time.Sleep(1 * time.Second)
				server.Close()
			}()
		}()

		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			errDescription := r.URL.Query().Get("error_description")
			errCh <- fmt.Errorf("oauth error: %s - %s", errMsg, errDescription)
			fmt.Fprintf(w, "Error: %s", errDescription)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			fmt.Fprintf(w, "Error: No authorization code received")
			return
		}

		codeCh <- code
		fmt.Fprintf(w, `
				<html>
						<body>
								<h2>✅ Authorization Successful!</h2>
								<p>You can close this window and return to your CLI.</p>
						</body>
				</html>
		`)
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server error: %v", err)
		}
	}()

	if err := openBrowserWithTempFile(authURL); err != nil {
		fmt.Printf("⚠️  Could not open browser automatically. Please visit:\n%s\n", authURL)
	}
	defer os.Remove(tempFileName)

	select {
	case code := <-codeCh:
		ctx := context.Background()
		token, err := config.Exchange(ctx, code)

		if err != nil {
			return nil, fmt.Errorf("failed to exchange code for token: %v", err)
		}

		fmt.Println("✅ Authentication successful!")
		return token, nil

	case err := <-errCh:
		return nil, err

	case <-time.After(5 * time.Minute):
		server.Close()
		return nil, fmt.Errorf("authentication timeout after 5 minutes")
	}

}

func storeTokens(service, user string, tokens TokenPair) error {
	data, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	return keyring.Set(service, user, string(data))
}

func loadTokens(service, user string) (TokenPair, error) {
	var tokens TokenPair
	data, err := keyring.Get(service, user)
	if err != nil {
		return tokens, err
	}
	err = json.Unmarshal([]byte(data), &tokens)
	return tokens, err
}

func GetAuthUrlAndConfig() (*oauth2.Config, string) {
	config := &oauth2.Config{
		ClientID:     os.Getenv("AT_ID"),
		ClientSecret: os.Getenv("AT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"read:jira-work"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://auth.atlassian.com/authorize",
			TokenURL: "https://auth.atlassian.com/oauth/token",
		},
	}

	authURL := config.AuthCodeURL("state-token",
		oauth2.SetAuthURLParam("audience", "api.atlassian.com"),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	return config, authURL
}
