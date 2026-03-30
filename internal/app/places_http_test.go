package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	placedelivery "cityhawk/backend/internal/place/delivery/http"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
)

func TestPlacesHandlers(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo))

	t.Run("list ok and method not allowed", func(t *testing.T) {
		list := http.HandlerFunc(handler.List)

		req := httptest.NewRequest(http.MethodGet, "/places", nil)
		rec := httptest.NewRecorder()
		list.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
		}
		payload := decodeJSONMap(t, rec.Body)
		items, ok := payload["items"].([]any)
		if !ok || len(items) == 0 {
			t.Fatalf("list response has no items: %+v", payload)
		}

		reqBad := httptest.NewRequest(http.MethodPost, "/places", nil)
		recBad := httptest.NewRecorder()
		list.ServeHTTP(recBad, reqBad)
		if recBad.Code != http.StatusMethodNotAllowed {
			t.Fatalf("list bad method status = %d, want %d", recBad.Code, http.StatusMethodNotAllowed)
		}
		errPayload := decodeJSONMap(t, recBad.Body)
		if errPayload["error"] != "method not allowed" {
			t.Fatalf("unexpected error response: %+v", errPayload)
		}
	})

	t.Run("details/category/best", func(t *testing.T) {
		details := http.HandlerFunc(handler.Details)
		req := httptest.NewRequest(http.MethodGet, "/places/futurione", nil)
		rec := httptest.NewRecorder()
		details.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("details status = %d, want %d", rec.Code, http.StatusOK)
		}
		detailsPayload := decodeJSONMap(t, rec.Body)
		item, ok := detailsPayload["item"].(map[string]any)
		if !ok || item["id"] != "futurione" {
			t.Fatalf("unexpected details response: %+v", detailsPayload)
		}

		missingReq := httptest.NewRequest(http.MethodGet, "/places/unknown", nil)
		missingRec := httptest.NewRecorder()
		details.ServeHTTP(missingRec, missingReq)
		if missingRec.Code != http.StatusNotFound {
			t.Fatalf("details missing status = %d, want %d", missingRec.Code, http.StatusNotFound)
		}
		missingPayload := decodeJSONMap(t, missingRec.Body)
		if missingPayload["error"] != "place not found" {
			t.Fatalf("unexpected details error: %+v", missingPayload)
		}

		cat := http.HandlerFunc(handler.ByCategory)
		catReq := httptest.NewRequest(http.MethodGet, "/places/category/park", nil)
		catRec := httptest.NewRecorder()
		cat.ServeHTTP(catRec, catReq)
		if catRec.Code != http.StatusOK {
			t.Fatalf("category status = %d, want %d", catRec.Code, http.StatusOK)
		}
		catPayload := decodeJSONMap(t, catRec.Body)
		catItems, ok := catPayload["items"].([]any)
		if !ok || len(catItems) == 0 {
			t.Fatalf("unexpected category response: %+v", catPayload)
		}

		catMissReq := httptest.NewRequest(http.MethodGet, "/places/category/unknown", nil)
		catMissRec := httptest.NewRecorder()
		cat.ServeHTTP(catMissRec, catMissReq)
		if catMissRec.Code != http.StatusNotFound {
			t.Fatalf("category missing status = %d, want %d", catMissRec.Code, http.StatusNotFound)
		}
		catMissPayload := decodeJSONMap(t, catMissRec.Body)
		if catMissPayload["error"] != "category not found" {
			t.Fatalf("unexpected category error: %+v", catMissPayload)
		}

		best := http.HandlerFunc(handler.Best)
		bestReq := httptest.NewRequest(http.MethodGet, "/places/best", nil)
		bestRec := httptest.NewRecorder()
		best.ServeHTTP(bestRec, bestReq)
		if bestRec.Code != http.StatusOK {
			t.Fatalf("best status = %d, want %d", bestRec.Code, http.StatusOK)
		}
		bestPayload := decodeJSONMap(t, bestRec.Body)
		bestItems, ok := bestPayload["items"].([]any)
		if !ok || len(bestItems) == 0 {
			t.Fatalf("unexpected best response: %+v", bestPayload)
		}
		first, ok := bestItems[0].(map[string]any)
		if !ok || first["id"] != "futurione" {
			t.Fatalf("unexpected first best item: %+v", bestPayload)
		}
	})
}

func TestHomeHandlerReturnsHomePayload(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo))
	h := http.HandlerFunc(handler.Home)

	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("home status = %d, want %d", rec.Code, http.StatusOK)
	}

	payload := decodeJSONMap(t, rec.Body)
	places, ok := payload["places"].([]any)
	if !ok || len(places) == 0 {
		t.Fatalf("home payload missing places: %+v", payload)
	}
	moodLeft, ok := payload["moodLeft"].([]any)
	if !ok || len(moodLeft) == 0 {
		t.Fatalf("home payload missing moodLeft: %+v", payload)
	}
	moodTall, ok := payload["moodTall"].(map[string]any)
	if !ok || moodTall["title"] == "" {
		t.Fatalf("home payload missing moodTall: %+v", payload)
	}

	badReq := httptest.NewRequest(http.MethodPost, "/api/home", nil)
	badRec := httptest.NewRecorder()
	h.ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("home bad method status = %d, want %d", badRec.Code, http.StatusMethodNotAllowed)
	}
}

func TestPlaceStoreAndHelpers(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())

	cards := repo.ListCards()
	if len(cards) == 0 {
		t.Fatal("listCards returned empty result")
	}

	p, ok := repo.GetByID("futurione")
	if !ok || p.ID != "futurione" {
		t.Fatalf("getByID(futurione) failed: ok=%v, place=%+v", ok, p)
	}

	parkCards, ok := repo.ListCardsByCategory("park")
	if !ok || len(parkCards) == 0 {
		t.Fatalf("listCardsByCategory(park) failed: ok=%v len=%d", ok, len(parkCards))
	}

	if placerepo.ContainsCategory([]string{"park", "museum"}, "walk") {
		t.Fatal("containsCategory returned true for missing category")
	}
	if !placerepo.ContainsCategory([]string{"park", "museum"}, "museum") {
		t.Fatal("containsCategory returned false for existing category")
	}
	if cards[0].ID == "" || parkCards[0].ID == "" {
		t.Fatalf("unexpected empty ids: cards[0]=%q parkCards[0]=%q", cards[0].ID, parkCards[0].ID)
	}
}
