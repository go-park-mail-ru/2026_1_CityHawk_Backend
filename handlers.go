package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const refreshCookieName = "refresh_token"

func setRefreshCookie(w http.ResponseWriter, refreshToken string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().UTC().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
	})
}

func clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func readRefreshCookie(r *http.Request) string {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func registerHandler(store *userStore, auth *authService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}

		email, password, username, err := validateRegisterInput(req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		u, err := createUser(email, password, username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		if err := store.create(u); err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}

		resp, err := auth.issueTokenPair(u)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue tokens"})
			return
		}

		setRefreshCookie(w, resp.RefreshToken, auth.refreshTTL)
		writeJSON(w, http.StatusCreated, resp)
	}
}

func loginHandler(store *userStore, auth *authService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}

		email, password, err := validateLoginInput(req)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		u, ok := store.getByEmail(email)
		if !ok || !verifyPassword(password, u.PasswordHash) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}

		resp, err := auth.issueTokenPair(u)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue tokens"})
			return
		}

		setRefreshCookie(w, resp.RefreshToken, auth.refreshTTL)
		writeJSON(w, http.StatusOK, resp)
	}
}

func refreshHandler(store *userStore, auth *authService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken := readRefreshCookie(r)
		if refreshToken == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
			return
		}

		resp, err := auth.rotateRefresh(refreshToken, store)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
			return
		}

		setRefreshCookie(w, resp.RefreshToken, auth.refreshTTL)
		writeJSON(w, http.StatusOK, map[string]string{"access_token": resp.AccessToken})
	}
}

func logoutHandler(auth *authService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken := readRefreshCookie(r)
		if refreshToken == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
			return
		}

		auth.revokeRefresh(refreshToken)
		clearRefreshCookie(w)
		writeJSON(w, http.StatusOK, map[string]string{"message": "logout successful"})
	}
}

func meHandler(store *userStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDContextKey).(string)
		if !ok || userID == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		u, ok := store.getByID(userID)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"id":       u.ID,
			"email":    u.Email,
			"username": u.Username,
		})
	})
}
