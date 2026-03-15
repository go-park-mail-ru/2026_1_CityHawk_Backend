package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleClientIDEnv     = "GOOGLE_OAUTH_CLIENT_ID"
	googleClientSecretEnv = "GOOGLE_OAUTH_CLIENT_SECRET"
	googleRedirectURLEnv  = "GOOGLE_OAUTH_REDIRECT_URL"
	googleStateCookieName = "google_oauth_state"
)

func newGoogleOAuthConfigFromEnv() (*oauth2.Config, error) {
	clientID := strings.TrimSpace(os.Getenv(googleClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(googleClientSecretEnv))
	redirectURL := strings.TrimSpace(os.Getenv(googleRedirectURLEnv))

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

func setGoogleOAuthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     googleStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().UTC().Add(10 * time.Minute),
		MaxAge:   600,
	})
}

func clearGoogleOAuthStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     googleStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func readGoogleOAuthStateCookie(r *http.Request) string {
	c, err := r.Cookie(googleStateCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(c.Value)
}

func googleLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return errMethodNotAllowed
		}

		state, err := newOAuthState()
		if err != nil {
			return errInternal
		}

		setGoogleOAuthStateCookie(w, state)
		http.Redirect(w, r, cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
		return nil
	})
}

func googleCallbackHandler(cfg *oauth2.Config, store *userStore, auth *authService) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return errMethodNotAllowed
		}

		receivedState := strings.TrimSpace(r.URL.Query().Get("state"))
		if receivedState == "" {
			return newHTTPError(http.StatusUnauthorized, "missing oauth state")
		}

		savedState := readGoogleOAuthStateCookie(r)
		if savedState == "" || savedState != receivedState {
			return newHTTPError(http.StatusUnauthorized, "invalid oauth state")
		}
		clearGoogleOAuthStateCookie(w)

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
			return errInternal
		}

		resp, err := auth.issueTokenPair(u)
		if err != nil {
			return errIssueTokens
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
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
		if errors.Is(err, errEmailExists) {
			if existing, ok := store.getByEmail(email); ok {
				return existing, nil
			}
		}
		return user{}, err
	}

	return u, nil
}
