package main

import (
	"net/http"
	"strings"
	"sync"
)

type place struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Category         string   `json:"category"`
	ShortDescription string   `json:"short_description"`
	FullDescription  string   `json:"full_description"`
	Address          string   `json:"address"`
	ImageURL         string   `json:"image_url"`
	Tags             []string `json:"tags"`
	WorkingHours     string   `json:"working_hours"`
	PriceLevel       string   `json:"price_level"`
}

type placeCard struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Category         string `json:"category"`
	ShortDescription string `json:"short_description"`
	Address          string `json:"address"`
	ImageURL         string `json:"image_url"`
}

type placeStore struct {
	mu    sync.RWMutex
	byID  map[string]place
	order []string
}

func newPlaceStore() *placeStore {
	s := &placeStore{
		byID: make(map[string]place),
	}

	for _, p := range []place{
		{
			ID:               "vdnh",
			Title:            "ВДНХ",
			Category:         "park",
			ShortDescription: "Большой парковый и выставочный комплекс для прогулок и активностей.",
			FullDescription:  "ВДНХ объединяет павильоны, парковые зоны и культурные площадки. Подходит для долгих прогулок, семейных выходных и посещения временных выставок.",
			Address:          "Москва, проспект Мира, 119",
			ImageURL:         "images/vdnh.jpg",
			Tags:             []string{"прогулка", "выставки", "семья"},
			WorkingHours:     "ежедневно, 10:00-22:00",
			PriceLevel:       "free",
		},
		{
			ID:               "garage",
			Title:            "Музей современного искусства Гараж",
			Category:         "museum",
			ShortDescription: "Современное искусство, выставки и лекции в Парке Горького.",
			FullDescription:  "Гараж проводит крупные выставки современных художников, образовательные программы и кинопоказы. Хороший вариант для культурного вечера и знакомства с актуальным искусством.",
			Address:          "Москва, ул. Крымский Вал, 9 стр. 32",
			ImageURL:         "images/garage.jpg",
			Tags:             []string{"искусство", "выставки", "лекции"},
			WorkingHours:     "вт-вс, 11:00-22:00",
			PriceLevel:       "medium",
		},
		{
			ID:               "zaryadye",
			Title:            "Парк Зарядье",
			Category:         "park",
			ShortDescription: "Центральный парк с панорамным мостом и видами на Кремль.",
			FullDescription:  "Зарядье расположен рядом с Красной площадью и известен парящим мостом, ландшафтными зонами и событийной программой. Подходит для короткой прогулки в центре города.",
			Address:          "Москва, ул. Варварка, 6",
			ImageURL:         "images/zaryadye.jpg",
			Tags:             []string{"центр", "панорама", "прогулка"},
			WorkingHours:     "ежедневно, 10:00-21:00",
			PriceLevel:       "free",
		},
	} {
		s.byID[p.ID] = p
		s.order = append(s.order, p.ID)
	}

	return s
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
			Category:         p.Category,
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
		id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/places/"))
		if id == "" || strings.Contains(id, "/") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "place not found"})
			return
		}

		p, ok := store.getByID(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "place not found"})
			return
		}

		writeJSON(w, http.StatusOK, p)
	}
}
