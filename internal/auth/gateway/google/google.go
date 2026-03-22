package google

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	authmodel "cityhawk/backend/internal/auth/model"
	appconfig "cityhawk/backend/internal/config"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const StateCookieName = "google_oauth_state"

func NewOAuthConfig(cfg appconfig.OAuthProviderConfig) (*oauth2.Config, error) {
	if cfg.ClientID == "" {
		return nil, errors.New("GOOGLE_OAUTH_CLIENT_ID is required")
	}
	if cfg.ClientSecret == "" {
		return nil, errors.New("GOOGLE_OAUTH_CLIENT_SECRET is required")
	}
	if cfg.RedirectURL == "" {
		return nil, errors.New("GOOGLE_OAUTH_REDIRECT_URL is required")
	}

	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
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

type Gateway struct {
	cfg *oauth2.Config
}

func NewGateway(cfg *oauth2.Config) *Gateway {
	return &Gateway{cfg: cfg}
}

func (g *Gateway) FetchIdentity(ctx context.Context, code string) (authmodel.OAuthIdentity, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return authmodel.OAuthIdentity{}, platformerrors.ErrInvalidOAuthCode
	}

	token, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return authmodel.OAuthIdentity{}, platformerrors.ErrInvalidOAuthCode
	}

	profile, err := FetchProfile(ctx, token.AccessToken)
	if err != nil {
		return authmodel.OAuthIdentity{}, platformerrors.ErrOAuthProfileFetch
	}

	username := strings.TrimSpace(profile.Name)
	if username == "" {
		username = strings.TrimSpace(profile.GivenName + "_" + profile.FamilyName)
	}

	return authmodel.OAuthIdentity{
		Provider:  "google",
		SubjectID: strings.TrimSpace(profile.ID),
		Email:     strings.TrimSpace(profile.Email),
		Username:  username,
	}, nil
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
