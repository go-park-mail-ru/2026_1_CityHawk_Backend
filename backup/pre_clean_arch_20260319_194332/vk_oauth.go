package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

const (
	vkClientIDEnv     = "VK_OAUTH_CLIENT_ID"
	vkClientSecretEnv = "VK_OAUTH_CLIENT_SECRET"
	vkRedirectURLEnv  = "VK_OAUTH_REDIRECT_URL"
	vkStateCookieName = "vk_oauth_state"
)

func newVKOAuthConfigFromEnv() (*oauth2.Config, error) {
	clientID := strings.TrimSpace(os.Getenv(vkClientIDEnv))
	clientSecret := strings.TrimSpace(os.Getenv(vkClientSecretEnv))
	redirectURL := strings.TrimSpace(os.Getenv(vkRedirectURLEnv))

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

func newOAuthState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func vkAuthCodeURL(cfg *oauth2.Config, state string) string {
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func exchangeVKCode(ctx context.Context, cfg *oauth2.Config, code string) (*oauth2.Token, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("oauth code is required")
	}
	return cfg.Exchange(ctx, code)
}

func vkTokenSource(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) oauth2.TokenSource {
	return cfg.TokenSource(ctx, token)
}

func setVKOAuthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     vkStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().UTC().Add(10 * time.Minute),
		MaxAge:   600,
	})
}

func clearVKOAuthStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     vkStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func readVKOAuthStateCookie(r *http.Request) string {
	c, err := r.Cookie(vkStateCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(c.Value)
}

func vkLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		state, err := newOAuthState()
		if err != nil {
			return ErrInternal
		}

		setVKOAuthStateCookie(w, state)
		http.Redirect(w, r, vkAuthCodeURL(cfg, state), http.StatusFound)
		return nil
	})
}

func vkCallbackHandler(cfg *oauth2.Config, store *userStore, auth *authService) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return ErrMethodNotAllowed
		}

		receivedState := strings.TrimSpace(r.URL.Query().Get("state"))
		if receivedState == "" {
			return newHTTPError(http.StatusUnauthorized, "missing oauth state")
		}

		savedState := readVKOAuthStateCookie(r)
		if savedState == "" || savedState != receivedState {
			return newHTTPError(http.StatusUnauthorized, "invalid oauth state")
		}
		clearVKOAuthStateCookie(w)

		if strings.TrimSpace(r.URL.Query().Get("error")) != "" {
			return newHTTPError(http.StatusUnauthorized, "vk oauth denied")
		}

		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			return newHTTPError(http.StatusBadRequest, "missing oauth code")
		}

		token, err := exchangeVKCode(r.Context(), cfg, code)
		if err != nil {
			return newHTTPError(http.StatusUnauthorized, "invalid oauth code")
		}

		u, err := findOrCreateUserFromVKToken(store, token)
		if err != nil {
			return ErrInternal
		}

		resp, err := auth.issueTokenPair(u)
		if err != nil {
			return ErrIssueTokens
		}

		setRefreshCookie(w, resp.RefreshToken, auth.refreshTTL)
		setAccessCookie(w, resp.AccessToken, auth.accessTTL)
		writeJSON(w, http.StatusOK, map[string]string{"message": "vk login successful"})
		return nil
	})
}

func findOrCreateUserFromVKToken(store *userStore, token *oauth2.Token) (user, error) {
	email := extractVKEmail(token)
	vkUserID := extractVKUserID(token)

	if email == "" {
		if vkUserID == "" {
			return user{}, errors.New("vk user identity is empty")
		}
		email = "vk_" + vkUserID + "@vk.local"
	}

	if u, ok := store.getByEmail(email); ok {
		return u, nil
	}

	username := buildVKUsername(vkUserID, email)
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

func extractVKEmail(token *oauth2.Token) string {
	if token == nil {
		return ""
	}
	email, _ := token.Extra("email").(string)
	return strings.ToLower(strings.TrimSpace(email))
}

func extractVKUserID(token *oauth2.Token) string {
	if token == nil {
		return ""
	}

	raw := token.Extra("user_id")
	switch v := raw.(type) {
	case string:
		return sanitizePart(v)
	case float64:
		if v == math.Trunc(v) {
			return sanitizePart(strconv.FormatInt(int64(v), 10))
		}
		return sanitizePart(fmt.Sprintf("%.0f", v))
	case int:
		return sanitizePart(strconv.Itoa(v))
	case int64:
		return sanitizePart(strconv.FormatInt(v, 10))
	default:
		return sanitizePart(fmt.Sprintf("%v", v))
	}
}

func buildVKUsername(vkUserID, email string) string {
	base := "vk_user"
	if vkUserID != "" {
		base = "vk_" + sanitizePart(vkUserID)
	} else {
		localPart := email
		if at := strings.Index(localPart, "@"); at > 0 {
			localPart = localPart[:at]
		}
		if sanitized := sanitizePart(localPart); sanitized != "" {
			base = sanitized
		}
	}

	if len(base) < 3 {
		base = base + "_id"
	}
	if len(base) > 32 {
		base = base[:32]
	}
	return base
}

func sanitizePart(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
