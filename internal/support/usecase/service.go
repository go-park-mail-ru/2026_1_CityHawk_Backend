package usecase

import (
	"context"
	"errors"
	"maps"
	"strings"

	supportmodel "cityhawk/backend/internal/support/model"
	usermodel "cityhawk/backend/internal/user/model"
)

const (
	DefaultTicketsLimit = 20
	MaxTicketsLimit     = 100
)

var (
	ErrValidation         = errors.New("validation failed")
	ErrTicketNotFound     = errors.New("support ticket not found")
	ErrForbidden          = errors.New("forbidden")
	ErrClosedTicketUpdate = errors.New("closed support ticket cannot be edited")
)

type Repository interface {
	Create(ctx context.Context, input CreateInput) (supportmodel.Ticket, error)
	ListByUser(ctx context.Context, userID string, filter supportmodel.Filter) ([]supportmodel.Ticket, error)
	ListAll(ctx context.Context, filter supportmodel.Filter) ([]supportmodel.Ticket, error)
	GetByID(ctx context.Context, id string) (supportmodel.Ticket, error)
	UpdateByUser(ctx context.Context, userID string, ticketID string, input UpdateInput) (supportmodel.Ticket, error)
	UpdateStatus(ctx context.Context, ticketID string, input StatusInput) (supportmodel.Ticket, error)
	CreateMessage(ctx context.Context, input CreateMessageInput) (supportmodel.Message, error)
	ListMessages(ctx context.Context, ticketID string) ([]supportmodel.Message, error)
	Stats(ctx context.Context, filter supportmodel.StatsFilter) (supportmodel.Stats, error)
}

type UserReader interface {
	GetByID(ctx context.Context, id string) (usermodel.User, bool)
}

type CreateInput struct {
	UserID   string
	Category supportmodel.Category
	Title    string
	Message  string
}

type UpdateInput struct {
	Category *supportmodel.Category
	Title    *string
	Message  *string
}

type StatusInput struct {
	Status supportmodel.Status
}

type CreateMessageInput struct {
	TicketID     string
	AuthorUserID string
	AuthorRole   supportmodel.MessageAuthorRole
	Body         string
}

type Service struct {
	repo  Repository
	users UserReader
}

func NewService(repo Repository, users ...UserReader) *Service {
	var userReader UserReader
	if len(users) > 0 {
		userReader = users[0]
	}
	return &Service{repo: repo, users: userReader}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (supportmodel.Ticket, map[string]string, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.Title = strings.TrimSpace(input.Title)
	input.Message = strings.TrimSpace(input.Message)

	details := validateUserID(input.UserID)
	addValidation(details, validateCategory("category", input.Category))
	addValidation(details, validateText("title", input.Title, 3, 200))
	addValidation(details, validateText("message", input.Message, 10, 5000))
	if len(details) > 0 {
		return supportmodel.Ticket{}, details, ErrValidation
	}

	if _, err := s.actor(ctx, input.UserID); err != nil {
		return supportmodel.Ticket{}, nil, err
	}

	ticket, err := s.repo.Create(ctx, input)
	return ticket, nil, err
}

func (s *Service) ListByUser(ctx context.Context, userID string, filter supportmodel.Filter) ([]supportmodel.Ticket, map[string]string, error) {
	userID = strings.TrimSpace(userID)
	details := validateUserID(userID)
	if filter.Status != "" {
		addValidation(details, validateStatus("status", filter.Status))
	}
	if filter.Category != "" {
		addValidation(details, validateCategory("category", filter.Category))
	}
	if filter.Limit <= 0 {
		filter.Limit = DefaultTicketsLimit
	}
	if filter.Limit > MaxTicketsLimit {
		filter.Limit = MaxTicketsLimit
	}
	if filter.Offset < 0 {
		details["offset"] = "offset must be a non-negative integer"
	}
	if len(details) > 0 {
		return nil, details, ErrValidation
	}

	actor, err := s.actor(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	var items []supportmodel.Ticket
	if actor.Role == usermodel.RoleAdmin {
		items, err = s.repo.ListAll(ctx, filter)
	} else {
		items, err = s.repo.ListByUser(ctx, userID, filter)
	}
	return items, nil, err
}

func (s *Service) GetByID(ctx context.Context, userID string, ticketID string) (supportmodel.Ticket, error) {
	userID = strings.TrimSpace(userID)
	ticketID = strings.TrimSpace(ticketID)
	if userID == "" || ticketID == "" {
		return supportmodel.Ticket{}, ErrTicketNotFound
	}

	actor, err := s.actor(ctx, userID)
	if err != nil {
		return supportmodel.Ticket{}, err
	}

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return supportmodel.Ticket{}, err
	}
	if actor.Role != usermodel.RoleAdmin && ticket.UserID != userID {
		return supportmodel.Ticket{}, ErrTicketNotFound
	}
	return ticket, nil
}

func (s *Service) UpdateByUser(ctx context.Context, userID string, ticketID string, input UpdateInput) (supportmodel.Ticket, map[string]string, error) {
	userID = strings.TrimSpace(userID)
	ticketID = strings.TrimSpace(ticketID)
	details := validateUserID(userID)
	if ticketID == "" {
		details["ticketId"] = "ticketId is required"
	}

	if input.Category != nil {
		addValidation(details, validateCategory("category", *input.Category))
	}
	if input.Title != nil {
		value := strings.TrimSpace(*input.Title)
		input.Title = &value
		addValidation(details, validateText("title", value, 3, 200))
	}
	if input.Message != nil {
		value := strings.TrimSpace(*input.Message)
		input.Message = &value
		addValidation(details, validateText("message", value, 10, 5000))
	}
	if input.Category == nil && input.Title == nil && input.Message == nil {
		details["body"] = "at least one field must be provided"
	}
	if len(details) > 0 {
		return supportmodel.Ticket{}, details, ErrValidation
	}

	current, err := s.GetByID(ctx, userID, ticketID)
	if err != nil {
		return supportmodel.Ticket{}, nil, err
	}
	if current.Status == supportmodel.StatusClosed {
		return supportmodel.Ticket{}, nil, ErrClosedTicketUpdate
	}

	ticket, err := s.repo.UpdateByUser(ctx, userID, ticketID, input)
	return ticket, nil, err
}

func (s *Service) UpdateStatus(ctx context.Context, userID string, ticketID string, input StatusInput) (supportmodel.Ticket, map[string]string, error) {
	userID = strings.TrimSpace(userID)
	ticketID = strings.TrimSpace(ticketID)
	details := validateUserID(userID)
	if ticketID == "" {
		details["ticketId"] = "ticketId is required"
	}
	addValidation(details, validateStatus("status", input.Status))
	if len(details) > 0 {
		return supportmodel.Ticket{}, details, ErrValidation
	}

	actor, err := s.actor(ctx, userID)
	if err != nil {
		return supportmodel.Ticket{}, nil, err
	}
	if actor.Role != usermodel.RoleAdmin {
		return supportmodel.Ticket{}, nil, ErrForbidden
	}
	if _, err := s.repo.GetByID(ctx, ticketID); err != nil {
		return supportmodel.Ticket{}, nil, err
	}

	ticket, err := s.repo.UpdateStatus(ctx, ticketID, input)
	return ticket, nil, err
}

func (s *Service) CreateMessage(ctx context.Context, userID string, ticketID string, body string) (supportmodel.Message, map[string]string, error) {
	userID = strings.TrimSpace(userID)
	ticketID = strings.TrimSpace(ticketID)
	body = strings.TrimSpace(body)

	details := validateUserID(userID)
	if ticketID == "" {
		details["ticketId"] = "ticketId is required"
	}
	addValidation(details, validateText("body", body, 1, 5000))
	if len(details) > 0 {
		return supportmodel.Message{}, details, ErrValidation
	}

	actor, err := s.actor(ctx, userID)
	if err != nil {
		return supportmodel.Message{}, nil, err
	}

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return supportmodel.Message{}, nil, err
	}
	if actor.Role != usermodel.RoleAdmin && ticket.UserID != userID {
		return supportmodel.Message{}, nil, ErrTicketNotFound
	}

	message, err := s.repo.CreateMessage(ctx, CreateMessageInput{
		TicketID:     ticketID,
		AuthorUserID: userID,
		AuthorRole:   messageAuthorRole(actor.Role),
		Body:         body,
	})
	return message, nil, err
}

func (s *Service) ListMessages(ctx context.Context, userID string, ticketID string) ([]supportmodel.Message, error) {
	userID = strings.TrimSpace(userID)
	ticketID = strings.TrimSpace(ticketID)
	if userID == "" || ticketID == "" {
		return nil, ErrTicketNotFound
	}

	if _, err := s.GetByID(ctx, userID, ticketID); err != nil {
		return nil, err
	}

	return s.repo.ListMessages(ctx, ticketID)
}

func (s *Service) Stats(ctx context.Context, userID string, filter supportmodel.StatsFilter) (supportmodel.Stats, error) {
	actor, err := s.actor(ctx, strings.TrimSpace(userID))
	if err != nil {
		return supportmodel.Stats{}, err
	}
	if actor.Role != usermodel.RoleAdmin {
		return supportmodel.Stats{}, ErrForbidden
	}
	return s.repo.Stats(ctx, filter)
}

func (s *Service) actor(ctx context.Context, userID string) (usermodel.User, error) {
	if strings.TrimSpace(userID) == "" {
		return usermodel.User{}, ErrTicketNotFound
	}
	if s.users == nil {
		return usermodel.User{ID: userID, Role: usermodel.RoleUser}, nil
	}
	user, ok := s.users.GetByID(ctx, userID)
	if !ok {
		return usermodel.User{}, ErrTicketNotFound
	}
	if user.Role == "" {
		user.Role = usermodel.RoleUser
	}
	return user, nil
}

func messageAuthorRole(role usermodel.Role) supportmodel.MessageAuthorRole {
	if role == usermodel.RoleAdmin {
		return supportmodel.MessageAuthorRoleAdmin
	}
	return supportmodel.MessageAuthorRoleUser
}

func validateUserID(userID string) map[string]string {
	if strings.TrimSpace(userID) == "" {
		return map[string]string{"userId": "userId is required"}
	}
	return map[string]string{}
}

func validateCategory(field string, category supportmodel.Category) map[string]string {
	if !category.Valid() {
		return map[string]string{field: "category must be one of bug, suggestion, product_complaint, other"}
	}
	return nil
}

func validateStatus(field string, status supportmodel.Status) map[string]string {
	if !status.Valid() {
		return map[string]string{field: "status must be one of open, in_progress, closed"}
	}
	return nil
}

func validateText(field string, value string, min int, max int) map[string]string {
	length := len([]rune(strings.TrimSpace(value)))
	if length < min || length > max {
		return map[string]string{field: field + " length is invalid"}
	}
	return nil
}

func addValidation(target map[string]string, details map[string]string) {
	maps.Copy(target, details)
}
