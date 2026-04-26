package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cityhawk/backend/internal/platform/httpx"
	supportmodel "cityhawk/backend/internal/support/model"
	supportusecase "cityhawk/backend/internal/support/usecase"
	usermodel "cityhawk/backend/internal/user/model"
)

func TestHandlerCreateAndList(t *testing.T) {
	uc := newFakeSupportUsecase()
	handler := NewHandler(uc)

	createReq := httptest.NewRequest(http.MethodPost, "/api/support/tickets", strings.NewReader(`{
		"category": "bug",
		"title": "Broken page",
		"message": "Something does not open"
	}`))
	createReq = createReq.WithContext(context.WithValue(createReq.Context(), httpx.UserIDContextKey, "user-1"))
	createRec := httptest.NewRecorder()
	handler.Tickets(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/support/tickets?limit=999", nil)
	listReq = listReq.WithContext(context.WithValue(listReq.Context(), httpx.UserIDContextKey, "user-1"))
	listRec := httptest.NewRecorder()
	handler.Tickets(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d body=%s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	payload := decodeSupportJSONMap(t, listRec)
	if payload["limit"] != float64(supportusecase.MaxTicketsLimit) {
		t.Fatalf("limit was not capped: %+v", payload)
	}
	items, ok := payload["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("unexpected list payload: %+v", payload)
	}
}

func TestHandlerRejectsBadCategory(t *testing.T) {
	handler := NewHandler(newFakeSupportUsecase())

	req := httptest.NewRequest(http.MethodPost, "/api/support/tickets", strings.NewReader(`{
		"category": "billing",
		"title": "Broken page",
		"message": "Something does not open"
	}`))
	req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, "admin-1"))
	rec := httptest.NewRecorder()
	handler.Tickets(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestHandlerUpdateStatus(t *testing.T) {
	uc := newFakeSupportUsecase()
	handler := NewHandler(uc)
	uc.items["ticket-1"] = supportmodel.Ticket{
		ID:        "ticket-1",
		UserID:    "user-1",
		Category:  supportmodel.CategoryBug,
		Status:    supportmodel.StatusOpen,
		Title:     "Broken page",
		Message:   "Something does not open",
		CreatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/support/tickets/ticket-1/status", strings.NewReader(`{
		"status": "closed"
	}`))
	req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, "admin-1"))
	rec := httptest.NewRecorder()
	handler.TicketByID(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	payload := decodeSupportJSONMap(t, rec)
	if payload["status"] != "closed" || payload["closedAt"] == nil {
		t.Fatalf("unexpected response: %+v", payload)
	}
}

func TestHandlerMessages(t *testing.T) {
	uc := newFakeSupportUsecase()
	handler := NewHandler(uc)
	uc.items["ticket-1"] = supportmodel.Ticket{
		ID:        "ticket-1",
		UserID:    "user-1",
		Category:  supportmodel.CategoryBug,
		Status:    supportmodel.StatusOpen,
		Title:     "Broken page",
		Message:   "Something does not open",
		CreatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC),
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/support/tickets/ticket-1/messages", strings.NewReader(`{
		"body": "Please help"
	}`))
	createReq = createReq.WithContext(context.WithValue(createReq.Context(), httpx.UserIDContextKey, "user-1"))
	createRec := httptest.NewRecorder()
	handler.TicketByID(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create message status = %d, want %d body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	createPayload := decodeSupportJSONMap(t, createRec)
	if createPayload["body"] != "Please help" || createPayload["authorRole"] != "user" {
		t.Fatalf("unexpected create message response: %+v", createPayload)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/support/tickets/ticket-1/messages", nil)
	listReq = listReq.WithContext(context.WithValue(listReq.Context(), httpx.UserIDContextKey, "user-1"))
	listRec := httptest.NewRecorder()
	handler.TicketByID(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list messages status = %d, want %d body=%s", listRec.Code, http.StatusOK, listRec.Body.String())
	}
	listPayload := decodeSupportJSONMap(t, listRec)
	items, ok := listPayload["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("unexpected list messages response: %+v", listPayload)
	}
}

func decodeSupportJSONMap(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}

type fakeSupportUsecase struct {
	service *supportusecase.Service
	repo    *fakeSupportRepo
	items   map[string]supportmodel.Ticket
}

func newFakeSupportUsecase() *fakeSupportUsecase {
	repo := &fakeSupportRepo{
		items:    map[string]supportmodel.Ticket{},
		messages: map[string][]supportmodel.Message{},
	}
	return &fakeSupportUsecase{
		service: supportusecase.NewService(repo, fakeSupportUsers{
			"user-1":  {ID: "user-1", Role: usermodel.RoleUser},
			"admin-1": {ID: "admin-1", Role: usermodel.RoleAdmin},
		}),
		repo:  repo,
		items: repo.items,
	}
}

func (f *fakeSupportUsecase) Create(ctx context.Context, input supportusecase.CreateInput) (supportmodel.Ticket, map[string]string, error) {
	return f.service.Create(ctx, input)
}

func (f *fakeSupportUsecase) ListByUser(ctx context.Context, userID string, filter supportmodel.Filter) ([]supportmodel.Ticket, map[string]string, error) {
	return f.service.ListByUser(ctx, userID, filter)
}

func (f *fakeSupportUsecase) GetByID(ctx context.Context, userID string, ticketID string) (supportmodel.Ticket, error) {
	return f.service.GetByID(ctx, userID, ticketID)
}

func (f *fakeSupportUsecase) UpdateByUser(ctx context.Context, userID string, ticketID string, input supportusecase.UpdateInput) (supportmodel.Ticket, map[string]string, error) {
	return f.service.UpdateByUser(ctx, userID, ticketID, input)
}

func (f *fakeSupportUsecase) UpdateStatus(ctx context.Context, userID string, ticketID string, input supportusecase.StatusInput) (supportmodel.Ticket, map[string]string, error) {
	return f.service.UpdateStatus(ctx, userID, ticketID, input)
}

func (f *fakeSupportUsecase) CreateMessage(ctx context.Context, userID string, ticketID string, body string) (supportmodel.Message, map[string]string, error) {
	return f.service.CreateMessage(ctx, userID, ticketID, body)
}

func (f *fakeSupportUsecase) ListMessages(ctx context.Context, userID string, ticketID string) ([]supportmodel.Message, error) {
	return f.service.ListMessages(ctx, userID, ticketID)
}

func (f *fakeSupportUsecase) Stats(ctx context.Context, userID string, filter supportmodel.StatsFilter) (supportmodel.Stats, error) {
	return f.service.Stats(ctx, userID, filter)
}

type fakeSupportRepo struct {
	items    map[string]supportmodel.Ticket
	messages map[string][]supportmodel.Message
}

func (r *fakeSupportRepo) Create(_ context.Context, input supportusecase.CreateInput) (supportmodel.Ticket, error) {
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
	item := supportmodel.Ticket{
		ID:        "ticket-1",
		UserID:    input.UserID,
		Category:  input.Category,
		Status:    supportmodel.StatusOpen,
		Title:     input.Title,
		Message:   input.Message,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.items[item.ID] = item
	return item, nil
}

func (r *fakeSupportRepo) ListByUser(_ context.Context, userID string, _ supportmodel.Filter) ([]supportmodel.Ticket, error) {
	items := make([]supportmodel.Ticket, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeSupportRepo) ListAll(_ context.Context, _ supportmodel.Filter) ([]supportmodel.Ticket, error) {
	items := make([]supportmodel.Ticket, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items, nil
}

func (r *fakeSupportRepo) GetByID(_ context.Context, id string) (supportmodel.Ticket, error) {
	item, ok := r.items[id]
	if !ok {
		return supportmodel.Ticket{}, supportusecase.ErrTicketNotFound
	}
	return item, nil
}

func (r *fakeSupportRepo) UpdateByUser(_ context.Context, userID string, ticketID string, input supportusecase.UpdateInput) (supportmodel.Ticket, error) {
	item, ok := r.items[ticketID]
	if !ok || item.UserID != userID {
		return supportmodel.Ticket{}, supportusecase.ErrTicketNotFound
	}
	if input.Title != nil {
		item.Title = *input.Title
	}
	if input.Message != nil {
		item.Message = *input.Message
	}
	if input.Category != nil {
		item.Category = *input.Category
	}
	r.items[ticketID] = item
	return item, nil
}

func (r *fakeSupportRepo) UpdateStatus(_ context.Context, ticketID string, input supportusecase.StatusInput) (supportmodel.Ticket, error) {
	item, ok := r.items[ticketID]
	if !ok {
		return supportmodel.Ticket{}, supportusecase.ErrTicketNotFound
	}
	item.Status = input.Status
	if input.Status == supportmodel.StatusClosed {
		closedAt := time.Date(2026, 4, 25, 11, 0, 0, 0, time.UTC)
		item.ClosedAt = &closedAt
	} else {
		item.ClosedAt = nil
	}
	r.items[ticketID] = item
	return item, nil
}

func (r *fakeSupportRepo) CreateMessage(_ context.Context, input supportusecase.CreateMessageInput) (supportmodel.Message, error) {
	item := supportmodel.Message{
		ID:           "message-1",
		TicketID:     input.TicketID,
		AuthorUserID: input.AuthorUserID,
		AuthorRole:   input.AuthorRole,
		Body:         input.Body,
		CreatedAt:    time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC),
	}
	r.messages[input.TicketID] = append(r.messages[input.TicketID], item)
	return item, nil
}

func (r *fakeSupportRepo) ListMessages(_ context.Context, ticketID string) ([]supportmodel.Message, error) {
	return append([]supportmodel.Message(nil), r.messages[ticketID]...), nil
}

func (r *fakeSupportRepo) Stats(_ context.Context, _ supportmodel.StatsFilter) (supportmodel.Stats, error) {
	stats := supportmodel.Stats{
		ByStatus:   map[supportmodel.Status]int{},
		ByCategory: map[supportmodel.Category]int{},
	}
	for _, item := range r.items {
		stats.Total++
		stats.ByStatus[item.Status]++
		stats.ByCategory[item.Category]++
	}
	return stats, nil
}

type fakeSupportUsers map[string]usermodel.User

func (f fakeSupportUsers) GetByID(_ context.Context, id string) (usermodel.User, bool) {
	user, ok := f[id]
	if !ok {
		return usermodel.User{}, false
	}
	return user, true
}
