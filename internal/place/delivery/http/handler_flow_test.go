package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	"cityhawk/backend/internal/platform/httpx"
)

func TestHandlerReadEndpoints(t *testing.T) {
	events := newFakeEventsUsecase()
	handler := NewHandler(events, nil)

	tests := []struct {
		name   string
		method string
		path   string
		call   func(http.ResponseWriter, *http.Request)
		status int
	}{
		{"home", http.MethodGet, "/api/home?city=Moscow", handler.Home, http.StatusOK},
		{"categories", http.MethodGet, "/api/categories", handler.Categories, http.StatusOK},
		{"tags", http.MethodGet, "/api/tags", handler.Tags, http.StatusOK},
		{"cities", http.MethodGet, "/api/cities", handler.Cities, http.StatusOK},
		{"collections", http.MethodGet, "/api/collections", handler.Collections, http.StatusOK},
		{"collection by id", http.MethodGet, "/api/collections/collection-1", handler.CollectionByID, http.StatusOK},
		{"search", http.MethodGet, "/api/search?query=ja&limit=5", handler.Search, http.StatusOK},
		{"events list", http.MethodGet, "/api/events?query=jazz&limit=10&offset=0&sort=dateAsc&dateFrom=2026-05-01&dateTo=2026-05-31", handler.Events, http.StatusOK},
		{"event details", http.MethodGet, "/api/events/event-1", handler.EventByID, http.StatusOK},
		{"map collections", http.MethodGet, "/api/map/collections?limit=1", handler.MapCollections, http.StatusOK},
		{"map filters", http.MethodGet, "/api/map/filters", handler.MapFilters, http.StatusOK},
		{"map collection spots", http.MethodGet, "/api/map/collections/collection-1/spots?limit=10&offset=0&sort=popular", handler.MapCollectionSpots, http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, "user-1"))
			rec := httptest.NewRecorder()

			tc.call(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandlerWriteEndpoints(t *testing.T) {
	events := newFakeEventsUsecase()
	handler := NewHandler(events, nil)

	ageLimit := 18
	createBody := createEventRequest{
		Title:            "Created event",
		ShortDescription: "Short description",
		FullDescription:  "Full description",
		AgeLimit:         &ageLimit,
		SourceURL:        stringPtr("https://events.example.com"),
		CategoryIDs:      []string{"cat-1"},
		TagIDs:           []string{"tag-1"},
		ImageURLs:        []string{"/uploads/event.png"},
		Sessions: []eventSessionRequest{{
			PlaceID: "place-1",
			StartAt: "2026-05-04T10:00:00Z",
			EndAt:   "2026-05-04T12:00:00Z",
			Price:   1200,
		}},
	}

	rec := performJSON(t, handler.Events, http.MethodPost, "/api/events", createBody, "user-1")
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}
	if events.created.AuthorUserID != "user-1" || events.created.Title == nil || *events.created.Title != "Created event" {
		t.Fatalf("create input was not passed: %+v", events.created)
	}

	patchBody := map[string]any{
		"title":       "Updated event",
		"sourceUrl":   nil,
		"categoryIds": []string{"cat-1", "cat-2"},
		"sessions": []map[string]any{{
			"placeId": "place-1",
			"startAt": "2026-05-05T10:00:00Z",
			"endAt":   "2026-05-05T12:00:00Z",
			"price":   1500,
		}},
	}
	rec = performJSON(t, handler.EventByID, http.MethodPatch, "/api/events/event-1", patchBody, "user-1")
	if rec.Code != http.StatusOK {
		t.Fatalf("patch status = %d body=%s", rec.Code, rec.Body.String())
	}
	if events.updated.ID != "event-1" || !events.updated.ClearSourceURL {
		t.Fatalf("patch input was not passed: %+v", events.updated)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/events/event-1", nil)
	req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, "user-1"))
	rec = httptest.NewRecorder()
	handler.EventByID(rec, req)
	if rec.Code != http.StatusOK || events.deletedID != "event-1" {
		t.Fatalf("delete status=%d deletedID=%q body=%s", rec.Code, events.deletedID, rec.Body.String())
	}
}

func TestHandlerValidationHelpers(t *testing.T) {
	if _, err := parseEventListFilter(httptest.NewRequest(http.MethodGet, "/api/events?limit=bad", nil)); err == nil {
		t.Fatal("parseEventListFilter() accepted invalid limit")
	}
	if _, _, err := parseSearchSuggestionsRequest(httptest.NewRequest(http.MethodGet, "/api/search?query=x", nil)); err == nil {
		t.Fatal("parseSearchSuggestionsRequest() accepted short query")
	}
	if _, err := validateCreateEventRequest(createEventRequest{}, "user-1"); err == nil {
		t.Fatal("validateCreateEventRequest() accepted empty request")
	}
	if _, err := validatePatchEventRequest(patchEventRequest{AgeLimit: intPtr(42)}, "event-1", "user-1"); err == nil {
		t.Fatal("validatePatchEventRequest() accepted invalid age")
	}
	_, details := validateSessions([]eventSessionRequest{{PlaceID: " ", StartAt: "bad", EndAt: "bad", Price: -1}})
	if details["sessions[0].placeId"] == "" {
		t.Fatalf("expected place validation detail, got %+v", details)
	}
}

func performJSON(t *testing.T, call func(http.ResponseWriter, *http.Request), method, path string, body any, userID string) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, userID))
	rec := httptest.NewRecorder()
	call(rec, req)
	return rec
}

type fakeEventsUsecase struct {
	details   placemodel.EventDetailsView
	card      placemodel.EventCardView
	created   placemodel.EventWriteInput
	updated   placemodel.EventWriteInput
	deletedID string
}

func newFakeEventsUsecase() *fakeEventsUsecase {
	now := time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC)
	details := placemodel.EventDetailsView{
		ID:               "event-1",
		Title:            "Jazz night",
		ShortDescription: "Short description",
		FullDescription:  "Full description",
		AgeLimit:         18,
		Author:           placemodel.EventAuthorView{ID: "user-1", Username: "alice"},
		Categories:       []placemodel.EventTaxonomyItem{{ID: "cat-1", Name: "Music", Slug: "music"}},
		Tags:             []placemodel.EventTaxonomyItem{{ID: "tag-1", Name: "Jazz", Slug: "jazz"}},
		Images:           []placemodel.EventImageView{{ID: "img-1", ImageURL: "/uploads/event.png"}},
		Sessions: []placemodel.EventSessionView{{
			ID:      "session-1",
			StartAt: now,
			EndAt:   now.Add(2 * time.Hour),
			Price:   1200,
			Place: placemodel.EventSessionPlaceView{
				ID:          "place-1",
				Name:        "Main Hall",
				AddressLine: "Lenina 1",
				Latitude:    55.75,
				Longitude:   37.61,
				City:        placemodel.EventSessionPlaceCityView{ID: "city-1", Name: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow"},
			},
		}},
		CreatedAt: now,
		UpdatedAt: now,
		IsOwner:   true,
	}
	card := placemodel.EventCardView{
		ID:               details.ID,
		Title:            details.Title,
		ShortDescription: details.ShortDescription,
		CoverImageURL:    "/uploads/event.png",
		Tags:             details.Tags,
		NextSession: &placemodel.EventCardNextSession{
			StartAt: now,
			Place:   placemodel.EventCardNextSessionPlace{Name: "Main Hall", AddressLine: "Lenina 1"},
		},
		Popularity: 7,
	}
	return &fakeEventsUsecase{details: details, card: card}
}

func (f *fakeEventsUsecase) HomePayload(context.Context, placemodel.HomeFilter) placemodel.HomePayload {
	return placemodel.HomePayload{
		FeaturedEvents: []placemodel.HomeFeaturedEvent{{ID: "event-1", Title: "Jazz night", CoverImageURL: "/uploads/event.png"}},
		Categories:     []placemodel.HomeCategory{{ID: "cat-1", Name: "Music", Slug: "music"}},
		Collections:    []placemodel.HomeCollection{{ID: "collection-1", Title: "Weekend", ImageURL: "/uploads/collection.png"}},
	}
}

func (f *fakeEventsUsecase) ListCategories(context.Context) []placemodel.HomeCategory {
	return []placemodel.HomeCategory{{ID: "cat-1", Name: "Music", Slug: "music"}}
}

func (f *fakeEventsUsecase) ListTags(context.Context) []placemodel.HomeTag {
	return []placemodel.HomeTag{{ID: "tag-1", Name: "Jazz", Slug: "jazz"}}
}

func (f *fakeEventsUsecase) ListCities(context.Context) []placemodel.City {
	return []placemodel.City{{ID: "city-1", Name: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow"}}
}

func (f *fakeEventsUsecase) ListCollections(context.Context) ([]placemodel.CollectionCardView, error) {
	return []placemodel.CollectionCardView{{ID: "collection-1", Title: "Weekend", Description: "Best", ImageURL: "/uploads/collection.png", IsPublic: true}}, nil
}

func (f *fakeEventsUsecase) GetCollectionByID(context.Context, string) (placemodel.CollectionDetailsView, bool, error) {
	return placemodel.CollectionDetailsView{ID: "collection-1", Title: "Weekend", Description: "Best", ImageURL: "/uploads/collection.png", IsPublic: true, Events: []placemodel.EventCardView{f.card}}, true, nil
}

func (f *fakeEventsUsecase) SearchSuggestions(context.Context, string, int) ([]placemodel.SearchSuggestion, error) {
	return []placemodel.SearchSuggestion{{ID: "event-1", Type: "event", Label: "Jazz night"}}, nil
}

func (f *fakeEventsUsecase) ListEvents(context.Context, placemodel.EventListFilter) ([]placemodel.EventCardView, int, error) {
	return []placemodel.EventCardView{f.card}, 1, nil
}

func (f *fakeEventsUsecase) GetByID(context.Context, string, string) (placemodel.EventDetailsView, bool, error) {
	return f.details, true, nil
}

func (f *fakeEventsUsecase) CreateEvent(_ context.Context, input placemodel.EventWriteInput) (string, error) {
	f.created = input
	return "event-created", nil
}

func (f *fakeEventsUsecase) UpdateEvent(_ context.Context, input placemodel.EventWriteInput) (bool, error) {
	f.updated = input
	return true, nil
}

func (f *fakeEventsUsecase) DeleteEvent(_ context.Context, id, _ string) (bool, error) {
	f.deletedID = id
	return true, nil
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}
