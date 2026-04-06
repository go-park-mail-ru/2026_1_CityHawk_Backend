package http

import (
	"encoding/json"
	"errors"
	"net/http"

	authvalidation "cityhawk/backend/internal/auth/validation"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
)

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse("invalid json", nil))
		return
	}

	resp, err := h.authUC.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		var validationErr authvalidation.ValidationError
		switch {
		case errors.As(err, &validationErr):
			httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse("Validation failed", validationErr.Details))
		case errors.Is(err, platformerrors.ErrInvalidCredentials):
			httpx.WriteJSON(w, http.StatusUnauthorized, httpx.NewErrorResponse("invalid credentials", nil))
		case errors.Is(err, platformerrors.ErrIssueTokens):
			httpx.WriteJSON(w, http.StatusInternalServerError, httpx.NewErrorResponse("failed to issue tokens", nil))
		default:
			httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse(err.Error(), nil))
		}
		return
	}

	SetRefreshCookie(w, resp.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.AccessToken, h.accessTTL)
	httpx.WriteJSON(w, http.StatusOK, messageResponse{Message: "login successful"})
}
