package http

import (
	"encoding/json"
	"errors"
	"net/http"

	authmodel "cityhawk/backend/internal/auth/model"
	authvalidation "cityhawk/backend/internal/auth/validation"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
	"cityhawk/backend/internal/platform/media"
	"cityhawk/backend/internal/platform/safety"
)

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse("invalid json", nil))
		return
	}

	resp, err := h.authUC.Register(r.Context(), authmodel.RegisterInput{
		Email:       req.Email,
		Username:    req.Username,
		UserSurname: req.UserSurname,
		Password:    req.Password,
		Birthday:    req.Birthday,
		CityID:      req.CityID,
	})
	if err != nil {
		var validationErr authvalidation.ValidationError
		switch {
		case errors.As(err, &validationErr):
			httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse("Validation failed", validationErr.Details))
		case errors.Is(err, platformerrors.ErrEmailExists):
			httpx.WriteJSON(w, http.StatusConflict, httpx.NewErrorResponse("User already exists", nil))
		case errors.Is(err, platformerrors.ErrIssueTokens):
			httpx.WriteJSON(w, http.StatusInternalServerError, httpx.NewErrorResponse("failed to issue tokens", nil))
		case errors.Is(err, platformerrors.ErrInternal):
			httpx.WriteJSON(w, http.StatusInternalServerError, httpx.NewErrorResponse("internal error", nil))
		default:
			httpx.WriteJSON(w, http.StatusInternalServerError, httpx.NewErrorResponse("internal error", nil))
		}
		return
	}

	SetRefreshCookie(w, resp.Tokens.RefreshToken, h.refreshTTL)
	SetAccessCookie(w, resp.Tokens.AccessToken, h.accessTTL)
	csrfToken, err := IssueCSRFToken()
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, httpx.NewErrorResponse("failed to issue csrf token", nil))
		return
	}
	SetCSRFCookie(w, csrfToken, h.refreshTTL)
	httpx.WriteJSON(w, http.StatusCreated, registerResponse{
		ID:          resp.User.ID,
		Email:       resp.User.Email,
		Username:    safety.EscapeText(resp.User.Username),
		UserSurname: safety.EscapeText(resp.User.UserSurname),
		AvatarURL:   media.PublicURLPtr(resp.User.AvatarURL),
		CreatedAt:   resp.User.CreatedAt.UTC(),
	})
}
