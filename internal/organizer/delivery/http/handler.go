package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	organizermodel "cityhawk/backend/internal/organizer/model"
	organizerusecase "cityhawk/backend/internal/organizer/usecase"
	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
)

type OrganizerUsecase interface {
	Create(ctx context.Context, input organizerusecase.CreateInput) (organizermodel.Application, map[string]string, error)
	GetMine(ctx context.Context, userID string) (organizermodel.Application, error)
	UpdateStatus(ctx context.Context, input organizerusecase.StatusInput) (organizermodel.Application, map[string]string, error)
}

type Handler struct {
	organizer OrganizerUsecase
}

func NewHandler(organizer OrganizerUsecase) *Handler {
	return &Handler{organizer: organizer}
}

func (h *Handler) Applications(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodPost:
			return h.handleCreate(w, r)
		default:
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
	}).ServeHTTP(w, r)
}

func (h *Handler) MyApplication(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		application, err := h.organizer.GetMine(r.Context(), userID)
		if err != nil {
			return mapOrganizerError(err, nil)
		}
		httpx.WriteJSON(w, http.StatusOK, toApplicationResponse(application))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) AdminApplicationByID(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPatch {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		applicationID, err := parseAdminApplicationPath(r.URL.Path)
		if err != nil {
			return err
		}
		return h.handleUpdateStatus(w, r, applicationID)
	}).ServeHTTP(w, r)
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	var req createApplicationRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	application, details, err := h.organizer.Create(r.Context(), organizerusecase.CreateInput{
		UserID:      userID,
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		City:        req.City,
		ProjectName: req.ProjectName,
		Categories:  req.Categories,
		Links:       req.Links,
		About:       req.About,
		Consent:     req.Consent,
	})
	if err != nil {
		return mapOrganizerError(err, details)
	}

	httpx.WriteJSON(w, http.StatusCreated, toApplicationCreateResponse(application))
	return nil
}

func (h *Handler) handleUpdateStatus(w http.ResponseWriter, r *http.Request, applicationID string) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	var req updateApplicationStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	application, details, err := h.organizer.UpdateStatus(r.Context(), organizerusecase.StatusInput{
		ActorUserID:   userID,
		ApplicationID: applicationID,
		Status:        organizermodel.ApplicationStatus(req.Status),
		ReviewComment: req.ReviewComment,
	})
	if err != nil {
		return mapOrganizerError(err, details)
	}

	httpx.WriteJSON(w, http.StatusOK, toApplicationStatusResponse(application))
	return nil
}

func requireUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(httpx.UserIDContextKey).(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	return strings.TrimSpace(userID), nil
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return httpx.NewHTTPError(http.StatusBadRequest, "invalid json")
	}
	return nil
}

func parseAdminApplicationPath(path string) (string, error) {
	id := strings.TrimPrefix(path, "/api/admin/organizer/applications/")
	if id == "" || strings.Contains(id, "/") {
		return "", httpx.NewHTTPError(http.StatusNotFound, "Organizer application not found")
	}
	return id, nil
}

func mapOrganizerError(err error, details map[string]string) error {
	switch {
	case errors.Is(err, organizerusecase.ErrValidation):
		return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", details)
	case errors.Is(err, organizerusecase.ErrApplicationNotFound):
		return httpx.NewHTTPError(http.StatusNotFound, "Organizer application not found")
	case errors.Is(err, organizerusecase.ErrActiveApplicationExists):
		return httpx.NewHTTPError(http.StatusConflict, "Active application already exists")
	case errors.Is(err, organizerusecase.ErrForbidden):
		return httpx.NewHTTPError(http.StatusForbidden, "Forbidden")
	default:
		return err
	}
}
