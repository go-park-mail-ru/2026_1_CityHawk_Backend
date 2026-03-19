package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	authmodel "cityhawk/backend/internal/auth/model"
	authusecase "cityhawk/backend/internal/auth/usecase"
)

const RefreshCookieName = "refresh_token"
const AccessCookieName = "access_token"

type RefreshUsecase interface {
	RotateRefresh(oldRefreshToken string) (authmodel.TokenPair, error)
	RevokeRefresh(token string)
}

type AuthHandler struct {
	authUC     authusecase.AuthFlowUsecase
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

func NewAuthHandler(authUC authusecase.AuthFlowUsecase, accessTTL, refreshTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		authUC:     authUC,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func SetRefreshCookie(w http.ResponseWriter, refreshToken string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().UTC().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
	})
}

func SetAccessCookie(w http.ResponseWriter, accessToken string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().UTC().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
	})
}

func ClearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func ClearAccessCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func ReadRefreshCookie(r *http.Request) string {
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func ReadAccessCookie(r *http.Request) string {
	cookie, err := r.Cookie(AccessCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
