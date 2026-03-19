package yandex

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

const (
	ClientIDEnv     = "YANDEX_OAUTH_CLIENT_ID"
	ClientSecretEnv = "YANDEX_OAUTH_CLIENT_SECRET"
	RedirectURLEnv  = "YANDEX_OAUTH_REDIRECT_URL"
	StateCookieName = "yandex_oauth_state"
)

var Endpoint = oauth2.Endpoint{
	AuthURL:  "https://oauth.yandex.ru/authorize",
	TokenURL: "https://oauth.yandex.ru/token",
}

func NewConfigFromEnv() (*oauth2.Config, error) {
	clientID := strings.TrimSpace(os.Getenv(ClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(ClientSecretEnv))
	redirectURL := strings.TrimSpace(os.Getenv(RedirectURLEnv))

	if clientID == "" {
		return nil, errors.New("YANDEX_OAUTH_CLIENT_ID is required")
	}
	if clientSecret == "" {
		return nil, errors.New("YANDEX_OAUTH_CLIENT_SECRET is required")
	}
	if redirectURL == "" {
		return nil, errors.New("YANDEX_OAUTH_REDIRECT_URL is required")
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     Endpoint,
		Scopes:       []string{"login:email", "login:info"},
	}, nil
}

type Profile struct {
	ID           string `json:"id"`
	Login        string `json:"login"`
	DefaultEmail string `json:"default_email"`
}

func FetchProfile(ctx context.Context, accessToken string) (Profile, error) {
	if strings.TrimSpace(accessToken) == "" {
		return Profile{}, errors.New("empty access token")
	}

	u, err := url.Parse("https://login.yandex.ru/info")
	if err != nil {
		return Profile{}, err
	}
	q := u.Query()
	q.Set("format", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Profile{}, err
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Profile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Profile{}, errors.New("yandex userinfo failed")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Profile{}, err
	}

	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return Profile{}, err
	}
	if strings.TrimSpace(profile.ID) == "" {
		return Profile{}, errors.New("yandex profile id is empty")
	}

	return profile, nil
}

