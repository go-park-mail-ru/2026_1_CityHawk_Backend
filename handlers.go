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
const accessCookieName = "access_token"

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

func setAccessCookie(w http.ResponseWriter, accessToken string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    accessToken,
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

func clearAccessCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
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

func readAccessCookie(r *http.Request) string {
	cookie, err := r.Cookie(accessCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return errMethodNotAllowed
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return nil
	}).ServeHTTP(w, r)
}

func registerHandler(store *userStore, auth *authService) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		var req registerRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			return errInvalidJSON
		}

		email, password, username, err := validateRegisterInput(req)
		if err != nil {
			return newHTTPError(http.StatusBadRequest, err.Error())
		}

		u, err := createUser(email, password, username)
		if err != nil {
			return errInternal
		}

		if err := store.create(u); err != nil {
			return err
		}

		resp, err := auth.issueTokenPair(u)
		if err != nil {
			return errIssueTokens
		}

		setRefreshCookie(w, resp.RefreshToken, auth.refreshTTL)
		setAccessCookie(w, resp.AccessToken, auth.accessTTL)
		writeJSON(w, http.StatusCreated, map[string]string{"message": "registration successful"})
		return nil
	})
}

func loginHandler(store *userStore, auth *authService) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		var req loginRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			return errInvalidJSON
		}

		email, password, err := validateLoginInput(req)
		if err != nil {
			return newHTTPError(http.StatusBadRequest, err.Error())
		}

		u, ok := store.getByEmail(email)
		if !ok || !verifyPassword(password, u.PasswordHash) {
			return errInvalidCredentials
		}

		resp, err := auth.issueTokenPair(u)
		if err != nil {
			return errIssueTokens
		}

		setRefreshCookie(w, resp.RefreshToken, auth.refreshTTL)
		setAccessCookie(w, resp.AccessToken, auth.accessTTL)
		writeJSON(w, http.StatusOK, map[string]string{"message": "login successful"})
		return nil
	})
}

func refreshHandler(store *userStore, auth *authService) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		refreshToken := readRefreshCookie(r)
		if refreshToken == "" {
			return errMissingRefresh
		}

		resp, err := auth.rotateRefresh(refreshToken, store)
		if err != nil {
			return errInvalidRefresh
		}

		setRefreshCookie(w, resp.RefreshToken, auth.refreshTTL)
		setAccessCookie(w, resp.AccessToken, auth.accessTTL)
		writeJSON(w, http.StatusOK, map[string]string{"access_token": resp.AccessToken})
		return nil
	})
}

func logoutHandler(auth *authService) http.HandlerFunc {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		refreshToken := readRefreshCookie(r)
		if refreshToken == "" {
			return errMissingRefresh
		}

		auth.revokeRefresh(refreshToken)
		clearRefreshCookie(w)
		clearAccessCookie(w)
		writeJSON(w, http.StatusOK, map[string]string{"message": "logout successful"})
		return nil
	})
}

func meHandler(store *userStore) http.Handler {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, ok := r.Context().Value(userIDContextKey).(string)
		if !ok || userID == "" {
			return errUnauthorized
		}

		u, ok := store.getByID(userID)
		if !ok {
			return errUserNotFound
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"id":       u.ID,
			"email":    u.Email,
			"username": u.Username,
		})
		return nil
	})
}
