package http

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	authmodel "cityhawk/backend/internal/auth/model"
	oauthcommon "cityhawk/backend/internal/auth/oauth/common"
	oauthgoogle "cityhawk/backend/internal/auth/oauth/google"
	oauthvk "cityhawk/backend/internal/auth/oauth/vk"
	oauthyandex "cityhawk/backend/internal/auth/oauth/yandex"
	authusecase "cityhawk/backend/internal/auth/usecase"
	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	usermodel "cityhawk/backend/internal/user/model"
	"golang.org/x/oauth2"
)

type OAuthUserService interface {
	FindOrCreateFromOAuth(identity authusecase.OAuthIdentity) (usermodel.User, error)
}

type TokenIssuer interface {
	IssueTokenPair(u usermodel.User) (authmodel.TokenPair, error)
}

func GoogleLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		state, err := oauthcommon.NewState()
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusInternalServerError, "internal error")
		}

		oauthcommon.SetStateCookie(w, oauthgoogle.StateCookieName, state, 10*time.Minute)
		http.Redirect(w, r, cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
		return nil
	})
}

func GoogleCallbackHandler(cfg *oauth2.Config, users OAuthUserService, issuer TokenIssuer, accessTTL, refreshTTL time.Duration) http.HandlerFunc {
	return platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		receivedState := strings.TrimSpace(r.URL.Query().Get("state"))
		if receivedState == "" {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "missing oauth state")
		}
		savedState := oauthcommon.ReadStateCookie(r, oauthgoogle.StateCookieName)
		if savedState == "" || savedState != receivedState {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "invalid oauth state")
		}
		oauthcommon.ClearStateCookie(w, oauthgoogle.StateCookieName)

		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			return platformmiddleware.NewHTTPError(http.StatusBadRequest, "missing oauth code")
		}

		token, err := cfg.Exchange(r.Context(), code)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "invalid oauth code")
		}

		profile, err := oauthgoogle.FetchProfile(r.Context(), token.AccessToken)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "failed to fetch google profile")
		}

		identity := buildGoogleIdentity(profile)
		u, err := users.FindOrCreateFromOAuth(identity)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusInternalServerError, "internal error")
		}

		resp, err := issuer.IssueTokenPair(u)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusInternalServerError, "failed to issue tokens")
		}

		SetRefreshCookie(w, resp.RefreshToken, refreshTTL)
		SetAccessCookie(w, resp.AccessToken, accessTTL)
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "google login successful"})
		return nil
	})
}

func YandexLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		state, err := oauthcommon.NewState()
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusInternalServerError, "internal error")
		}

		oauthcommon.SetStateCookie(w, oauthyandex.StateCookieName, state, 10*time.Minute)
		http.Redirect(w, r, cfg.AuthCodeURL(state, oauth2.AccessTypeOnline), http.StatusFound)
		return nil
	})
}

func YandexCallbackHandler(cfg *oauth2.Config, users OAuthUserService, issuer TokenIssuer, accessTTL, refreshTTL time.Duration) http.HandlerFunc {
	return platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		receivedState := strings.TrimSpace(r.URL.Query().Get("state"))
		if receivedState == "" {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "missing oauth state")
		}
		savedState := oauthcommon.ReadStateCookie(r, oauthyandex.StateCookieName)
		if savedState == "" || savedState != receivedState {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "invalid oauth state")
		}
		oauthcommon.ClearStateCookie(w, oauthyandex.StateCookieName)

		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			return platformmiddleware.NewHTTPError(http.StatusBadRequest, "missing oauth code")
		}

		token, err := cfg.Exchange(r.Context(), code)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "invalid oauth code")
		}

		profile, err := oauthyandex.FetchProfile(r.Context(), token.AccessToken)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "failed to fetch yandex profile")
		}

		identity := buildYandexIdentity(profile)
		u, err := users.FindOrCreateFromOAuth(identity)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusInternalServerError, "internal error")
		}

		resp, err := issuer.IssueTokenPair(u)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusInternalServerError, "failed to issue tokens")
		}

		SetRefreshCookie(w, resp.RefreshToken, refreshTTL)
		SetAccessCookie(w, resp.AccessToken, accessTTL)
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "yandex login successful"})
		return nil
	})
}

func VKLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		state, err := oauthcommon.NewState()
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusInternalServerError, "internal error")
		}

		oauthcommon.SetStateCookie(w, oauthvk.StateCookieName, state, 10*time.Minute)
		http.Redirect(w, r, oauthvk.AuthCodeURL(cfg, state), http.StatusFound)
		return nil
	})
}

func VKCallbackHandler(cfg *oauth2.Config, users OAuthUserService, issuer TokenIssuer, accessTTL, refreshTTL time.Duration) http.HandlerFunc {
	return platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		receivedState := strings.TrimSpace(r.URL.Query().Get("state"))
		if receivedState == "" {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "missing oauth state")
		}
		savedState := oauthcommon.ReadStateCookie(r, oauthvk.StateCookieName)
		if savedState == "" || savedState != receivedState {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "invalid oauth state")
		}
		oauthcommon.ClearStateCookie(w, oauthvk.StateCookieName)

		if strings.TrimSpace(r.URL.Query().Get("error")) != "" {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "vk oauth denied")
		}

		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			return platformmiddleware.NewHTTPError(http.StatusBadRequest, "missing oauth code")
		}

		token, err := oauthvk.ExchangeCode(r.Context(), cfg, code)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "invalid oauth code")
		}

		identity := buildVKIdentity(token)
		u, err := users.FindOrCreateFromOAuth(identity)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusInternalServerError, "internal error")
		}

		resp, err := issuer.IssueTokenPair(u)
		if err != nil {
			return platformmiddleware.NewHTTPError(http.StatusInternalServerError, "failed to issue tokens")
		}

		SetRefreshCookie(w, resp.RefreshToken, refreshTTL)
		SetAccessCookie(w, resp.AccessToken, accessTTL)
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"message": "vk login successful"})
		return nil
	})
}

func buildGoogleIdentity(profile oauthgoogle.Profile) authusecase.OAuthIdentity {
	username := strings.TrimSpace(profile.Name)
	if username == "" {
		username = strings.TrimSpace(profile.GivenName + "_" + profile.FamilyName)
	}

	return authusecase.OAuthIdentity{
		Provider:  "google",
		SubjectID: strings.TrimSpace(profile.ID),
		Email:     strings.TrimSpace(profile.Email),
		Username:  username,
	}
}

func buildYandexIdentity(profile oauthyandex.Profile) authusecase.OAuthIdentity {
	return authusecase.OAuthIdentity{
		Provider:  "yandex",
		SubjectID: strings.TrimSpace(profile.ID),
		Email:     strings.TrimSpace(profile.DefaultEmail),
		Username:  strings.TrimSpace(profile.Login),
	}
}

func buildVKIdentity(token *oauth2.Token) authusecase.OAuthIdentity {
	email := extractVKEmail(token)
	userID := extractVKUserID(token)
	username := ""
	if userID != "" {
		username = "vk_" + userID
	}

	return authusecase.OAuthIdentity{
		Provider:  "vk",
		SubjectID: userID,
		Email:     email,
		Username:  username,
	}
}

func extractVKEmail(token *oauth2.Token) string {
	if token == nil {
		return ""
	}
	email, _ := token.Extra("email").(string)
	return strings.TrimSpace(email)
}

func extractVKUserID(token *oauth2.Token) string {
	if token == nil {
		return ""
	}

	raw := token.Extra("user_id")
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == math.Trunc(v) {
			return strconv.FormatInt(int64(v), 10)
		}
		return fmt.Sprintf("%.0f", v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
