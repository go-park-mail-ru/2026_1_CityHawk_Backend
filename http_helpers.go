package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func extractBearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func authMiddleware(auth *authService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r.Header.Get("Authorization"))
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			return
		}

		c, err := auth.parseToken(token)
		if err != nil || c.Type != "access" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid access token"})
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, c.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
