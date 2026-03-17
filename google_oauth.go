package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleClientIDEnv     = "GOOGLE_OAUTH_CLIENT_ID"
	googleClientSecretEnv = "GOOGLE_OAUTH_CLIENT_SECRET"
	googleRedirectURLEnv  = "GOOGLE_OAUTH_REDIRECT_URL"
	googleStateCookieName = "google_oauth_state"
	googleUserInfoURL     = "https://www.googleapis.com/oauth2/v2/userinfo"
	required              = "is required"
)

func newGoogleOAuthConfigFromEnv() (*oauth2.Config, error) {
	clientID := strings.TrimSpace(os.Getenv(googleClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(googleClientSecretEnv))
	redirectURL := strings.TrimSpace(os.Getenv(googleRedirectURLEnv))

	if clientID == "" {
		return nil, fmt.Errorf("%s %s", googleClientIDEnv, required)
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("%s %s", googleClientSecretEnv, required)
	}
	if redirectURL == "" {
		return nil, fmt.Errorf("%s %s", googleRedirectURLEnv, required)
	}

	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"openid", "email", "profile"},
	}, nil
}

func googleLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		state, err := newOAuthState()
		if err != nil {
			return ErrInternal
		}

		setOAuthStateCookie(w, state, googleStateCookieName)
		http.Redirect(w, r, cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
		return nil
	})
}

func googleCallbackHandler(cfg *oauth2.Config, store *userStore, auth *authService) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		receivedState := strings.TrimSpace(r.URL.Query().Get("state"))
		if receivedState == "" {
			return newHTTPError(http.StatusUnauthorized, "missing oauth state")
		}

		savedState := readOAuthStateCookie(r, googleStateCookieName)
		if savedState == "" || savedState != receivedState {
			return newHTTPError(http.StatusUnauthorized, "invalid oauth state")
		}
		clearOAuthStateCookie(w, googleStateCookieName)

		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			return newHTTPError(http.StatusBadRequest, "missing oauth code")
		}

		token, err := cfg.Exchange(r.Context(), code)
		if err != nil {
			return newHTTPError(http.StatusUnauthorized, "invalid oauth code")
		}

		profile, err := fetchGoogleProfile(r.Context(), token.AccessToken)
		if err != nil {
			return newHTTPError(http.StatusUnauthorized, "failed to fetch google profile")
		}

		u, err := findOrCreateUserFromGoogleProfile(store, profile)
		if err != nil {
			return ErrInternal
		}

		resp, err := auth.issueTokenPair(u)
		if err != nil {
			return ErrIssueTokens
		}

		setRefreshCookie(w, resp.RefreshToken, auth.refreshTTL)
		setAccessCookie(w, resp.AccessToken, auth.accessTTL)
		writeJSON(w, http.StatusOK, map[string]string{"message": "google login successful"})
		return nil
	})
}

type googleProfile struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

func fetchGoogleProfile(ctx context.Context, accessToken string) (googleProfile, error) {
	if strings.TrimSpace(accessToken) == "" {
		return googleProfile{}, errors.New("empty access token")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return googleProfile{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return googleProfile{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return googleProfile{}, errors.New("google userinfo failed")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return googleProfile{}, err
	}

	var profile googleProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return googleProfile{}, err
	}

	if strings.TrimSpace(profile.ID) == "" {
		return googleProfile{}, errors.New("google profile id is empty")
	}
	return profile, nil
}

func findOrCreateUserFromGoogleProfile(store *userStore, profile googleProfile) (user, error) {
	email := strings.ToLower(strings.TrimSpace(profile.Email))
	if email == "" {
		email = "google_" + sanitizePart(profile.ID) + "@google.local"
	}

	if u, ok := store.getByEmail(email); ok {
		return u, nil
	}

	usernameBase := sanitizePart(profile.Name)
	if usernameBase == "" {
		usernameBase = sanitizePart(profile.GivenName + "_" + profile.FamilyName)
	}
	if usernameBase == "" {
		usernameBase = "google_" + sanitizePart(profile.ID)
	}

	username := usernameBase
	if len(username) < 3 {
		username = "google_user"
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
			existing, _ := store.getByEmail(email)
			return existing, nil
		}
		return user{}, err
	}

	return u, nil
}
