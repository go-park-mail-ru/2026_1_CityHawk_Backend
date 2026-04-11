package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	authvalidation "cityhawk/backend/internal/auth/validation"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	usermodel "cityhawk/backend/internal/user/model"
)

type UserReader interface {
	GetByID(ctx context.Context, id string) (usermodel.User, bool)
	UpdateProfile(ctx context.Context, id string, patch usermodel.ProfilePatch) (usermodel.User, bool, error)
}

type MeHandler struct {
	users UserReader
}

func NewMeHandler(users UserReader) *MeHandler {
	return &MeHandler{users: users}
}

func (h *MeHandler) Me(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return h.handleGet(w, r)
		case http.MethodPatch:
			return h.handlePatch(w, r)
		default:
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
	}).ServeHTTP(w, r)
}

func (h *MeHandler) handleGet(w http.ResponseWriter, r *http.Request) error {
	userID, ok := r.Context().Value(httpx.UserIDContextKey).(string)
	if !ok || userID == "" {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	u, ok := h.users.GetByID(r.Context(), userID)
	if !ok {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	httpx.WriteJSON(w, http.StatusOK, makeMeResponse(u))
	return nil
}

func (h *MeHandler) handlePatch(w http.ResponseWriter, r *http.Request) error {
	userID, ok := r.Context().Value(httpx.UserIDContextKey).(string)
	if !ok || userID == "" {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var req patchMeRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse("invalid json", nil))
		return nil
	}

	username, userSurname, birthday, cityID, avatarURL, err := authvalidation.ValidateProfilePatch(
		req.Username,
		req.UserSurname,
		req.Birthday,
		req.CityID,
		req.AvatarURL,
	)
	if err != nil {
		var validationErr authvalidation.ValidationError
		if errors.As(err, &validationErr) {
			httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse("Validation failed", validationErr.Details))
			return nil
		}
		return err
	}

	patch := usermodel.ProfilePatch{}
	if req.Username != nil {
		patch.Username = &username
	}
	if req.UserSurname != nil {
		patch.UserSurname = &userSurname
	}
	if req.Birthday != nil {
		patch.Birthday = birthday
	}
	if req.CityID != nil {
		patch.CityID = &cityID
	}
	if req.AvatarURL != nil {
		patch.AvatarURL = &avatarURL
	}

	u, ok, err := h.users.UpdateProfile(r.Context(), userID, patch)
	if err != nil {
		switch {
		case errors.Is(err, platformerrors.ErrInvalidCity):
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{
				"cityId": "cityId references unknown city",
			})
		default:
			return err
		}
	}
	if !ok {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	httpx.WriteJSON(w, http.StatusOK, makePatchMeResponse(u))
	return nil
}

func makeMeResponse(u usermodel.User) meResponse {
	var birthday *string
	if u.Birthday != nil {
		value := u.Birthday.UTC().Format("2006-01-02")
		birthday = &value
	}

	var city *cityResponse
	if u.City != nil {
		city = &cityResponse{
			ID:          u.City.ID,
			Name:        u.City.Name,
			CountryName: u.City.CountryName,
			Timezone:    u.City.Timezone,
		}
	}

	return meResponse{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		UserSurname: u.UserSurname,
		Birthday:    birthday,
		AvatarURL:   u.AvatarURL,
		City:        city,
		CreatedAt:   u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func makePatchMeResponse(u usermodel.User) patchMeResponse {
	var birthday *string
	if u.Birthday != nil {
		value := u.Birthday.UTC().Format("2006-01-02")
		birthday = &value
	}

	return patchMeResponse{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		UserSurname: u.UserSurname,
		Birthday:    birthday,
		AvatarURL:   u.AvatarURL,
		UpdatedAt:   u.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
