package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	placemodel "cityhawk/backend/internal/place/model"
	placeusecase "cityhawk/backend/internal/place/usecase"
	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
)

type PlaceLookupUsecase interface {
	Suggest(ctx context.Context, query string, limit int) ([]placemodel.PlaceSuggestion, error)
	Resolve(ctx context.Context, input placemodel.PlaceResolveInput) (placemodel.PlaceResolved, error)
}

type PlaceLookupHandler struct {
	lookup PlaceLookupUsecase
}

type placeResolveRequest struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type placeSuggestionsResponse struct {
	Items []placeSuggestionResponse `json:"items"`
}

type placeSuggestionResponse struct {
	Token       string                        `json:"token"`
	Label       string                        `json:"label"`
	Name        string                        `json:"name"`
	AddressLine string                        `json:"addressLine"`
	CityName    string                        `json:"cityName"`
	CountryName string                        `json:"countryName"`
	Timezone    string                        `json:"timezone"`
	Latitude    float64                       `json:"latitude"`
	Longitude   float64                       `json:"longitude"`
	Postcode    string                        `json:"postcode,omitempty"`
	District    string                        `json:"district,omitempty"`
	Source      placeSuggestionSourceResponse `json:"source"`
}

type placeSuggestionSourceResponse struct {
	Provider string `json:"provider"`
	OSMID    int64  `json:"osmId"`
	OSMType  string `json:"osmType"`
	OSMKey   string `json:"osmKey"`
	OSMValue string `json:"osmValue"`
}

type placeResolvedResponse struct {
	ID          string  `json:"id"`
	CityID      string  `json:"cityId"`
	Name        string  `json:"name"`
	AddressLine string  `json:"addressLine"`
	CityName    string  `json:"cityName"`
	CountryName string  `json:"countryName"`
	Timezone    string  `json:"timezone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

func NewPlaceLookupHandler(lookup PlaceLookupUsecase) *PlaceLookupHandler {
	return &PlaceLookupHandler{lookup: lookup}
}

func (h *PlaceLookupHandler) Suggestions(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		query := strings.TrimSpace(r.URL.Query().Get("query"))
		if query == "" {
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"query": "query is required"})
		}

		limit := 5
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value <= 0 || value > 10 {
				return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"limit": "limit must be between 1 and 10"})
			}
			limit = value
		}

		items, err := h.lookup.Suggest(r.Context(), query, limit)
		if err != nil {
			return err
		}

		resp := placeSuggestionsResponse{Items: make([]placeSuggestionResponse, 0, len(items))}
		for _, item := range items {
			resp.Items = append(resp.Items, placeSuggestionResponse{
				Token:       item.Token,
				Label:       item.Label,
				Name:        item.Name,
				AddressLine: item.AddressLine,
				CityName:    item.CityName,
				CountryName: item.CountryName,
				Timezone:    item.Timezone,
				Latitude:    item.Latitude,
				Longitude:   item.Longitude,
				Postcode:    item.Postcode,
				District:    item.District,
				Source: placeSuggestionSourceResponse{
					Provider: "photon",
					OSMID:    item.PhotonSource.OSMID,
					OSMType:  item.PhotonSource.OSMType,
					OSMKey:   item.PhotonSource.OSMKey,
					OSMValue: item.PhotonSource.OSMValue,
				},
			})
		}

		httpx.WriteJSON(w, http.StatusOK, resp)
		return nil
	}).ServeHTTP(w, r)
}

func (h *PlaceLookupHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		var req placeResolveRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse("invalid json", nil))
			return nil
		}

		if strings.TrimSpace(req.Token) == "" {
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"token": "token is required"})
		}

		item, err := h.lookup.Resolve(r.Context(), placemodel.PlaceResolveInput{Token: req.Token, Name: req.Name})
		if err != nil {
			if errors.Is(err, placeusecase.ErrInvalidPlaceSuggestion) {
				return httpx.NewHTTPError(http.StatusBadRequest, "invalid place suggestion")
			}
			return err
		}

		httpx.WriteJSON(w, http.StatusOK, placeResolvedResponse{
			ID:          item.ID,
			CityID:      item.CityID,
			Name:        item.Name,
			AddressLine: item.AddressLine,
			CityName:    item.CityName,
			CountryName: item.CountryName,
			Timezone:    item.Timezone,
			Latitude:    item.Latitude,
			Longitude:   item.Longitude,
		})
		return nil
	}).ServeHTTP(w, r)
}
