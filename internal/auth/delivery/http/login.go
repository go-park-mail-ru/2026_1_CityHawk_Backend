package http

import (
	"encoding/json"
	"errors"
	"net/http"

	platformerrors "cityhawk/backend/internal/platform/errors"
)

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	resp, err := h.authUC.Login(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, platformerrors.ErrInvalidCredentials):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
		case errors.Is(err, platformerrors.ErrIssueTokens):
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to issue tokens"})
		default:
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		}
		return
	}

	SetRefreshCookie(w, resp.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.AccessToken, h.accessTTL)
	writeJSON(w, http.StatusOK, messageResponse{Message: "login successful"})
}
