package usecase

import (
	"context"
	"sort"

	placemodel "cityhawk/backend/internal/place/model"
)

const DefaultBestPlacesLimit = 8

type Repository interface {
	ListCards(ctx context.Context) []placemodel.EventCardView
	ListFeaturedEvents(ctx context.Context, limit int) []placemodel.HomeFeaturedEvent
	ListHomeCategories(ctx context.Context, limit int) []placemodel.HomeCategory
	ListHomeCollections(ctx context.Context, limit int) []placemodel.HomeCollection
	GetByID(ctx context.Context, id string) (placemodel.EventDetailsView, bool)
	ListCardsByCategory(ctx context.Context, category string) ([]placemodel.EventCardView, bool)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) HomePayload(ctx context.Context) placemodel.HomePayload {
	return placemodel.HomePayload{
		FeaturedEvents: s.repo.ListFeaturedEvents(ctx, DefaultBestPlacesLimit),
		Categories:     s.repo.ListHomeCategories(ctx, DefaultBestPlacesLimit),
		Collections:    s.repo.ListHomeCollections(ctx, DefaultBestPlacesLimit),
	}
}

func (s *Service) ListCards(ctx context.Context) []placemodel.EventCardView {
	return s.repo.ListCards(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (placemodel.EventDetailsView, bool) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByCategory(ctx context.Context, category string) ([]placemodel.EventCardView, bool) {
	return s.repo.ListCardsByCategory(ctx, category)
}

func (s *Service) Best(ctx context.Context, limit int) []placemodel.EventCardView {
	cards := s.repo.ListCards(ctx)
	if len(cards) == 0 {
		return []placemodel.EventCardView{}
	}

	sort.Slice(cards, func(i, j int) bool {
		return cards[i].LikeCount > cards[j].LikeCount
	})

	if limit <= 0 {
		limit = DefaultBestPlacesLimit
	}
	if len(cards) < limit {
		limit = len(cards)
	}
	return cards[:limit]
}
