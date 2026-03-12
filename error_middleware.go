package main

import (
	"errors"
	"net/http"
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

var (
	errMethodNotAllowed   = errors.New("method not allowed")
	errInvalidJSON        = errors.New("invalid json")
	errInternal           = errors.New("internal error")
	errIssueTokens        = errors.New("failed to issue tokens")
	errInvalidCredentials = errors.New("invalid credentials")
	errMissingRefresh     = errors.New("missing refresh token")
	errInvalidRefresh     = errors.New("invalid refresh token")
	errUnauthorized       = errors.New("unauthorized")
	errUserNotFound       = errors.New("user not found")
	errMissingAccess      = errors.New("missing access token")
	errInvalidAccess      = errors.New("invalid access token")
	errPlaceNotFound      = errors.New("place not found")
	errCategoryNotFound   = errors.New("category not found")
	errEmailExists        = errors.New("email already exists")
)

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
	case errors.Is(err, errMethodNotAllowed):
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	case errors.Is(err, errInvalidJSON):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
	case errors.Is(err, errInternal):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	case errors.Is(err, errIssueTokens):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue tokens"})
	case errors.Is(err, errInvalidCredentials):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	case errors.Is(err, errMissingRefresh):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing refresh token"})
	case errors.Is(err, errInvalidRefresh):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
	case errors.Is(err, errUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	case errors.Is(err, errUserNotFound):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "user not found"})
	case errors.Is(err, errMissingAccess):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing access token"})
	case errors.Is(err, errInvalidAccess):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid access token"})
	case errors.Is(err, errPlaceNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "place not found"})
	case errors.Is(err, errCategoryNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
	case errors.Is(err, errEmailExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email already exists"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}
