package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	authmodel "cityhawk/backend/internal/auth/model"
	platformcookies "cityhawk/backend/internal/platform/cookies"
	"cityhawk/backend/internal/platform/httpx"
)

const RefreshCookieName = "refresh_token"
const AccessCookieName = "access_token"
const CSRFCookieName = "csrf_token"

type RefreshUsecase interface {
	RotateRefresh(ctx context.Context, oldRefreshToken string) (authmodel.TokenPair, error)
	RevokeRefresh(ctx context.Context, token string) error
}

type AuthFlowUsecase interface {
	Register(ctx context.Context, input authmodel.RegisterInput) (authmodel.RegistrationResult, error)
	Login(ctx context.Context, input authmodel.LoginInput) (authmodel.SessionResult, error)
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

func SetCSRFCookie(w http.ResponseWriter, csrfToken string, ttl time.Duration) {
	platformcookies.Set(w, CSRFCookieName, csrfToken, csrfCookieOptions(ttl))
	w.Header().Set(httpx.CSRFHeader, csrfToken)
}

func ClearCSRFCookie(w http.ResponseWriter) {
	platformcookies.Clear(w, CSRFCookieName, csrfCookieOptions(0))
	w.Header().Del(httpx.CSRFHeader)
}

func ReadCSRFCookie(r *http.Request) string {
	return platformcookies.Read(r, CSRFCookieName)
}

func IssueCSRFToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	var dst [64]byte
	hex.Encode(dst[:], b[:])
	return string(dst[:]), nil
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

func csrfCookieOptions(ttl time.Duration) platformcookies.Options {
	return platformcookies.Options{
		Path:     "/",
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		TTL:      ttl,
	}
}
