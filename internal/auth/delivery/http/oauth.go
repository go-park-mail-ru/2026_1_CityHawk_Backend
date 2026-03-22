package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	gatewaycommon "cityhawk/backend/internal/auth/gateway/common"
	gatewaygoogle "cityhawk/backend/internal/auth/gateway/google"
	gatewayvk "cityhawk/backend/internal/auth/gateway/vk"
	gatewayyandex "cityhawk/backend/internal/auth/gateway/yandex"
	authmodel "cityhawk/backend/internal/auth/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	"golang.org/x/oauth2"
)

type OAuthLoginUsecase interface {
	LoginWithGoogle(ctx context.Context, code string) (authmodel.TokenPair, error)
	LoginWithYandex(ctx context.Context, code string) (authmodel.TokenPair, error)
	LoginWithVK(ctx context.Context, code string) (authmodel.TokenPair, error)
}

type oauthCallbackFunc func(context.Context, string) (authmodel.TokenPair, error)

type oauthCallbackConfig struct {
	stateCookieName    string
	successMessage     string
	profileFetchErrMsg string
	login              oauthCallbackFunc
	denyQueryError     string
	denyErrorMessage   string
}

func GoogleLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return newOAuthLoginHandler(
		gatewaygoogle.StateCookieName,
		func(state string) string {
			return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
		},
	)
}

func GoogleCallbackHandler(oauthUC OAuthLoginUsecase, accessTTL, refreshTTL time.Duration) http.HandlerFunc {
	return newOAuthCallbackHandler(accessTTL, refreshTTL, oauthCallbackConfig{
		stateCookieName:    gatewaygoogle.StateCookieName,
		successMessage:     "google login successful",
		profileFetchErrMsg: "failed to fetch google profile",
		login:              oauthUC.LoginWithGoogle,
	})
}

func YandexLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return newOAuthLoginHandler(
		gatewayyandex.StateCookieName,
		func(state string) string {
			return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
		},
	)
}

func YandexCallbackHandler(oauthUC OAuthLoginUsecase, accessTTL, refreshTTL time.Duration) http.HandlerFunc {
	return newOAuthCallbackHandler(accessTTL, refreshTTL, oauthCallbackConfig{
		stateCookieName:    gatewayyandex.StateCookieName,
		successMessage:     "yandex login successful",
		profileFetchErrMsg: "failed to fetch yandex profile",
		login:              oauthUC.LoginWithYandex,
	})
}

func VKLoginHandler(cfg *oauth2.Config) http.HandlerFunc {
	return newOAuthLoginHandler(
		gatewayvk.StateCookieName,
		func(state string) string {
			return gatewayvk.AuthCodeURL(cfg, state)
		},
	)
}

func VKCallbackHandler(oauthUC OAuthLoginUsecase, accessTTL, refreshTTL time.Duration) http.HandlerFunc {
	return newOAuthCallbackHandler(accessTTL, refreshTTL, oauthCallbackConfig{
		stateCookieName:  gatewayvk.StateCookieName,
		successMessage:   "vk login successful",
		login:            oauthUC.LoginWithVK,
		denyQueryError:   "error",
		denyErrorMessage: "vk oauth denied",
	})
}

func newOAuthLoginHandler(stateCookieName string, authURL func(string) string) http.HandlerFunc {
	return platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		state, err := gatewaycommon.NewState()
		if err != nil {
			return httpx.NewHTTPError(http.StatusInternalServerError, "internal error")
		}

		gatewaycommon.SetStateCookie(w, stateCookieName, state, 10*time.Minute)
		http.Redirect(w, r, authURL(state), http.StatusFound)
		return nil
	})
}

func newOAuthCallbackHandler(accessTTL, refreshTTL time.Duration, cfg oauthCallbackConfig) http.HandlerFunc {
	return platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		receivedState := strings.TrimSpace(r.URL.Query().Get("state"))
		if receivedState == "" {
			return httpx.NewHTTPError(http.StatusUnauthorized, "missing oauth state")
		}

		savedState := gatewaycommon.ReadStateCookie(r, cfg.stateCookieName)
		if savedState == "" || savedState != receivedState {
			return httpx.NewHTTPError(http.StatusUnauthorized, "invalid oauth state")
		}
		gatewaycommon.ClearStateCookie(w, cfg.stateCookieName)

		if cfg.denyQueryError != "" && strings.TrimSpace(r.URL.Query().Get(cfg.denyQueryError)) != "" {
			return httpx.NewHTTPError(http.StatusUnauthorized, cfg.denyErrorMessage)
		}

		code := strings.TrimSpace(r.URL.Query().Get("code"))
		if code == "" {
			return httpx.NewHTTPError(http.StatusBadRequest, "missing oauth code")
		}

		resp, err := cfg.login(r.Context(), code)
		if err != nil {
			switch {
			case errors.Is(err, platformerrors.ErrInvalidOAuthCode):
				return httpx.NewHTTPError(http.StatusUnauthorized, "invalid oauth code")
			case errors.Is(err, platformerrors.ErrOAuthProfileFetch):
				return httpx.NewHTTPError(http.StatusUnauthorized, cfg.profileFetchErrMsg)
			case errors.Is(err, platformerrors.ErrIssueTokens):
				return httpx.NewHTTPError(http.StatusInternalServerError, "failed to issue tokens")
			default:
				return httpx.NewHTTPError(http.StatusInternalServerError, "internal error")
			}
		}

		SetRefreshCookie(w, resp.RefreshToken, refreshTTL)
		SetAccessCookie(w, resp.AccessToken, accessTTL)
		httpx.WriteJSON(w, http.StatusOK, messageResponse{Message: cfg.successMessage})
		return nil
	})
}
