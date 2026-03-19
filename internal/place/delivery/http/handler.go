package http

import (
	"net/http"
	"strings"

	placeusecase "cityhawk/backend/internal/place/usecase"
	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
)

type Handler struct {
	places placeusecase.PlaceUsecase
}

func NewHandler(places placeusecase.PlaceUsecase) *Handler {
	return &Handler{places: places}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		httpx.WriteJSON(w, http.StatusOK, toHomePayloadResponse(h.places.HomePayload()))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": toPlaceCardResponses(h.places.ListCards())})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Details(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		id := strings.TrimPrefix(r.URL.Path, "/places/")
		if id == "" || strings.Contains(id, "/") {
			return platformmiddleware.NewHTTPError(http.StatusNotFound, "place not found")
		}
		p, ok := h.places.GetByID(id)
		if !ok {
			return platformmiddleware.NewHTTPError(http.StatusNotFound, "place not found")
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"item": toPlaceResponse(p)})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) ByCategory(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		category := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/places/category/"))
		if category == "" || strings.Contains(category, "/") {
			return platformmiddleware.NewHTTPError(http.StatusNotFound, "category not found")
		}
		items, ok := h.places.ListByCategory(category)
		if !ok {
			return platformmiddleware.NewHTTPError(http.StatusNotFound, "category not found")
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": toPlaceCardResponses(items)})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Best(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return platformmiddleware.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": toPlaceCardResponses(h.places.Best(placeusecase.DefaultBestPlacesLimit))})
		return nil
	}).ServeHTTP(w, r)
}
