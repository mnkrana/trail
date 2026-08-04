package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Client struct {
	service *calendar.Service
}

type TokenData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

func NewOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/api/oauth/calendar/callback",
		Scopes:       []string{calendar.CalendarReadonlyScope},
	}
}

func NewClient(ctx context.Context) (*Client, error) {
	cfg := NewOAuthConfig()

	tokens, err := loadTokens()
	if err != nil {
		return nil, fmt.Errorf("not authenticated — run 'trail auth' first")
	}

	token := &oauth2.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		Expiry:       tokens.Expiry,
	}

	ts := cfg.TokenSource(ctx, token)

	realTok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	if realTok.AccessToken != tokens.AccessToken {
		newTokens := &TokenData{
			AccessToken:  realTok.AccessToken,
			RefreshToken: realTok.RefreshToken,
			TokenType:    realTok.TokenType,
			Expiry:       realTok.Expiry,
		}
		if err := saveTokens(newTokens); err != nil {
			fmt.Printf("Warning: failed to save refreshed tokens: %v\n", err)
		}
	}

	srv, err := calendar.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	return &Client{service: srv}, nil
}

func tokensPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "trail")
	return filepath.Join(dir, "tokens.json"), nil
}

func loadTokens() (*TokenData, error) {
	path, err := tokensPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tokens TokenData
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, err
	}

	return &tokens, nil
}

func saveTokens(tokens *TokenData) error {
	path, err := tokensPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Client) FetchEvents(ctx context.Context, from, to time.Time) ([]*calendar.Event, error) {
	timeMin := from.Format(time.RFC3339)
	timeMax := to.Format(time.RFC3339)

	var events []*calendar.Event
	pageToken := ""

	for {
		call := c.service.Events.List("primary").
			ShowDeleted(false).
			SingleEvents(true).
			TimeMin(timeMin).
			TimeMax(timeMax).
			OrderBy("startTime").
			MaxResults(250)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch events: %w", err)
		}

		events = append(events, result.Items...)

		pageToken = result.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return events, nil
}
