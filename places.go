package main

import (
	"net/http"
	"sort"
	"strings"
	"sync"
)

const COUNT_OF_BEST_PLACES = 8

type place struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Categories       []string `json:"categories"`
	LikeCount        int      `json:"like_count"`
	ShortDescription string   `json:"short_description"`
	FullDescription  string   `json:"full_description"`
	Address          string   `json:"address"`
	ImageURL         string   `json:"image_url"`
	WorkingHours     string   `json:"working_hours"`
	PriceLevel       string   `json:"price_level"`
}

type placeCard struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Categories       []string `json:"categories"`
	LikeCount        int      `json:"like_count"`
	ShortDescription string   `json:"short_description"`
	Address          string   `json:"address"`
	ImageURL         string   `json:"image_url"`
}

type placeStore struct {
	mu    sync.RWMutex
	byID  map[string]place
	order []string
}

func (s *placeStore) listCards() []placeCard {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cards := make([]placeCard, 0, len(s.order))
	for _, id := range s.order {
		p := s.byID[id]
		cards = append(cards, placeCard{
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

func (s *placeStore) getByID(id string) (place, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.byID[id]
	return p, ok
}

func (s *placeStore) listCardsByCategory(category string) ([]placeCard, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var cards []placeCard
	for _, id := range s.order {
		p := s.byID[id]
		if containsCategory(p.Categories, category) {
			cards = append(cards, placeCard{
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

func containsCategory(categories []string, category string) bool {
	for _, c := range categories {
		if c == category {
			return true
		}
	}
	return false
}

func placesListHandler(store *placeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": store.listCards()})
	}
}

func placeDetailsHandler(store *placeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/places/")
		if id == "" || strings.Contains(id, "/") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "place not found"})
			return
		}

		p, ok := store.getByID(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "place not found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": p})
	}
}

func placesByCategoryListHandler(store *placeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		category := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/places/category/"))
		if category == "" || strings.Contains(category, "/") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
			return
		}

		p, ok := store.listCardsByCategory(category)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "category not found"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": p})
	}
}

func placesBestHandler(store *placeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}

		cards := store.listCards()
		if len(cards) == 0 {
			writeJSON(w, http.StatusOK, map[string]any{"items": []placeCard{}})
			return
		}

		sort.Slice(cards, func(i, j int) bool {
			return cards[i].LikeCount > cards[j].LikeCount
		})

		limit := COUNT_OF_BEST_PLACES
		if len(cards) < limit {
			limit = len(cards)
		}

		writeJSON(w, http.StatusOK, map[string]any{"items": cards[:limit]})
	}
}
