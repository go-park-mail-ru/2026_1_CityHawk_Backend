package usecase

import (
	"context"

	placemodel "cityhawk/backend/internal/place/model"
)

const DefaultEventsLimit = 12
const DefaultHomeLimit = 8

type Repository interface {
	HomePayload(ctx context.Context, filter placemodel.HomeFilter) placemodel.HomePayload
	ListCategories(ctx context.Context) []placemodel.HomeCategory
	ListTags(ctx context.Context) []placemodel.HomeTag
	ListCities(ctx context.Context) []placemodel.City
	ListCollections(ctx context.Context) ([]placemodel.CollectionCardView, error)
	GetCollectionByID(ctx context.Context, id string) (placemodel.CollectionDetailsView, bool, error)
	SearchSuggestions(ctx context.Context, query string, limit int) ([]placemodel.SearchSuggestion, error)
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

func (s *Service) HomePayload(ctx context.Context, filter placemodel.HomeFilter) placemodel.HomePayload {
	return s.repo.HomePayload(ctx, filter)
}

func (s *Service) ListCategories(ctx context.Context) []placemodel.HomeCategory {
	return s.repo.ListCategories(ctx)
}

func (s *Service) ListTags(ctx context.Context) []placemodel.HomeTag {
	return s.repo.ListTags(ctx)
}

func (s *Service) ListCities(ctx context.Context) []placemodel.City {
	return s.repo.ListCities(ctx)
}

func (s *Service) ListCollections(ctx context.Context) ([]placemodel.CollectionCardView, error) {
	return s.repo.ListCollections(ctx)
}

func (s *Service) GetCollectionByID(ctx context.Context, id string) (placemodel.CollectionDetailsView, bool, error) {
	return s.repo.GetCollectionByID(ctx, id)
}

func (s *Service) SearchSuggestions(ctx context.Context, query string, limit int) ([]placemodel.SearchSuggestion, error) {
	return s.repo.SearchSuggestions(ctx, query, limit)
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
