package grpc

import (
	"context"
	"testing"
	"time"

	supportmodel "cityhawk/backend/internal/support/model"
	supportusecase "cityhawk/backend/internal/support/usecase"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
	supportv1 "cityhawk/backend/pkg/pb/support/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServerSupportFlow(t *testing.T) {
	uc := &fakeSupportUsecase{now: time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC)}
	server := NewServer(uc)
	actor := &commonv1.UserContext{UserId: "user-1", Authenticated: true}

	ticket, err := server.CreateTicket(context.Background(), &supportv1.CreateTicketRequest{
		Actor:    actor,
		Category: supportv1.TicketCategory_TICKET_CATEGORY_BUG,
		Title:    "Bug",
		Message:  "Something is broken",
	})
	if err != nil || ticket.GetId() == "" || ticket.GetCategory() != supportv1.TicketCategory_TICKET_CATEGORY_BUG {
		t.Fatalf("CreateTicket() = (%+v, %v)", ticket, err)
	}

	list, err := server.ListTickets(context.Background(), &supportv1.ListTicketsRequest{
		Actor: actor,
		Page:  &commonv1.PageRequest{Limit: 10, Offset: 2},
	})
	if err != nil || len(list.GetItems()) != 1 || list.GetPage().GetLimit() != 10 {
		t.Fatalf("ListTickets() = (%+v, %v)", list, err)
	}

	got, err := server.GetTicket(context.Background(), &supportv1.GetTicketRequest{Actor: actor, TicketId: "ticket-1"})
	if err != nil || got.GetId() != "ticket-1" {
		t.Fatalf("GetTicket() = (%+v, %v)", got, err)
	}

	newTitle := "New title"
	updated, err := server.UpdateTicket(context.Background(), &supportv1.UpdateTicketRequest{
		Actor:    actor,
		TicketId: "ticket-1",
		Title:    &newTitle,
	})
	if err != nil || updated.GetTitle() != newTitle {
		t.Fatalf("UpdateTicket() = (%+v, %v)", updated, err)
	}

	closed, err := server.UpdateTicketStatus(context.Background(), &supportv1.UpdateTicketStatusRequest{
		Actor:    actor,
		TicketId: "ticket-1",
		Status:   supportv1.TicketStatus_TICKET_STATUS_CLOSED,
	})
	if err != nil || closed.GetStatus() != supportv1.TicketStatus_TICKET_STATUS_CLOSED {
		t.Fatalf("UpdateTicketStatus() = (%+v, %v)", closed, err)
	}

	message, err := server.CreateTicketMessage(context.Background(), &supportv1.CreateTicketMessageRequest{
		Actor:    actor,
		TicketId: "ticket-1",
		Body:     "Hello",
	})
	if err != nil || message.GetAuthorRole() != supportv1.MessageAuthorRole_MESSAGE_AUTHOR_ROLE_USER {
		t.Fatalf("CreateTicketMessage() = (%+v, %v)", message, err)
	}

	messages, err := server.ListTicketMessages(context.Background(), &supportv1.ListTicketMessagesRequest{Actor: actor, TicketId: "ticket-1"})
	if err != nil || len(messages.GetItems()) != 1 {
		t.Fatalf("ListTicketMessages() = (%+v, %v)", messages, err)
	}

	stats, err := server.GetTicketStats(context.Background(), &supportv1.GetTicketStatsRequest{
		Actor: actor,
		From:  timestamppb.New(uc.now.Add(-time.Hour)),
		To:    timestamppb.New(uc.now.Add(time.Hour)),
	})
	if err != nil || stats.GetTotal() != 1 || stats.GetClosedTotal() != 1 {
		t.Fatalf("GetTicketStats() = (%+v, %v)", stats, err)
	}
}

func TestSupportErrorMapsUsecaseErrors(t *testing.T) {
	for _, err := range []error{
		supportusecase.ErrValidation,
		supportusecase.ErrTicketNotFound,
		supportusecase.ErrClosedTicketUpdate,
		supportusecase.ErrForbidden,
	} {
		if got := supportError(err, nil); got == nil {
			t.Fatalf("supportError(%v) = nil", err)
		}
	}
}

type fakeSupportUsecase struct {
	now time.Time
}

func (f *fakeSupportUsecase) ticket() supportmodel.Ticket {
	closedAt := f.now.Add(time.Hour)
	return supportmodel.Ticket{
		ID:        "ticket-1",
		UserID:    "user-1",
		Category:  supportmodel.CategoryBug,
		Status:    supportmodel.StatusClosed,
		Title:     "Bug",
		Message:   "Something is broken",
		CreatedAt: f.now,
		UpdatedAt: f.now,
		ClosedAt:  &closedAt,
	}
}

func (f *fakeSupportUsecase) Create(context.Context, supportusecase.CreateInput) (supportmodel.Ticket, map[string]string, error) {
	return f.ticket(), nil, nil
}

func (f *fakeSupportUsecase) ListByUser(context.Context, string, supportmodel.Filter) ([]supportmodel.Ticket, map[string]string, error) {
	return []supportmodel.Ticket{f.ticket()}, nil, nil
}

func (f *fakeSupportUsecase) GetByID(context.Context, string, string) (supportmodel.Ticket, error) {
	return f.ticket(), nil
}

func (f *fakeSupportUsecase) UpdateByUser(_ context.Context, _ string, _ string, input supportusecase.UpdateInput) (supportmodel.Ticket, map[string]string, error) {
	ticket := f.ticket()
	if input.Title != nil {
		ticket.Title = *input.Title
	}
	return ticket, nil, nil
}

func (f *fakeSupportUsecase) UpdateStatus(_ context.Context, _ string, _ string, input supportusecase.StatusInput) (supportmodel.Ticket, map[string]string, error) {
	ticket := f.ticket()
	ticket.Status = input.Status
	return ticket, nil, nil
}

func (f *fakeSupportUsecase) CreateMessage(context.Context, string, string, string) (supportmodel.Message, map[string]string, error) {
	return supportmodel.Message{
		ID:           "message-1",
		TicketID:     "ticket-1",
		AuthorUserID: "user-1",
		AuthorRole:   supportmodel.MessageAuthorRoleUser,
		Body:         "Hello",
		CreatedAt:    f.now,
	}, nil, nil
}

func (f *fakeSupportUsecase) ListMessages(context.Context, string, string) ([]supportmodel.Message, error) {
	message, _, _ := f.CreateMessage(context.Background(), "user-1", "ticket-1", "Hello")
	return []supportmodel.Message{message}, nil
}

func (f *fakeSupportUsecase) Stats(context.Context, string, supportmodel.StatsFilter) (supportmodel.Stats, error) {
	return supportmodel.Stats{
		Total:       1,
		ClosedTotal: 1,
		ByStatus:    map[supportmodel.Status]int{supportmodel.StatusClosed: 1},
		ByCategory:  map[supportmodel.Category]int{supportmodel.CategoryBug: 1},
	}, nil
}
