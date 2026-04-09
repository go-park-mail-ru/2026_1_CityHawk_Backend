package repository

import (
	"context"
	"sync"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
)

type InMemoryRepository struct {
	mu    sync.RWMutex
	byID  map[string]placemodel.EventDetailsView
	order []string
}

func NewInMemoryRepository(places []placemodel.EventDetailsView) *InMemoryRepository {
	r := &InMemoryRepository{
		byID: make(map[string]placemodel.EventDetailsView),
	}
	for _, p := range places {
		r.byID[p.ID] = p
		r.order = append(r.order, p.ID)
	}
	return r
}

func (r *InMemoryRepository) ListCards(_ context.Context) []placemodel.EventCardView {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cards := make([]placemodel.EventCardView, 0, len(r.order))
	for _, id := range r.order {
		p := r.byID[id]
		cards = append(cards, placemodel.EventCardView{
			ID:                  p.ID,
			Title:               p.Title,
			Categories:          p.Categories,
			LikeCount:           p.LikeCount,
			LocationDescription: p.LocationDescription,
			AddressLine:         p.AddressLine,
			ImageURL:            p.ImageURL,
		})
	}
	return cards
}

func (r *InMemoryRepository) ListFeaturedEvents(_ context.Context, limit int) []placemodel.HomeFeaturedEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 || limit > len(r.order) {
		limit = len(r.order)
	}

	items := make([]placemodel.HomeFeaturedEvent, 0, limit)
	for i := 0; i < limit; i++ {
		p := r.byID[r.order[i]]
		items = append(items, placemodel.HomeFeaturedEvent{
			ID:            p.ID,
			Title:         p.Title,
			CoverImageURL: p.ImageURL,
			Tags:          toHomeTags(p.Categories),
			NextSession: placemodel.HomeNextSession{
				StartAt: time.Date(2026, time.April, 10+i, 19, 0, 0, 0, time.UTC),
				Place: placemodel.HomeNextSessionPlace{
					Name:        p.Title,
					AddressLine: p.AddressLine,
				},
			},
		})
	}

	return items
}

func (r *InMemoryRepository) ListHomeCategories(_ context.Context, limit int) []placemodel.HomeCategory {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]placemodel.EventDetailsView, 0, len(r.order))
	for _, id := range r.order {
		items = append(items, r.byID[id])
	}

	return uniqueSortedCategories(items, limit)
}

func (r *InMemoryRepository) ListHomeCollections(_ context.Context, limit int) []placemodel.HomeCollection {
	items := []placemodel.HomeCollection{
		{
			ID:          "weekend-picks",
			Title:       "Weekend Picks",
			Description: "Best events for weekend",
			ImageURL:    "https://example.com/collection.jpg",
		},
	}

	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

func (r *InMemoryRepository) GetByID(_ context.Context, id string) (placemodel.EventDetailsView, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	return p, ok
}

func (r *InMemoryRepository) ListCardsByCategory(_ context.Context, category string) ([]placemodel.EventCardView, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var cards []placemodel.EventCardView
	for _, id := range r.order {
		p := r.byID[id]
		if ContainsCategory(p.Categories, category) {
			cards = append(cards, placemodel.EventCardView{
				ID:                  p.ID,
				Title:               p.Title,
				Categories:          p.Categories,
				LikeCount:           p.LikeCount,
				LocationDescription: p.LocationDescription,
				AddressLine:         p.AddressLine,
				ImageURL:            p.ImageURL,
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

func toHomeTags(categories []string) []placemodel.HomeTag {
	tags := make([]placemodel.HomeTag, 0, len(categories))
	for _, category := range categories {
		tags = append(tags, placemodel.HomeTag{
			ID:   slugify(category),
			Name: category,
			Slug: slugify(category),
		})
	}
	return tags
}
