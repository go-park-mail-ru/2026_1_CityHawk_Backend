package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func registerHandler(store *userStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}

		email := strings.TrimSpace(strings.ToLower(req.Email))
		password := strings.TrimSpace(req.Password)
		if email == "" || password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
			return
		}

		u, err := createUser(email, password)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		if err := store.create(u); err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]string{
			"id":    u.ID,
			"email": u.Email,
		})
	}
}

func loginHandler(store *userStore, auth *authService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req authRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}

		email := strings.TrimSpace(strings.ToLower(req.Email))
		password := strings.TrimSpace(req.Password)
		if email == "" || password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password are required"})
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

		writeJSON(w, http.StatusOK, resp)
	}
}

func refreshHandler(auth *authService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}

		refreshToken := strings.TrimSpace(req.RefreshToken)
		if refreshToken == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
			return
		}

		resp, err := auth.rotateRefresh(refreshToken)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func logoutHandler(auth *authService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req refreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}

		refreshToken := strings.TrimSpace(req.RefreshToken)
		if refreshToken == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "refresh_token is required"})
			return
		}

		auth.revokeRefresh(refreshToken)
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
			"id":    u.ID,
			"email": u.Email,
		})
	})
}
