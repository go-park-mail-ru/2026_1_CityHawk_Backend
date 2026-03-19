package http

import (
	"encoding/json"
	"errors"
	"net/http"

	authusecase "cityhawk/backend/internal/auth/usecase"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	resp, err := h.authUC.Login(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, authusecase.ErrInvalidCredentials):
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		case errors.Is(err, authusecase.ErrIssueTokens):
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue tokens"})
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return
	}

	SetRefreshCookie(w, resp.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.AccessToken, h.accessTTL)
	writeJSON(w, http.StatusOK, map[string]string{"message": "login successful"})
}
