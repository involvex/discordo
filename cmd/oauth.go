package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
)

var oauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("DISGO_CLI_CLIENT_ID"),
	ClientSecret: os.Getenv("DISGO_CLI_CLIENT_SECRET"),
	RedirectURL:  "http://localhost:4444/oauth/callback",
	Scopes:       []string{"identify"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://discord.com/api/oauth2/authorize",
		TokenURL: "https://discord.com/api/oauth2/token",
	},
}

func generateState() string {
	// Simple state generation - in production, use a cryptographically secure random string
	return fmt.Sprintf("disgo-cli-%d", time.Now().Unix())
}

// OAuth user info structure
type DiscordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// Exchange code for token and get user info
func exchangeCodeForToken(code string) (string, error) {
	ctx := context.Background()

	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// Use Discord API to get user information
	client := &http.Client{}

	// Create request to Discord API user endpoint
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("discord API returned status: %s", resp.Status)
	}

	var user DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to decode user info: %w", err)
	}

	slog.Info("OAuth login successful", "user_id", user.ID, "username", user.Username)

	// Discord returns the OAuth2 access token, which can be used as a user token
	userToken := token.AccessToken

	return userToken, nil
}

func initOAuthConfig() {
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("DISGO_CLI_CLIENT_ID"),
		ClientSecret: os.Getenv("DISGO_CLI_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:4444/oauth/callback",
		Scopes:       []string{"identify"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
}
