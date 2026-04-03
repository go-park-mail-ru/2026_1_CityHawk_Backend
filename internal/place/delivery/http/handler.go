package http

import (
	"context"
	"net/http"
	"strings"

	placemodel "cityhawk/backend/internal/place/model"
	placeusecase "cityhawk/backend/internal/place/usecase"
	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
)

type PlaceUsecase interface {
	HomePayload(ctx context.Context) placemodel.HomePayload
	ListCards(ctx context.Context) []placemodel.PlaceCard
	GetByID(ctx context.Context, id string) (placemodel.Place, bool)
	ListByCategory(ctx context.Context, category string) ([]placemodel.PlaceCard, bool)
	Best(ctx context.Context, limit int) []placemodel.PlaceCard
}

type Handler struct {
	places PlaceUsecase
}

func NewHandler(places PlaceUsecase) *Handler {
	return &Handler{places: places}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		httpx.WriteJSON(w, http.StatusOK, toHomePayloadResponse(h.places.HomePayload(r.Context())))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": toPlaceCardResponses(h.places.ListCards(r.Context()))})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Details(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		id := strings.TrimPrefix(r.URL.Path, "/places/")
		if id == "" || strings.Contains(id, "/") {
			return httpx.NewHTTPError(http.StatusNotFound, "place not found")
		}
		p, ok := h.places.GetByID(r.Context(), id)
		if !ok {
			return httpx.NewHTTPError(http.StatusNotFound, "place not found")
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"item": toPlaceResponse(p)})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) ByCategory(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		category := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/places/category/"))
		if category == "" || strings.Contains(category, "/") {
			return httpx.NewHTTPError(http.StatusNotFound, "category not found")
		}
		items, ok := h.places.ListByCategory(r.Context(), category)
		if !ok {
			return httpx.NewHTTPError(http.StatusNotFound, "category not found")
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": toPlaceCardResponses(items)})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Best(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": toPlaceCardResponses(h.places.Best(r.Context(), placeusecase.DefaultBestPlacesLimit))})
		return nil
	}).ServeHTTP(w, r)
}
