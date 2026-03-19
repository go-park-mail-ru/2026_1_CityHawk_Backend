package http

import (
	"net/http"

	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	usermodel "cityhawk/backend/internal/user/model"
)

type UserReader interface {
	GetByID(id string) (usermodel.User, bool)
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
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		u, ok := h.users.GetByID(userID)
		if !ok {
			return platformmiddleware.NewHTTPError(http.StatusUnauthorized, "user not found")
		}

		httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"id":       u.ID,
			"email":    u.Email,
			"username": u.Username,
		})
		return nil
	}).ServeHTTP(w, r)
}
