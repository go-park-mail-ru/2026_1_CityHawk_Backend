package usecase

import (
	"context"
	"errors"
	"maps"
	"net/mail"
	"strings"

	organizermodel "cityhawk/backend/internal/organizer/model"
	usermodel "cityhawk/backend/internal/user/model"
)

var (
	ErrValidation              = errors.New("validation failed")
	ErrApplicationNotFound     = errors.New("organizer application not found")
	ErrActiveApplicationExists = errors.New("active application already exists")
	ErrForbidden               = errors.New("forbidden")
)

type Repository interface {
	Create(ctx context.Context, input CreateInput) (organizermodel.Application, error)
	GetByUserID(ctx context.Context, userID string) (organizermodel.Application, bool, error)
	UpdateStatus(ctx context.Context, input StatusInput) (organizermodel.Application, bool, error)
}

type UserReader interface {
	GetByID(ctx context.Context, id string) (usermodel.User, bool)
}

type CreateInput struct {
	UserID      string
	Name        string
	Email       string
	Phone       string
	City        string
	ProjectName string
	Categories  string
	Links       string
	About       string
	Consent     bool
}

type StatusInput struct {
	ActorUserID   string
	ApplicationID string
	Status        organizermodel.ApplicationStatus
	ReviewComment string
}

type Service struct {
	repo  Repository
	users UserReader
}

func NewService(repo Repository, users UserReader) *Service {
	return &Service{repo: repo, users: users}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (organizermodel.Application, map[string]string, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.Phone = strings.TrimSpace(input.Phone)
	input.City = strings.TrimSpace(input.City)
	input.ProjectName = strings.TrimSpace(input.ProjectName)
	input.Categories = strings.TrimSpace(input.Categories)
	input.Links = strings.TrimSpace(input.Links)
	input.About = strings.TrimSpace(input.About)

	details := validateUserID(input.UserID)
	addValidation(details, validateText("name", input.Name, 1, 200))
	addValidation(details, validateEmail(input.Email))
	addValidation(details, validateText("phone", input.Phone, 1, 64))
	addValidation(details, validateText("city", input.City, 1, 100))
	addValidation(details, validateText("projectName", input.ProjectName, 1, 200))
	addValidation(details, validateText("categories", input.Categories, 1, 500))
	addValidation(details, validateText("about", input.About, 1, 5000))
	if len([]rune(input.Links)) > 2000 {
		details["links"] = "links length is invalid"
	}
	if !input.Consent {
		details["consent"] = "consent is required"
	}
	if len(details) > 0 {
		return organizermodel.Application{}, details, ErrValidation
	}
	if _, err := s.actor(ctx, input.UserID); err != nil {
		return organizermodel.Application{}, nil, err
	}

	application, err := s.repo.Create(ctx, input)
	return application, nil, err
}

func (s *Service) GetMine(ctx context.Context, userID string) (organizermodel.Application, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return organizermodel.Application{}, ErrApplicationNotFound
	}
	if _, err := s.actor(ctx, userID); err != nil {
		return organizermodel.Application{}, err
	}
	application, ok, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return organizermodel.Application{}, err
	}
	if !ok {
		return organizermodel.Application{}, ErrApplicationNotFound
	}
	return application, nil
}

func (s *Service) UpdateStatus(ctx context.Context, input StatusInput) (organizermodel.Application, map[string]string, error) {
	input.ActorUserID = strings.TrimSpace(input.ActorUserID)
	input.ApplicationID = strings.TrimSpace(input.ApplicationID)
	input.ReviewComment = strings.TrimSpace(input.ReviewComment)

	details := validateUserID(input.ActorUserID)
	if input.ApplicationID == "" {
		details["applicationId"] = "applicationId is required"
	}
	if !input.Status.Valid() {
		details["status"] = "status must be one of pending, needs_info, approved, rejected"
	}
	if len([]rune(input.ReviewComment)) > 2000 {
		details["reviewComment"] = "reviewComment length is invalid"
	}
	if len(details) > 0 {
		return organizermodel.Application{}, details, ErrValidation
	}

	actor, err := s.actor(ctx, input.ActorUserID)
	if err != nil {
		return organizermodel.Application{}, nil, err
	}
	if actor.Role != usermodel.RoleAdmin {
		return organizermodel.Application{}, nil, ErrForbidden
	}

	application, ok, err := s.repo.UpdateStatus(ctx, input)
	if err != nil {
		return organizermodel.Application{}, nil, err
	}
	if !ok {
		return organizermodel.Application{}, nil, ErrApplicationNotFound
	}
	return application, nil, nil
}

func (s *Service) actor(ctx context.Context, userID string) (usermodel.User, error) {
	if strings.TrimSpace(userID) == "" {
		return usermodel.User{}, ErrApplicationNotFound
	}
	if s.users == nil {
		return usermodel.User{ID: userID, Role: usermodel.RoleUser}, nil
	}
	user, ok := s.users.GetByID(ctx, userID)
	if !ok {
		return usermodel.User{}, ErrApplicationNotFound
	}
	if user.Role == "" {
		user.Role = usermodel.RoleUser
	}
	return user, nil
}

func validateUserID(userID string) map[string]string {
	if strings.TrimSpace(userID) == "" {
		return map[string]string{"userId": "userId is required"}
	}
	return map[string]string{}
}

func validateText(field, value string, min, max int) map[string]string {
	length := len([]rune(strings.TrimSpace(value)))
	if length < min || length > max {
		return map[string]string{field: field + " length is invalid"}
	}
	return nil
}

func validateEmail(value string) map[string]string {
	if _, err := mail.ParseAddress(value); err != nil || strings.Contains(value, " ") {
		return map[string]string{"email": "email is invalid"}
	}
	return nil
}

func addValidation(target map[string]string, details map[string]string) {
	maps.Copy(target, details)
}
