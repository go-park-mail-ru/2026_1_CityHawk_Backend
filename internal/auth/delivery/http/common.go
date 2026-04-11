package http

import (
	"context"
	"net/http"
	"time"

	authmodel "cityhawk/backend/internal/auth/model"
	platformcookies "cityhawk/backend/internal/platform/cookies"
)

const RefreshCookieName = "refresh_token"
const AccessCookieName = "access_token"

type RefreshUsecase interface {
	RotateRefresh(ctx context.Context, oldRefreshToken string) (authmodel.TokenPair, error)
	RevokeRefresh(ctx context.Context, token string) error
}

type AuthFlowUsecase interface {
	Register(ctx context.Context, email, username, userSurname, password, birthday, cityID string) (authmodel.RegistrationResult, error)
	Login(ctx context.Context, email, password string) (authmodel.SessionResult, error)
}

type AuthHandler struct {
	authUC     AuthFlowUsecase
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type RefreshHandler struct {
	refreshUC  RefreshUsecase
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewRefreshHandler(refreshUC RefreshUsecase, accessTTL, refreshTTL time.Duration) *RefreshHandler {
	return &RefreshHandler{
		refreshUC:  refreshUC,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func NewAuthHandler(authUC AuthFlowUsecase, accessTTL, refreshTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		authUC:     authUC,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func SetRefreshCookie(w http.ResponseWriter, refreshToken string, ttl time.Duration) {
	platformcookies.Set(w, RefreshCookieName, refreshToken, authCookieOptions(ttl))
}

func SetAccessCookie(w http.ResponseWriter, accessToken string, ttl time.Duration) {
	platformcookies.Set(w, AccessCookieName, accessToken, authCookieOptions(ttl))
}

func ClearRefreshCookie(w http.ResponseWriter) {
	platformcookies.Clear(w, RefreshCookieName, authCookieOptions(0))
}

func ClearAccessCookie(w http.ResponseWriter) {
	platformcookies.Clear(w, AccessCookieName, authCookieOptions(0))
}

func ReadRefreshCookie(r *http.Request) string {
	return platformcookies.Read(r, RefreshCookieName)
}

func ReadAccessCookie(r *http.Request) string {
	return platformcookies.Read(r, AccessCookieName)
}

func authCookieOptions(ttl time.Duration) platformcookies.Options {
	return platformcookies.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		TTL:      ttl,
	}
}
