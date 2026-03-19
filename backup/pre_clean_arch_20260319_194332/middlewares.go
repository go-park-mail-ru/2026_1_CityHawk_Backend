package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

type appHandler func(http.ResponseWriter, *http.Request) error

type httpError struct {
	status  int
	message string
}

func (e httpError) Error() string {
	return e.message
}

func newHTTPError(status int, message string) error {
	return httpError{status: status, message: message}
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigin := os.Getenv("FRONTEND_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://cityhawk.ru"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Printf("panic recovered: %v, path=%s, method=%s\n%s", rec, r.URL.Path, r.Method, debug.Stack())
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func errorMiddleware(next appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			writeMappedError(w, err)
		}
	}
}

func writeMappedError(w http.ResponseWriter, err error) {
	var he httpError
	if errors.As(err, &he) {
		writeJSON(w, he.status, map[string]string{"error": he.message})
		return
	}

	switch {
	case errors.Is(err, ErrMethodNotAllowed):
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	case errors.Is(err, ErrInvalidJSON):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
	case errors.Is(err, ErrInternal):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	case errors.Is(err, ErrIssueTokens):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue tokens"})
	case errors.Is(err, ErrInvalidCredentials):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	case errors.Is(err, ErrMissingRefresh):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
	case errors.Is(err, ErrInvalidRefresh):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	case errors.Is(err, ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	case errors.Is(err, ErrUserNotFound):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
	case errors.Is(err, ErrMissingAccess):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing access token"})
	case errors.Is(err, ErrInvalidAccess):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid access token"})
	case errors.Is(err, ErrPlaceNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "place not found"})
	case errors.Is(err, ErrCategoryNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
	case errors.Is(err, ErrEmailExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already exists"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}

func authMiddleware(auth *authService, next http.Handler) http.Handler {
	return errorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		token := readAccessCookie(r)
		if token == "" {
			return ErrMissingAccess
		}

		c, err := auth.parseToken(token)
		if err != nil || c.Type != "access" {
			return ErrInvalidAccess
		}
		if strings.TrimSpace(c.Subject) == "" {
			return ErrInvalidAccess
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, c.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
		return nil
	})
}
