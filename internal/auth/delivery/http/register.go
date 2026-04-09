package http

import (
	"encoding/json"
	"errors"
	"net/http"

	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
)

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	resp, err := h.authUC.Register(r.Context(), req.Email, req.Password, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, platformerrors.ErrEmailExists):
			httpx.WriteJSON(w, http.StatusConflict, errorResponse{Error: "email already exists"})
		case errors.Is(err, platformerrors.ErrIssueTokens):
			httpx.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to issue tokens"})
		case errors.Is(err, platformerrors.ErrInternal):
			httpx.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		default:
			httpx.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		}
		return
	}

	SetRefreshCookie(w, resp.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.AccessToken, h.accessTTL)
	httpx.WriteJSON(w, http.StatusCreated, messageResponse{Message: "registration successful"})
}
