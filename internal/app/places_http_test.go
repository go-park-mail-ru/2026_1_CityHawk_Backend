package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	placedelivery "cityhawk/backend/internal/place/delivery/http"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
	"cityhawk/backend/internal/platform/httpx"
)

func TestEventsHandlers(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo))

	t.Run("list ok and method not allowed", func(t *testing.T) {
		h := http.HandlerFunc(handler.Events)

		req := httptest.NewRequest(http.MethodGet, "/api/events?limit=12&offset=0", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("list status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		items, ok := payload["items"].([]any)
		if !ok || len(items) == 0 {
			t.Fatalf("events response has no items: %+v", payload)
		}
		if payload["total"] == nil || payload["limit"] != float64(12) || payload["offset"] != float64(0) {
			t.Fatalf("unexpected list pagination: %+v", payload)
		}

		reqBad := httptest.NewRequest(http.MethodPut, "/api/events", nil)
		recBad := httptest.NewRecorder()
		h.ServeHTTP(recBad, reqBad)
		if recBad.Code != http.StatusMethodNotAllowed {
			t.Fatalf("bad method status = %d, want %d", recBad.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("details get", func(t *testing.T) {
		h := http.HandlerFunc(handler.EventByID)

		req := httptest.NewRequest(http.MethodGet, "/api/events/futurione", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("details status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		if payload["id"] != "futurione" {
			t.Fatalf("unexpected details response: %+v", payload)
		}
		if _, ok := payload["sessions"].([]any); !ok {
			t.Fatalf("missing sessions in details: %+v", payload)
		}

		missingReq := httptest.NewRequest(http.MethodGet, "/api/events/unknown", nil)
		missingRec := httptest.NewRecorder()
		h.ServeHTTP(missingRec, missingReq)
		if missingRec.Code != http.StatusNotFound {
			t.Fatalf("details missing status = %d, want %d", missingRec.Code, http.StatusNotFound)
		}
	})

	t.Run("create patch delete", func(t *testing.T) {
		events := http.HandlerFunc(handler.Events)
		eventByID := http.HandlerFunc(handler.EventByID)

		createReq := httptest.NewRequest(http.MethodPost, "/api/events", mustJSONBody(t, map[string]any{
			"title":            "New Event",
			"shortDescription": "Short text",
			"fullDescription":  "Long event description",
			"categoryIds":      []string{"music"},
			"tagIds":           []string{"rock"},
			"imageUrls":        []string{"https://example.com/new.jpg"},
			"sessions": []map[string]any{
				{
					"placeId": "place-1",
					"startAt": "2026-04-20T19:00:00Z",
					"endAt":   "2026-04-20T21:00:00Z",
					"price":   1200,
				},
			},
		}))
		createReq = createReq.WithContext(context.WithValue(createReq.Context(), httpx.UserIDContextKey, "user-1"))
		createRec := httptest.NewRecorder()
		events.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("create status = %d, want %d body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
		}

		createPayload := decodeJSONMap(t, createRec.Body)
		eventID, ok := createPayload["id"].(string)
		if !ok || eventID == "" {
			t.Fatalf("unexpected create payload: %+v", createPayload)
		}

		patchReq := httptest.NewRequest(http.MethodPatch, "/api/events/"+eventID, mustJSONBody(t, map[string]any{
			"title":       "Updated Event",
			"sourceUrl":   nil,
			"categoryIds": []string{},
		}))
		patchReq = patchReq.WithContext(context.WithValue(patchReq.Context(), httpx.UserIDContextKey, "user-1"))
		patchRec := httptest.NewRecorder()
		eventByID.ServeHTTP(patchRec, patchReq)
		if patchRec.Code != http.StatusOK {
			t.Fatalf("patch status = %d, want %d body=%s", patchRec.Code, http.StatusOK, patchRec.Body.String())
		}

		deleteReq := httptest.NewRequest(http.MethodDelete, "/api/events/"+eventID, nil)
		deleteReq = deleteReq.WithContext(context.WithValue(deleteReq.Context(), httpx.UserIDContextKey, "user-1"))
		deleteRec := httptest.NewRecorder()
		eventByID.ServeHTTP(deleteRec, deleteReq)
		if deleteRec.Code != http.StatusOK {
			t.Fatalf("delete status = %d, want %d body=%s", deleteRec.Code, http.StatusOK, deleteRec.Body.String())
		}
	})

	t.Run("create without sessions", func(t *testing.T) {
		events := http.HandlerFunc(handler.Events)
		eventByID := http.HandlerFunc(handler.EventByID)

		createReq := httptest.NewRequest(http.MethodPost, "/api/events", mustJSONBody(t, map[string]any{
			"title":            "No Sessions Event",
			"shortDescription": "Short text",
			"fullDescription":  "Long event description",
			"categoryIds":      []string{"music"},
			"imageUrls":        []string{"https://example.com/new.jpg"},
		}))
		createReq = createReq.WithContext(context.WithValue(createReq.Context(), httpx.UserIDContextKey, "user-1"))
		createRec := httptest.NewRecorder()
		events.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("create without sessions status = %d, want %d body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
		}

		createPayload := decodeJSONMap(t, createRec.Body)
		eventID, ok := createPayload["id"].(string)
		if !ok || eventID == "" {
			t.Fatalf("unexpected create payload: %+v", createPayload)
		}

		detailsReq := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID, nil)
		detailsRec := httptest.NewRecorder()
		eventByID.ServeHTTP(detailsRec, detailsReq)
		if detailsRec.Code != http.StatusOK {
			t.Fatalf("details status = %d, want %d body=%s", detailsRec.Code, http.StatusOK, detailsRec.Body.String())
		}

		detailsPayload := decodeJSONMap(t, detailsRec.Body)
		sessions, ok := detailsPayload["sessions"].([]any)
		if !ok {
			t.Fatalf("missing sessions in details: %+v", detailsPayload)
		}
		if len(sessions) != 0 {
			t.Fatalf("sessions length = %d, want 0", len(sessions))
		}
	})
}

func TestHomeHandlerReturnsHomePayload(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(handler.Home).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("home status = %d, want %d", rec.Code, http.StatusOK)
	}

	payload := decodeJSONMap(t, rec.Body)
	if _, ok := payload["featuredEvents"].([]any); !ok {
		t.Fatalf("home payload missing featuredEvents: %+v", payload)
	}
	if _, ok := payload["categories"].([]any); !ok {
		t.Fatalf("home payload missing categories: %+v", payload)
	}
	if _, ok := payload["collections"].([]any); !ok {
		t.Fatalf("home payload missing collections: %+v", payload)
	}
}

func TestCategoriesHandlerReturnsCategories(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(handler.Categories).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("categories status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	payload := decodeJSONMap(t, rec.Body)
	items, ok := payload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("categories payload missing items: %+v", payload)
	}

	badReq := httptest.NewRequest(http.MethodPost, "/api/categories", nil)
	badRec := httptest.NewRecorder()
	http.HandlerFunc(handler.Categories).ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("bad method status = %d, want %d", badRec.Code, http.StatusMethodNotAllowed)
	}
}

func TestTagsHandlerReturnsTags(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(handler.Tags).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("tags status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	payload := decodeJSONMap(t, rec.Body)
	items, ok := payload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("tags payload missing items: %+v", payload)
	}

	badReq := httptest.NewRequest(http.MethodPost, "/api/tags", nil)
	badRec := httptest.NewRecorder()
	http.HandlerFunc(handler.Tags).ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("bad method status = %d, want %d", badRec.Code, http.StatusMethodNotAllowed)
	}
}

func TestCollectionsHandlers(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo))

	t.Run("list collections", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/collections", nil)
		rec := httptest.NewRecorder()
		http.HandlerFunc(handler.Collections).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("collections status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		items, ok := payload["items"].([]any)
		if !ok || len(items) == 0 {
			t.Fatalf("collections payload missing items: %+v", payload)
		}

		badReq := httptest.NewRequest(http.MethodPost, "/api/collections", nil)
		badRec := httptest.NewRecorder()
		http.HandlerFunc(handler.Collections).ServeHTTP(badRec, badReq)
		if badRec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("bad method status = %d, want %d", badRec.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("collection details", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/collections/weekend-picks", nil)
		rec := httptest.NewRecorder()
		http.HandlerFunc(handler.CollectionByID).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("collection details status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		if payload["id"] != "weekend-picks" {
			t.Fatalf("unexpected collection details response: %+v", payload)
		}
		if _, ok := payload["events"].([]any); !ok {
			t.Fatalf("missing events in collection details: %+v", payload)
		}

		missingReq := httptest.NewRequest(http.MethodGet, "/api/collections/unknown", nil)
		missingRec := httptest.NewRecorder()
		http.HandlerFunc(handler.CollectionByID).ServeHTTP(missingRec, missingReq)
		if missingRec.Code != http.StatusNotFound {
			t.Fatalf("collection missing status = %d, want %d", missingRec.Code, http.StatusNotFound)
		}
	})
}

func TestSearchSuggestionsHandler(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo))

	req := httptest.NewRequest(http.MethodGet, "/api/search?query=ro&limit=5", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(handler.Search).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	payload := decodeJSONMap(t, rec.Body)
	items, ok := payload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("search payload missing items: %+v", payload)
	}

	badReq := httptest.NewRequest(http.MethodGet, "/api/search?query=r", nil)
	badRec := httptest.NewRecorder()
	http.HandlerFunc(handler.Search).ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("bad query status = %d, want %d body=%s", badRec.Code, http.StatusBadRequest, badRec.Body.String())
	}

	badLimitReq := httptest.NewRequest(http.MethodGet, "/api/search?query=ro&limit=11", nil)
	badLimitRec := httptest.NewRecorder()
	http.HandlerFunc(handler.Search).ServeHTTP(badLimitRec, badLimitReq)
	if badLimitRec.Code != http.StatusBadRequest {
		t.Fatalf("bad limit status = %d, want %d body=%s", badLimitRec.Code, http.StatusBadRequest, badLimitRec.Body.String())
	}
}
