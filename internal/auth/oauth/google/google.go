package google

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	ClientIDEnv     = "GOOGLE_OAUTH_CLIENT_ID"
	ClientSecretEnv = "GOOGLE_OAUTH_CLIENT_SECRET"
	RedirectURLEnv  = "GOOGLE_OAUTH_REDIRECT_URL"
	StateCookieName = "google_oauth_state"
)

func NewConfigFromEnv() (*oauth2.Config, error) {
	clientID := strings.TrimSpace(os.Getenv(ClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(ClientSecretEnv))
	redirectURL := strings.TrimSpace(os.Getenv(RedirectURLEnv))

	if clientID == "" {
		return nil, errors.New("GOOGLE_OAUTH_CLIENT_ID is required")
	}
	if clientSecret == "" {
		return nil, errors.New("GOOGLE_OAUTH_CLIENT_SECRET is required")
	}
	if redirectURL == "" {
		return nil, errors.New("GOOGLE_OAUTH_REDIRECT_URL is required")
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "email", "profile"},
	}, nil
}

type Profile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

func FetchProfile(ctx context.Context, accessToken string) (Profile, error) {
	if strings.TrimSpace(accessToken) == "" {
		return Profile{}, errors.New("empty access token")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return Profile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Profile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Profile{}, errors.New("google userinfo failed")
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
		return Profile{}, errors.New("google profile id is empty")
	}

	return profile, nil
}

