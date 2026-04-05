package http

import (
	"encoding/json"
	"errors"
	"net/http"

	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
)

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	resp, err := h.authUC.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, platformerrors.ErrInvalidCredentials):
			httpx.WriteJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
		case errors.Is(err, platformerrors.ErrIssueTokens):
			httpx.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to issue tokens"})
		default:
			httpx.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		}
		return
	}

	SetRefreshCookie(w, resp.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.AccessToken, h.accessTTL)
	httpx.WriteJSON(w, http.StatusOK, messageResponse{Message: "login successful"})
}
