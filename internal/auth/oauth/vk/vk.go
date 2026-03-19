package vk

import (
	"context"
	"errors"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

const (
	ClientIDEnv     = "VK_OAUTH_CLIENT_ID"
	ClientSecretEnv = "VK_OAUTH_CLIENT_SECRET"
	RedirectURLEnv  = "VK_OAUTH_REDIRECT_URL"
	StateCookieName = "vk_oauth_state"
)

func NewConfigFromEnv() (*oauth2.Config, error) {
	clientID := strings.TrimSpace(os.Getenv(ClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(ClientSecretEnv))
	redirectURL := strings.TrimSpace(os.Getenv(RedirectURLEnv))

	if clientID == "" {
		return nil, errors.New("VK_OAUTH_CLIENT_ID is required")
	}
	if clientSecret == "" {
		return nil, errors.New("VK_OAUTH_CLIENT_SECRET is required")
	}
	if redirectURL == "" {
		return nil, errors.New("VK_OAUTH_REDIRECT_URL is required")
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     vk.Endpoint,
		Scopes:       []string{"email"},
	}, nil
}

func AuthCodeURL(cfg *oauth2.Config, state string) string {
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code string) (*oauth2.Token, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("oauth code is required")
	}
	return cfg.Exchange(ctx, code)
}

