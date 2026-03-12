package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func authMiddleware(auth *authService, next http.Handler) http.Handler {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		token := readAccessCookie(r)
		if token == "" {
			return errMissingAccess
		}

		c, err := auth.parseToken(token)
		if err != nil || c.Type != "access" {
			return errInvalidAccess
		}
		if strings.TrimSpace(c.Subject) == "" {
			return errInvalidAccess
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, c.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
		return nil
	})
}
