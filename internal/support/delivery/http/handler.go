package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	supportmodel "cityhawk/backend/internal/support/model"
	supportusecase "cityhawk/backend/internal/support/usecase"
)

type SupportUsecase interface {
	Create(ctx context.Context, input supportusecase.CreateInput) (supportmodel.Ticket, map[string]string, error)
	ListByUser(ctx context.Context, userID string, filter supportmodel.Filter) ([]supportmodel.Ticket, map[string]string, error)
	GetByID(ctx context.Context, userID string, ticketID string) (supportmodel.Ticket, error)
	UpdateByUser(ctx context.Context, userID string, ticketID string, input supportusecase.UpdateInput) (supportmodel.Ticket, map[string]string, error)
	UpdateStatus(ctx context.Context, userID string, ticketID string, input supportusecase.StatusInput) (supportmodel.Ticket, map[string]string, error)
	CreateMessage(ctx context.Context, userID string, ticketID string, body string) (supportmodel.Message, map[string]string, error)
	ListMessages(ctx context.Context, userID string, ticketID string) ([]supportmodel.Message, error)
	Stats(ctx context.Context, userID string, filter supportmodel.StatsFilter) (supportmodel.Stats, error)
}

type Handler struct {
	support SupportUsecase
}

func NewHandler(support SupportUsecase) *Handler {
	return &Handler{support: support}
}

func (h *Handler) Tickets(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return h.handleList(w, r)
		case http.MethodPost:
			return h.handleCreate(w, r)
		default:
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
	}).ServeHTTP(w, r)
}

func (h *Handler) TicketByID(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return h.handleGetByID(w, r)
		case http.MethodPatch:
			parts, err := parseTicketPath(r.URL.Path)
			if err != nil {
				return err
			}
			if parts.messagesAction {
				return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
			}
			if parts.statusAction {
				return h.handleUpdateStatus(w, r, parts.ticketID)
			}
			return h.handleUpdate(w, r, parts.ticketID)
		case http.MethodPost:
			parts, err := parseTicketPath(r.URL.Path)
			if err != nil {
				return err
			}
			if parts.messagesAction {
				return h.handleCreateMessage(w, r, parts.ticketID)
			}
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		default:
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
	}).ServeHTTP(w, r)
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		filter, err := parseStatsFilter(r)
		if err != nil {
			return err
		}
		stats, err := h.support.Stats(r.Context(), userID, filter)
		if err != nil {
			return mapSupportError(err, nil)
		}
		httpx.WriteJSON(w, http.StatusOK, toTicketStatsResponse(stats))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	var req createTicketRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	ticket, details, err := h.support.Create(r.Context(), supportusecase.CreateInput{
		UserID:   userID,
		Category: supportmodel.Category(req.Category),
		Title:    req.Title,
		Message:  req.Message,
	})
	if err != nil {
		return mapSupportError(err, details)
	}

	httpx.WriteJSON(w, http.StatusCreated, toTicketResponse(ticket))
	return nil
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	filter, err := parseTicketFilter(r)
	if err != nil {
		return err
	}

	items, details, err := h.support.ListByUser(r.Context(), userID, filter)
	if err != nil {
		return mapSupportError(err, details)
	}

	httpx.WriteJSON(w, http.StatusOK, toTicketListResponse(items, filter.Limit, filter.Offset))
	return nil
}

func (h *Handler) handleGetByID(w http.ResponseWriter, r *http.Request) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}
	parts, err := parseTicketPath(r.URL.Path)
	if err != nil {
		return err
	}
	if parts.statusAction {
		return httpx.NewHTTPError(http.StatusNotFound, "Support ticket not found")
	}
	if parts.messagesAction {
		return h.handleListMessages(w, r, parts.ticketID)
	}

	ticket, err := h.support.GetByID(r.Context(), userID, parts.ticketID)
	if err != nil {
		return mapSupportError(err, nil)
	}

	httpx.WriteJSON(w, http.StatusOK, toTicketResponse(ticket))
	return nil
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request, ticketID string) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	var req updateTicketRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	var category *supportmodel.Category
	if req.Category != nil {
		value := supportmodel.Category(*req.Category)
		category = &value
	}

	ticket, details, err := h.support.UpdateByUser(r.Context(), userID, ticketID, supportusecase.UpdateInput{
		Category: category,
		Title:    req.Title,
		Message:  req.Message,
	})
	if err != nil {
		return mapSupportError(err, details)
	}

	httpx.WriteJSON(w, http.StatusOK, toTicketResponse(ticket))
	return nil
}

func (h *Handler) handleUpdateStatus(w http.ResponseWriter, r *http.Request, ticketID string) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	var req updateTicketStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	ticket, details, err := h.support.UpdateStatus(r.Context(), userID, ticketID, supportusecase.StatusInput{
		Status: supportmodel.Status(req.Status),
	})
	if err != nil {
		return mapSupportError(err, details)
	}

	httpx.WriteJSON(w, http.StatusOK, toTicketResponse(ticket))
	return nil
}

func (h *Handler) handleCreateMessage(w http.ResponseWriter, r *http.Request, ticketID string) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	var req createTicketMessageRequest
	if err := decodeJSON(r, &req); err != nil {
		return err
	}

	message, details, err := h.support.CreateMessage(r.Context(), userID, ticketID, req.Body)
	if err != nil {
		return mapSupportError(err, details)
	}

	httpx.WriteJSON(w, http.StatusCreated, toTicketMessageResponse(message))
	return nil
}

func (h *Handler) handleListMessages(w http.ResponseWriter, r *http.Request, ticketID string) error {
	userID, err := requireUserID(r)
	if err != nil {
		return err
	}

	items, err := h.support.ListMessages(r.Context(), userID, ticketID)
	if err != nil {
		return mapSupportError(err, nil)
	}

	httpx.WriteJSON(w, http.StatusOK, toTicketMessageListResponse(items))
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

func parseTicketFilter(r *http.Request) (supportmodel.Filter, error) {
	filter := supportmodel.Filter{
		Status:   supportmodel.Status(strings.TrimSpace(r.URL.Query().Get("status"))),
		Category: supportmodel.Category(strings.TrimSpace(r.URL.Query().Get("category"))),
		Limit:    supportusecase.DefaultTicketsLimit,
		Offset:   0,
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return supportmodel.Filter{}, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"limit": "limit must be a positive integer"})
		}
		filter.Limit = value
	}
	if filter.Limit > supportusecase.MaxTicketsLimit {
		filter.Limit = supportusecase.MaxTicketsLimit
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return supportmodel.Filter{}, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"offset": "offset must be a non-negative integer"})
		}
		filter.Offset = value
	}
	return filter, nil
}

type ticketPath struct {
	ticketID       string
	statusAction   bool
	messagesAction bool
}

func parseTicketPath(path string) (ticketPath, error) {
	rest := strings.TrimPrefix(path, "/api/support/tickets/")
	parts := strings.Split(rest, "/")
	if len(parts) == 1 && strings.TrimSpace(parts[0]) != "" {
		return ticketPath{ticketID: strings.TrimSpace(parts[0])}, nil
	}
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && parts[1] == "status" {
		return ticketPath{ticketID: strings.TrimSpace(parts[0]), statusAction: true}, nil
	}
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" && parts[1] == "messages" {
		return ticketPath{ticketID: strings.TrimSpace(parts[0]), messagesAction: true}, nil
	}
	return ticketPath{}, httpx.NewHTTPError(http.StatusNotFound, "Support ticket not found")
}

func parseStatsFilter(r *http.Request) (supportmodel.StatsFilter, error) {
	var filter supportmodel.StatsFilter
	from, err := parseOptionalTime(r.URL.Query().Get("from"), "from")
	if err != nil {
		return supportmodel.StatsFilter{}, err
	}
	to, err := parseOptionalTime(r.URL.Query().Get("to"), "to")
	if err != nil {
		return supportmodel.StatsFilter{}, err
	}
	if from != nil && to != nil && !from.Before(*to) {
		return supportmodel.StatsFilter{}, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"to": "to must be after from"})
	}
	filter.From = from
	filter.To = to
	return filter, nil
}

func parseOptionalTime(raw string, field string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{field: field + " must be RFC3339 timestamp"})
	}
	return &value, nil
}

func mapSupportError(err error, details map[string]string) error {
	switch {
	case errors.Is(err, supportusecase.ErrValidation):
		return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", details)
	case errors.Is(err, supportusecase.ErrTicketNotFound):
		return httpx.NewHTTPError(http.StatusNotFound, "Support ticket not found")
	case errors.Is(err, supportusecase.ErrClosedTicketUpdate):
		return httpx.NewHTTPError(http.StatusConflict, "Support ticket is closed")
	case errors.Is(err, supportusecase.ErrForbidden):
		return httpx.NewHTTPError(http.StatusForbidden, "Forbidden")
	default:
		return err
	}
}
