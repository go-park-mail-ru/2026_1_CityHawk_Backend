package usecase

import (
	"sort"

	placemodel "cityhawk/backend/internal/place/model"
)

const DefaultBestPlacesLimit = 8

type Repository interface {
	ListCards() []placemodel.PlaceCard
	ListHomeCards() []placemodel.HomePlaceCard
	GetByID(id string) (placemodel.Place, bool)
	ListCardsByCategory(category string) ([]placemodel.PlaceCard, bool)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) HomePayload() placemodel.HomePayload {
	return placemodel.HomePayload{
		Places: s.repo.ListHomeCards(),
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

func (s *Service) ListCards() []placemodel.PlaceCard {
	return s.repo.ListCards()
}

func (s *Service) GetByID(id string) (placemodel.Place, bool) {
	return s.repo.GetByID(id)
}

func (s *Service) ListByCategory(category string) ([]placemodel.PlaceCard, bool) {
	return s.repo.ListCardsByCategory(category)
}

func (s *Service) Best(limit int) []placemodel.PlaceCard {
	cards := s.repo.ListCards()
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
