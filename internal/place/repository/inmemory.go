package repository

import (
	"context"
	"sync"

	placemodel "cityhawk/backend/internal/place/model"
)

type InMemoryRepository struct {
	mu    sync.RWMutex
	byID  map[string]placemodel.Place
	order []string
}

func NewInMemoryRepository(places []placemodel.Place) *InMemoryRepository {
	r := &InMemoryRepository{
		byID: make(map[string]placemodel.Place),
	}
	for _, p := range places {
		r.byID[p.ID] = p
		r.order = append(r.order, p.ID)
	}
	return r
}

func (r *InMemoryRepository) ListCards(_ context.Context) []placemodel.PlaceCard {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cards := make([]placemodel.PlaceCard, 0, len(r.order))
	for _, id := range r.order {
		p := r.byID[id]
		cards = append(cards, placemodel.PlaceCard{
			ID:               p.ID,
			Title:            p.Title,
			Categories:       p.Categories,
			LikeCount:        p.LikeCount,
			ShortDescription: p.ShortDescription,
			Address:          p.Address,
			ImageURL:         p.ImageURL,
		})
	}
	return cards
}

func (r *InMemoryRepository) ListHomeCards(_ context.Context) []placemodel.HomePlaceCard {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cards := make([]placemodel.HomePlaceCard, 0, len(r.order))
	for _, id := range r.order {
		p := r.byID[id]
		cards = append(cards, placemodel.HomePlaceCard{
			ID:          p.ID,
			ImageURL:    p.ImageURL,
			Title:       p.Title,
			Description: p.ShortDescription,
		})
	}
	return cards
}

func (r *InMemoryRepository) GetByID(_ context.Context, id string) (placemodel.Place, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	return p, ok
}

func (r *InMemoryRepository) ListCardsByCategory(_ context.Context, category string) ([]placemodel.PlaceCard, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var cards []placemodel.PlaceCard
	for _, id := range r.order {
		p := r.byID[id]
		if ContainsCategory(p.Categories, category) {
			cards = append(cards, placemodel.PlaceCard{
				ID:               p.ID,
				Title:            p.Title,
				Categories:       p.Categories,
				LikeCount:        p.LikeCount,
				ShortDescription: p.ShortDescription,
				Address:          p.Address,
				ImageURL:         p.ImageURL,
			})
		}
	}
	return cards, len(cards) > 0
}

func ContainsCategory(categories []string, category string) bool {
	for _, c := range categories {
		if c == category {
			return true
		}
	}
	return false
}
