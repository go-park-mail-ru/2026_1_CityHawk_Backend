package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPlacesHandlers(t *testing.T) {
	store := newPlaceStore()

	t.Run("list ok and method not allowed", func(t *testing.T) {
		list := placesListHandler(store)

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
		details := placeDetailsHandler(store)
		req := httptest.NewRequest(http.MethodGet, "/places/vdnh", nil)
		rec := httptest.NewRecorder()
		details.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("details status = %d, want %d", rec.Code, http.StatusOK)
		}
		detailsPayload := decodeJSONMap(t, rec.Body)
		item, ok := detailsPayload["item"].(map[string]any)
		if !ok || item["id"] != "vdnh" {
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

		cat := placesByCategoryListHandler(store)
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

		best := placesBestHandler(store)
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
		if !ok || first["id"] != "zaryadye" {
			t.Fatalf("unexpected first best item: %+v", bestPayload)
		}
	})
}

func TestPlaceStoreAndHelpers(t *testing.T) {
	store := newPlaceStore()

	cards := store.listCards()
	if len(cards) == 0 {
		t.Fatal("listCards returned empty result")
	}

	p, ok := store.getByID("vdnh")
	if !ok || p.ID != "vdnh" {
		t.Fatalf("getByID(vdnh) failed: ok=%v, place=%+v", ok, p)
	}

	parkCards, ok := store.listCardsByCategory("park")
	if !ok || len(parkCards) == 0 {
		t.Fatalf("listCardsByCategory(park) failed: ok=%v len=%d", ok, len(parkCards))
	}

	if containsCategory([]string{"park", "museum"}, "walk") {
		t.Fatal("containsCategory returned true for missing category")
	}
	if !containsCategory([]string{"park", "museum"}, "museum") {
		t.Fatal("containsCategory returned false for existing category")
	}

	got := []string{cards[0].ID, parkCards[0].ID}

	if reflect.DeepEqual(got[0], "") || reflect.DeepEqual(got[1], "") {
		t.Fatalf("unexpected empty ids: %#v", got)
	}
}
