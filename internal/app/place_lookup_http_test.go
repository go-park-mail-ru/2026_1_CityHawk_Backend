package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	placedelivery "cityhawk/backend/internal/place/delivery/http"
	placemodel "cityhawk/backend/internal/place/model"
	placeusecase "cityhawk/backend/internal/place/usecase"
)

type stubPlaceLookupUsecase struct{}

func (stubPlaceLookupUsecase) Suggest(_ context.Context, query string, limit int) ([]placemodel.PlaceSuggestion, error) {
	return []placemodel.PlaceSuggestion{
		{
			Token:       "signed-token",
			Label:       query,
			Name:        "ВДНХ",
			AddressLine: "проспект Мира, 119",
			CityName:    "Москва",
			CountryName: "Russia",
			Timezone:    "Europe/Moscow",
			Latitude:    55.8298,
			Longitude:   37.6331,
			PhotonSource: placemodel.PlaceSuggestionSource{
				OSMID:    1,
				OSMType:  "W",
				OSMKey:   "tourism",
				OSMValue: "attraction",
			},
		},
	}, nil
}

func (stubPlaceLookupUsecase) Resolve(_ context.Context, input placemodel.PlaceResolveInput) (placemodel.PlaceResolved, error) {
	if input.Token != "signed-token" {
		return placemodel.PlaceResolved{}, placeusecase.ErrInvalidPlaceSuggestion
	}
	return placemodel.PlaceResolved{
		ID:          "place-123",
		CityID:      "city-123",
		Name:        "ВДНХ",
		AddressLine: "проспект Мира, 119",
		CityName:    "Москва",
		CountryName: "Russia",
		Timezone:    "Europe/Moscow",
		Latitude:    55.8298,
		Longitude:   37.6331,
	}, nil
}

func TestPlaceLookupHandler(t *testing.T) {
	handler := placedelivery.NewPlaceLookupHandler(stubPlaceLookupUsecase{})

	t.Run("suggestions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/place-suggestions?query=вднх&limit=5", nil)
		rec := httptest.NewRecorder()

		http.HandlerFunc(handler.Suggestions).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("suggestions status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		items, ok := payload["items"].([]any)
		if !ok || len(items) != 1 {
			t.Fatalf("unexpected suggestions payload: %+v", payload)
		}
	})

	t.Run("resolve", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/places/resolve", mustJSONBody(t, map[string]any{
			"token": "signed-token",
		}))
		rec := httptest.NewRecorder()

		http.HandlerFunc(handler.Resolve).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("resolve status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		if payload["id"] != "place-123" {
			t.Fatalf("unexpected resolve payload: %+v", payload)
		}
	})
}
