package usecase

import (
	"context"

	placemodel "cityhawk/backend/internal/place/model"
)

const DefaultEventsLimit = 12
const DefaultHomeLimit = 8

type Repository interface {
	HomePayload(ctx context.Context) placemodel.HomePayload
	ListEvents(ctx context.Context, filter placemodel.EventListFilter) ([]placemodel.EventCardView, int, error)
	GetByID(ctx context.Context, id, userID string) (placemodel.EventDetailsView, bool, error)
	CreateEvent(ctx context.Context, input placemodel.EventWriteInput) (string, error)
	UpdateEvent(ctx context.Context, input placemodel.EventWriteInput) (bool, error)
	DeleteEvent(ctx context.Context, id, userID string) (bool, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) HomePayload(ctx context.Context) placemodel.HomePayload {
	return s.repo.HomePayload(ctx)
}

func (s *Service) ListEvents(ctx context.Context, filter placemodel.EventListFilter) ([]placemodel.EventCardView, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = DefaultEventsLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.ListEvents(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id, userID string) (placemodel.EventDetailsView, bool, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) CreateEvent(ctx context.Context, input placemodel.EventWriteInput) (string, error) {
	return s.repo.CreateEvent(ctx, input)
}

func (s *Service) UpdateEvent(ctx context.Context, input placemodel.EventWriteInput) (bool, error) {
	return s.repo.UpdateEvent(ctx, input)
}

func (s *Service) DeleteEvent(ctx context.Context, id, userID string) (bool, error) {
	return s.repo.DeleteEvent(ctx, id, userID)
}
