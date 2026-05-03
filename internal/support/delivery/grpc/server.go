package grpc

import (
	"context"
	"errors"

	"cityhawk/backend/internal/grpcconv"
	supportmodel "cityhawk/backend/internal/support/model"
	supportusecase "cityhawk/backend/internal/support/usecase"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
	supportv1 "cityhawk/backend/pkg/pb/support/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

type Server struct {
	supportv1.UnimplementedSupportServiceServer

	support SupportUsecase
}

func NewServer(support SupportUsecase) *Server {
	return &Server{support: support}
}

func (s *Server) CreateTicket(ctx context.Context, req *supportv1.CreateTicketRequest) (*supportv1.Ticket, error) {
	ticket, details, err := s.support.Create(ctx, supportusecase.CreateInput{
		UserID:   grpcconv.UserIDFromContext(req.GetActor()),
		Category: categoryFromProto(req.GetCategory()),
		Title:    req.GetTitle(),
		Message:  req.GetMessage(),
	})
	if err != nil {
		return nil, supportError(err, details)
	}
	return ticketToProto(ticket), nil
}

func (s *Server) ListTickets(ctx context.Context, req *supportv1.ListTicketsRequest) (*supportv1.ListTicketsResponse, error) {
	filter := supportmodel.Filter{
		Status:   statusFromProto(req.GetStatus()),
		Category: categoryFromProto(req.GetCategory()),
		Limit:    int(req.GetPage().GetLimit()),
		Offset:   int(req.GetPage().GetOffset()),
	}
	items, details, err := s.support.ListByUser(ctx, grpcconv.UserIDFromContext(req.GetActor()), filter)
	if err != nil {
		return nil, supportError(err, details)
	}

	response := &supportv1.ListTicketsResponse{
		Items: make([]*supportv1.Ticket, 0, len(items)),
		Page: &commonv1.PageResponse{
			Total:  int32(len(items)),
			Limit:  int32(filter.Limit),
			Offset: int32(filter.Offset),
		},
	}
	for _, item := range items {
		response.Items = append(response.Items, ticketToProto(item))
	}
	return response, nil
}

func (s *Server) GetTicket(ctx context.Context, req *supportv1.GetTicketRequest) (*supportv1.Ticket, error) {
	ticket, err := s.support.GetByID(ctx, grpcconv.UserIDFromContext(req.GetActor()), req.GetTicketId())
	if err != nil {
		return nil, supportError(err, nil)
	}
	return ticketToProto(ticket), nil
}

func (s *Server) UpdateTicket(ctx context.Context, req *supportv1.UpdateTicketRequest) (*supportv1.Ticket, error) {
	var category *supportmodel.Category
	if req.Category != nil {
		value := categoryFromProto(req.GetCategory())
		category = &value
	}

	ticket, details, err := s.support.UpdateByUser(ctx, grpcconv.UserIDFromContext(req.GetActor()), req.GetTicketId(), supportusecase.UpdateInput{
		Category: category,
		Title:    req.Title,
		Message:  req.Message,
	})
	if err != nil {
		return nil, supportError(err, details)
	}
	return ticketToProto(ticket), nil
}

func (s *Server) UpdateTicketStatus(ctx context.Context, req *supportv1.UpdateTicketStatusRequest) (*supportv1.Ticket, error) {
	ticket, details, err := s.support.UpdateStatus(ctx, grpcconv.UserIDFromContext(req.GetActor()), req.GetTicketId(), supportusecase.StatusInput{
		Status: statusFromProto(req.GetStatus()),
	})
	if err != nil {
		return nil, supportError(err, details)
	}
	return ticketToProto(ticket), nil
}

func (s *Server) CreateTicketMessage(ctx context.Context, req *supportv1.CreateTicketMessageRequest) (*supportv1.TicketMessage, error) {
	message, details, err := s.support.CreateMessage(ctx, grpcconv.UserIDFromContext(req.GetActor()), req.GetTicketId(), req.GetBody())
	if err != nil {
		return nil, supportError(err, details)
	}
	return messageToProto(message), nil
}

func (s *Server) ListTicketMessages(ctx context.Context, req *supportv1.ListTicketMessagesRequest) (*supportv1.ListTicketMessagesResponse, error) {
	items, err := s.support.ListMessages(ctx, grpcconv.UserIDFromContext(req.GetActor()), req.GetTicketId())
	if err != nil {
		return nil, supportError(err, nil)
	}

	response := &supportv1.ListTicketMessagesResponse{Items: make([]*supportv1.TicketMessage, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, messageToProto(item))
	}
	return response, nil
}

func (s *Server) GetTicketStats(ctx context.Context, req *supportv1.GetTicketStatsRequest) (*supportv1.TicketStats, error) {
	stats, err := s.support.Stats(ctx, grpcconv.UserIDFromContext(req.GetActor()), supportmodel.StatsFilter{
		From: grpcconv.TimeFromProto(req.GetFrom()),
		To:   grpcconv.TimeFromProto(req.GetTo()),
	})
	if err != nil {
		return nil, supportError(err, nil)
	}
	return statsToProto(stats), nil
}

func ticketToProto(item supportmodel.Ticket) *supportv1.Ticket {
	return &supportv1.Ticket{
		Id:        item.ID,
		UserId:    item.UserID,
		Category:  categoryToProto(item.Category),
		Status:    statusToProto(item.Status),
		Title:     item.Title,
		Message:   item.Message,
		CreatedAt: grpcconv.TimeToProto(item.CreatedAt),
		UpdatedAt: grpcconv.TimeToProto(item.UpdatedAt),
		ClosedAt:  grpcconv.OptionalTimeToProto(item.ClosedAt),
	}
}

func messageToProto(item supportmodel.Message) *supportv1.TicketMessage {
	return &supportv1.TicketMessage{
		Id:           item.ID,
		TicketId:     item.TicketID,
		AuthorUserId: item.AuthorUserID,
		AuthorRole:   authorRoleToProto(item.AuthorRole),
		Body:         item.Body,
		CreatedAt:    grpcconv.TimeToProto(item.CreatedAt),
	}
}

func statsToProto(stats supportmodel.Stats) *supportv1.TicketStats {
	byStatus := make(map[string]int32, len(stats.ByStatus))
	for status, count := range stats.ByStatus {
		byStatus[string(status)] = int32(count)
	}
	byCategory := make(map[string]int32, len(stats.ByCategory))
	for category, count := range stats.ByCategory {
		byCategory[string(category)] = int32(count)
	}
	return &supportv1.TicketStats{
		Total:           int32(stats.Total),
		ByStatus:        byStatus,
		ByCategory:      byCategory,
		OpenTotal:       int32(stats.OpenTotal),
		InProgressTotal: int32(stats.InProgressTotal),
		ClosedTotal:     int32(stats.ClosedTotal),
	}
}

func categoryFromProto(category supportv1.TicketCategory) supportmodel.Category {
	switch category {
	case supportv1.TicketCategory_TICKET_CATEGORY_BUG:
		return supportmodel.CategoryBug
	case supportv1.TicketCategory_TICKET_CATEGORY_SUGGESTION:
		return supportmodel.CategorySuggestion
	case supportv1.TicketCategory_TICKET_CATEGORY_PRODUCT_COMPLAINT:
		return supportmodel.CategoryProductComplaint
	case supportv1.TicketCategory_TICKET_CATEGORY_OTHER:
		return supportmodel.CategoryOther
	default:
		return ""
	}
}

func categoryToProto(category supportmodel.Category) supportv1.TicketCategory {
	switch category {
	case supportmodel.CategoryBug:
		return supportv1.TicketCategory_TICKET_CATEGORY_BUG
	case supportmodel.CategorySuggestion:
		return supportv1.TicketCategory_TICKET_CATEGORY_SUGGESTION
	case supportmodel.CategoryProductComplaint:
		return supportv1.TicketCategory_TICKET_CATEGORY_PRODUCT_COMPLAINT
	case supportmodel.CategoryOther:
		return supportv1.TicketCategory_TICKET_CATEGORY_OTHER
	default:
		return supportv1.TicketCategory_TICKET_CATEGORY_UNSPECIFIED
	}
}

func statusFromProto(ticketStatus supportv1.TicketStatus) supportmodel.Status {
	switch ticketStatus {
	case supportv1.TicketStatus_TICKET_STATUS_OPEN:
		return supportmodel.StatusOpen
	case supportv1.TicketStatus_TICKET_STATUS_IN_PROGRESS:
		return supportmodel.StatusInProgress
	case supportv1.TicketStatus_TICKET_STATUS_CLOSED:
		return supportmodel.StatusClosed
	default:
		return ""
	}
}

func statusToProto(ticketStatus supportmodel.Status) supportv1.TicketStatus {
	switch ticketStatus {
	case supportmodel.StatusOpen:
		return supportv1.TicketStatus_TICKET_STATUS_OPEN
	case supportmodel.StatusInProgress:
		return supportv1.TicketStatus_TICKET_STATUS_IN_PROGRESS
	case supportmodel.StatusClosed:
		return supportv1.TicketStatus_TICKET_STATUS_CLOSED
	default:
		return supportv1.TicketStatus_TICKET_STATUS_UNSPECIFIED
	}
}

func authorRoleToProto(role supportmodel.MessageAuthorRole) supportv1.MessageAuthorRole {
	switch role {
	case supportmodel.MessageAuthorRoleUser:
		return supportv1.MessageAuthorRole_MESSAGE_AUTHOR_ROLE_USER
	case supportmodel.MessageAuthorRoleSupport:
		return supportv1.MessageAuthorRole_MESSAGE_AUTHOR_ROLE_SUPPORT
	case supportmodel.MessageAuthorRoleAdmin:
		return supportv1.MessageAuthorRole_MESSAGE_AUTHOR_ROLE_ADMIN
	default:
		return supportv1.MessageAuthorRole_MESSAGE_AUTHOR_ROLE_UNSPECIFIED
	}
}

func supportError(err error, details map[string]string) error {
	switch {
	case errors.Is(err, supportusecase.ErrValidation):
		return status.Error(codes.InvalidArgument, "validation failed")
	case errors.Is(err, supportusecase.ErrTicketNotFound):
		return status.Error(codes.NotFound, "support ticket not found")
	case errors.Is(err, supportusecase.ErrClosedTicketUpdate):
		return status.Error(codes.FailedPrecondition, "support ticket is closed")
	case errors.Is(err, supportusecase.ErrForbidden):
		return status.Error(codes.PermissionDenied, "forbidden")
	default:
		_ = details
		return grpcconv.Error(err)
	}
}
