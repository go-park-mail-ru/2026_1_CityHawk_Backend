package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	yandexClientIDEnv     = "YANDEX_OAUTH_CLIENT_ID"
	yandexClientSecretEnv = "YANDEX_OAUTH_CLIENT_SECRET"
	yandexRedirectURLEnv  = "YANDEX_OAUTH_REDIRECT_URL"
	yandexStateCookieName = "yandex_oauth_state"
)

var yandexEndpoint = oauth2.Endpoint{
	AuthURL:  "https://oauth.yandex.ru/authorize",
	TokenURL: "https://oauth.yandex.ru/token",
}

func newYandexOAuthConfigFromEnv() (*oauth2.Config, error) {
	clientID := strings.TrimSpace(os.Getenv(yandexClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(yandexClientSecretEnv))
	redirectURL := strings.TrimSpace(os.Getenv(yandexRedirectURLEnv))

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
		Endpoint:     yandexEndpoint,
		Scopes:       []string{"login:email", "login:info"},
	}, nil
}

func setYandexOAuthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     yandexStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().UTC().Add(10 * time.Minute),
		MaxAge:   600,
	})
}

func clearYandexOAuthStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     yandexStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func readYandexOAuthStateCookie(r *http.Request) string {
	c, err := r.Cookie(yandexStateCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(c.Value)
}

func yandexLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		state, err := newOAuthState()
		if err != nil {
			return ErrInternal
		}

		setYandexOAuthStateCookie(w, state)
		http.Redirect(w, r, cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
		return nil
	})
}

func yandexCallbackHandler(cfg *oauth2.Config, store *userStore, auth *authService) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		receivedState := strings.TrimSpace(r.URL.Query().Get("state"))
		if receivedState == "" {
			return newHTTPError(http.StatusUnauthorized, "missing oauth state")
		}

		savedState := readYandexOAuthStateCookie(r)
		if savedState == "" || savedState != receivedState {
			return newHTTPError(http.StatusUnauthorized, "invalid oauth state")
		}
		clearYandexOAuthStateCookie(w)

		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			return newHTTPError(http.StatusBadRequest, "missing oauth code")
		}

		token, err := cfg.Exchange(r.Context(), code)
		if err != nil {
			return newHTTPError(http.StatusUnauthorized, "invalid oauth code")
		}

		profile, err := fetchYandexProfile(r.Context(), token.AccessToken)
		if err != nil {
			return newHTTPError(http.StatusUnauthorized, "failed to fetch yandex profile")
		}

		u, err := findOrCreateUserFromYandexProfile(store, profile)
		if err != nil {
			return ErrInternal
		}

		resp, err := auth.issueTokenPair(u)
		if err != nil {
			return ErrIssueTokens
		}

		setRefreshCookie(w, resp.RefreshToken, auth.refreshTTL)
		setAccessCookie(w, resp.AccessToken, auth.accessTTL)
		writeJSON(w, http.StatusOK, map[string]string{"message": "yandex login successful"})
		return nil
	})
}

type yandexProfile struct {
	ID           string `json:"id"`
	Login        string `json:"login"`
	DefaultEmail string `json:"default_email"`
}

func fetchYandexProfile(ctx context.Context, accessToken string) (yandexProfile, error) {
	if strings.TrimSpace(accessToken) == "" {
		return yandexProfile{}, errors.New("empty access token")
	}

	u, err := url.Parse("https://login.yandex.ru/info")
	if err != nil {
		return yandexProfile{}, err
	}
	q := u.Query()
	q.Set("format", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return yandexProfile{}, err
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return yandexProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return yandexProfile{}, errors.New("yandex userinfo failed")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return yandexProfile{}, err
	}

	var profile yandexProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return yandexProfile{}, err
	}

	if strings.TrimSpace(profile.ID) == "" {
		return yandexProfile{}, errors.New("yandex profile id is empty")
	}
	return profile, nil
}

func findOrCreateUserFromYandexProfile(store *userStore, profile yandexProfile) (user, error) {
	email := strings.ToLower(strings.TrimSpace(profile.DefaultEmail))
	if email == "" {
		email = "yandex_" + sanitizePart(profile.ID) + "@yandex.local"
	}

	if u, ok := store.getByEmail(email); ok {
		return u, nil
	}

	username := sanitizePart(profile.Login)
	if username == "" {
		username = "ya_" + sanitizePart(profile.ID)
	}
	if len(username) < 3 {
		username = "ya_user"
	}
	if len(username) > 32 {
		username = username[:32]
	}

	password, err := randomHex(16)
	if err != nil {
		return user{}, err
	}

	u, err := createUser(email, password, username)
	if err != nil {
		return user{}, err
	}

	if err := store.create(u); err != nil {
		if errors.Is(err, ErrEmailExists) {
			if existing, ok := store.getByEmail(email); ok {
				return existing, nil
			}
		}
		return user{}, err
	}

	return u, nil
}
