package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	supportmodel "cityhawk/backend/internal/support/model"
	usermodel "cityhawk/backend/internal/user/model"
)

func TestServiceCreateValidatesAndTrims(t *testing.T) {
	repo := newFakeRepository()
	service := NewService(repo)

	ticket, details, err := service.Create(context.Background(), CreateInput{
		UserID:   " user-1 ",
		Category: supportmodel.CategoryBug,
		Title:    " Broken page ",
		Message:  " Something does not open ",
	})
	if err != nil {
		t.Fatalf("Create err = %v details=%v", err, details)
	}
	if ticket.UserID != "user-1" || ticket.Title != "Broken page" || ticket.Message != "Something does not open" {
		t.Fatalf("ticket was not normalized: %+v", ticket)
	}
	if ticket.Status != supportmodel.StatusOpen {
		t.Fatalf("status = %q, want open", ticket.Status)
	}
}

func TestServiceCreateRejectsUnknownCategory(t *testing.T) {
	service := NewService(newFakeRepository())

	_, details, err := service.Create(context.Background(), CreateInput{
		UserID:   "user-1",
		Category: "billing",
		Title:    "Broken page",
		Message:  "Something does not open",
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("err = %v, want validation", err)
	}
	if details["category"] == "" {
		t.Fatalf("missing category detail: %+v", details)
	}
}

func TestServiceGetByIDHidesForeignTicket(t *testing.T) {
	repo := newFakeRepository()
	repo.items["ticket-1"] = supportmodel.Ticket{ID: "ticket-1", UserID: "user-2", Status: supportmodel.StatusOpen}
	service := NewService(repo)

	_, err := service.GetByID(context.Background(), "user-1", "ticket-1")
	if !errors.Is(err, ErrTicketNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestServiceUpdateRejectsClosedTicket(t *testing.T) {
	repo := newFakeRepository()
	repo.items["ticket-1"] = supportmodel.Ticket{
		ID:       "ticket-1",
		UserID:   "user-1",
		Status:   supportmodel.StatusClosed,
		Title:    "Broken page",
		Message:  "Something does not open",
		Category: supportmodel.CategoryBug,
	}
	service := NewService(repo)

	title := "Updated title"
	_, _, err := service.UpdateByUser(context.Background(), "user-1", "ticket-1", UpdateInput{Title: &title})
	if !errors.Is(err, ErrClosedTicketUpdate) {
		t.Fatalf("err = %v, want closed ticket update", err)
	}
}

func TestServiceUpdateStatusClosesAndReopens(t *testing.T) {
	repo := newFakeRepository()
	repo.items["ticket-1"] = supportmodel.Ticket{
		ID:       "ticket-1",
		UserID:   "user-1",
		Status:   supportmodel.StatusOpen,
		Title:    "Broken page",
		Message:  "Something does not open",
		Category: supportmodel.CategoryBug,
	}
	service := NewService(repo, fakeUsers{
		"admin-1": {ID: "admin-1", Role: usermodel.RoleAdmin},
	})

	closed, details, err := service.UpdateStatus(context.Background(), "admin-1", "ticket-1", StatusInput{
		Status: supportmodel.StatusClosed,
	})
	if err != nil {
		t.Fatalf("UpdateStatus close err = %v details=%v", err, details)
	}
	if closed.Status != supportmodel.StatusClosed || closed.ClosedAt == nil {
		t.Fatalf("ticket was not closed correctly: %+v", closed)
	}

	reopened, details, err := service.UpdateStatus(context.Background(), "admin-1", "ticket-1", StatusInput{Status: supportmodel.StatusOpen})
	if err != nil {
		t.Fatalf("UpdateStatus reopen err = %v details=%v", err, details)
	}
	if reopened.Status != supportmodel.StatusOpen || reopened.ClosedAt != nil {
		t.Fatalf("ticket was not reopened correctly: %+v", reopened)
	}
}

func TestServiceMessages(t *testing.T) {
	repo := newFakeRepository()
	repo.items["ticket-1"] = supportmodel.Ticket{
		ID:       "ticket-1",
		UserID:   "user-1",
		Status:   supportmodel.StatusOpen,
		Title:    "Broken page",
		Message:  "Something does not open",
		Category: supportmodel.CategoryBug,
	}
	service := NewService(repo)

	message, details, err := service.CreateMessage(context.Background(), "user-1", "ticket-1", " Please help ")
	if err != nil {
		t.Fatalf("CreateMessage err = %v details=%v", err, details)
	}
	if message.Body != "Please help" || message.AuthorUserID != "user-1" || message.AuthorRole != supportmodel.MessageAuthorRoleUser {
		t.Fatalf("unexpected message: %+v", message)
	}

	items, err := service.ListMessages(context.Background(), "user-1", "ticket-1")
	if err != nil {
		t.Fatalf("ListMessages err = %v", err)
	}
	if len(items) != 1 || items[0].Body != "Please help" {
		t.Fatalf("unexpected messages: %+v", items)
	}
}

type fakeRepository struct {
	items    map[string]supportmodel.Ticket
	messages map[string][]supportmodel.Message
	next     int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		items:    map[string]supportmodel.Ticket{},
		messages: map[string][]supportmodel.Message{},
	}
}

func (r *fakeRepository) Create(_ context.Context, input CreateInput) (supportmodel.Ticket, error) {
	r.next++
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

func (r *fakeRepository) ListByUser(_ context.Context, userID string, _ supportmodel.Filter) ([]supportmodel.Ticket, error) {
	items := make([]supportmodel.Ticket, 0)
	for _, item := range r.items {
		if item.UserID == userID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeRepository) ListAll(_ context.Context, _ supportmodel.Filter) ([]supportmodel.Ticket, error) {
	items := make([]supportmodel.Ticket, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items, nil
}

func (r *fakeRepository) GetByID(_ context.Context, id string) (supportmodel.Ticket, error) {
	item, ok := r.items[id]
	if !ok {
		return supportmodel.Ticket{}, ErrTicketNotFound
	}
	return item, nil
}

func (r *fakeRepository) UpdateByUser(_ context.Context, userID string, ticketID string, input UpdateInput) (supportmodel.Ticket, error) {
	item, ok := r.items[ticketID]
	if !ok || item.UserID != userID {
		return supportmodel.Ticket{}, ErrTicketNotFound
	}
	if input.Category != nil {
		item.Category = *input.Category
	}
	if input.Title != nil {
		item.Title = *input.Title
	}
	if input.Message != nil {
		item.Message = *input.Message
	}
	r.items[ticketID] = item
	return item, nil
}

func (r *fakeRepository) UpdateStatus(_ context.Context, ticketID string, input StatusInput) (supportmodel.Ticket, error) {
	item, ok := r.items[ticketID]
	if !ok {
		return supportmodel.Ticket{}, ErrTicketNotFound
	}
	item.Status = input.Status
	if input.Status == supportmodel.StatusClosed {
		now := time.Date(2026, 4, 25, 11, 0, 0, 0, time.UTC)
		item.ClosedAt = &now
	} else {
		item.ClosedAt = nil
	}
	r.items[ticketID] = item
	return item, nil
}

func (r *fakeRepository) CreateMessage(_ context.Context, input CreateMessageInput) (supportmodel.Message, error) {
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

func (r *fakeRepository) ListMessages(_ context.Context, ticketID string) ([]supportmodel.Message, error) {
	return append([]supportmodel.Message(nil), r.messages[ticketID]...), nil
}

func (r *fakeRepository) Stats(_ context.Context, _ supportmodel.StatsFilter) (supportmodel.Stats, error) {
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

type fakeUsers map[string]usermodel.User

func (f fakeUsers) GetByID(_ context.Context, id string) (usermodel.User, bool) {
	user, ok := f[id]
	if !ok {
		return usermodel.User{}, false
	}
	return user, true
}
