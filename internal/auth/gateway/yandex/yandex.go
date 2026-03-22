package yandex

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	authmodel "cityhawk/backend/internal/auth/model"
	appconfig "cityhawk/backend/internal/config"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"golang.org/x/oauth2"
)

const StateCookieName = "yandex_oauth_state"

var Endpoint = oauth2.Endpoint{
	AuthURL:  "https://oauth.yandex.ru/authorize",
	TokenURL: "https://oauth.yandex.ru/token",
}

func NewOAuthConfig(cfg appconfig.OAuthProviderConfig) (*oauth2.Config, error) {
	if cfg.ClientID == "" {
		return nil, errors.New("YANDEX_OAUTH_CLIENT_ID is required")
	}
	if cfg.ClientSecret == "" {
		return nil, errors.New("YANDEX_OAUTH_CLIENT_SECRET is required")
	}
	if cfg.RedirectURL == "" {
		return nil, errors.New("YANDEX_OAUTH_REDIRECT_URL is required")
	}

	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     Endpoint,
		Scopes:       []string{"login:email", "login:info"},
	}, nil
}

type Profile struct {
	ID           string `json:"id"`
	Login        string `json:"login"`
	DefaultEmail string `json:"default_email"`
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

	return authmodel.OAuthIdentity{
		Provider:  "yandex",
		SubjectID: strings.TrimSpace(profile.ID),
		Email:     strings.TrimSpace(profile.DefaultEmail),
		Username:  strings.TrimSpace(profile.Login),
	}, nil
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
