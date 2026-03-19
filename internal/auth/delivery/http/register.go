package http

import (
	"encoding/json"
	"errors"
	"net/http"

	authusecase "cityhawk/backend/internal/auth/usecase"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	resp, err := h.authUC.Register(req.Email, req.Password, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, authusecase.ErrEmailExists):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already exists"})
		case errors.Is(err, authusecase.ErrIssueTokens):
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to issue tokens"})
		case errors.Is(err, authusecase.ErrInternal):
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		default:
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return
	}

	SetRefreshCookie(w, resp.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.AccessToken, h.accessTTL)
	writeJSON(w, http.StatusCreated, map[string]string{"message": "registration successful"})
}
