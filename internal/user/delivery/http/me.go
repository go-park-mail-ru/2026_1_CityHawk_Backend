package http

import (
	"context"
	"net/http"

	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	usermodel "cityhawk/backend/internal/user/model"
)

type UserReader interface {
	GetByID(ctx context.Context, id string) (usermodel.User, bool)
}

type MeHandler struct {
	users UserReader
}

func NewMeHandler(users UserReader) *MeHandler {
	return &MeHandler{users: users}
}

func (h *MeHandler) Me(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, ok := r.Context().Value(httpx.UserIDContextKey).(string)
		if !ok || userID == "" {
			return httpx.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		u, ok := h.users.GetByID(r.Context(), userID)
		if !ok {
			return httpx.NewHTTPError(http.StatusUnauthorized, "user not found")
		}

		httpx.WriteJSON(w, http.StatusOK, meResponse{
			ID:       u.ID,
			Email:    u.Email,
			Username: u.Username,
		})
		return nil
	}).ServeHTTP(w, r)
}
