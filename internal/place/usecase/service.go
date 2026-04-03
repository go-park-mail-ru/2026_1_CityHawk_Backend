package usecase

import (
	"context"
	"sort"

	placemodel "cityhawk/backend/internal/place/model"
)

const DefaultBestPlacesLimit = 8

type Repository interface {
	ListCards(ctx context.Context) []placemodel.PlaceCard
	ListHomeCards(ctx context.Context) []placemodel.HomePlaceCard
	GetByID(ctx context.Context, id string) (placemodel.Place, bool)
	ListCardsByCategory(ctx context.Context, category string) ([]placemodel.PlaceCard, bool)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) HomePayload(ctx context.Context) placemodel.HomePayload {
	return placemodel.HomePayload{
		Places: s.repo.ListHomeCards(ctx),
		MoodLeft: []placemodel.MoodCard{
			{ID: "mood-night", ImageURL: "/public/static/img/club.jpeg", Title: "Ночная жизнь", Modifier: "mood-card--wide"},
			{ID: "mood-eat", ImageURL: "/public/static/img/eat.jpg", Title: "Гастро-места"},
			{ID: "mood-love", ImageURL: "/public/static/img/love.jpg", Title: "Романтический вечер"},
		},
		MoodTall: placemodel.HomeMoodTall{
			ID:       "mood-photo",
			ImageURL: "/public/static/img/photo.jpeg",
			Title:    "Места для фото",
		},
	}
}

func (s *Service) ListCards(ctx context.Context) []placemodel.PlaceCard {
	return s.repo.ListCards(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (placemodel.Place, bool) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByCategory(ctx context.Context, category string) ([]placemodel.PlaceCard, bool) {
	return s.repo.ListCardsByCategory(ctx, category)
}

func (s *Service) Best(ctx context.Context, limit int) []placemodel.PlaceCard {
	cards := s.repo.ListCards(ctx)
	if len(cards) == 0 {
		return []placemodel.PlaceCard{}
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
