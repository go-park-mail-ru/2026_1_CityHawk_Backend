package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	photonintegration "cityhawk/backend/internal/integration/photon"
	placemodel "cityhawk/backend/internal/place/model"
	placerepo "cityhawk/backend/internal/place/repository"
)

var ErrInvalidPlaceSuggestion = errors.New("invalid place suggestion token")

type PhotonSearcher interface {
	Search(ctx context.Context, query string, limit int) ([]photonintegration.SearchResult, error)
}

type PlaceLookupRepository interface {
	EnsurePlace(ctx context.Context, input placerepo.PlaceLookupWriteInput) (placemodel.PlaceResolved, error)
}

type PlaceLookupService struct {
	photon          PhotonSearcher
	repo            PlaceLookupRepository
	signingKey      []byte
	defaultCountry  string
	defaultTimezone string
}

type signedPlacePayload struct {
	Name        string  `json:"name"`
	AddressLine string  `json:"addressLine"`
	CityName    string  `json:"cityName"`
	CountryName string  `json:"countryName"`
	Timezone    string  `json:"timezone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Postcode    string  `json:"postcode"`
	District    string  `json:"district"`
	OSMID       int64   `json:"osmId"`
	OSMType     string  `json:"osmType"`
	OSMKey      string  `json:"osmKey"`
	OSMValue    string  `json:"osmValue"`
}

func NewPlaceLookupService(photon PhotonSearcher, repo PlaceLookupRepository, signingKey []byte, defaultCountry, defaultTimezone string) *PlaceLookupService {
	return &PlaceLookupService{
		photon:          photon,
		repo:            repo,
		signingKey:      signingKey,
		defaultCountry:  strings.TrimSpace(defaultCountry),
		defaultTimezone: strings.TrimSpace(defaultTimezone),
	}
}

func (s *PlaceLookupService) Suggest(ctx context.Context, query string, limit int) ([]placemodel.PlaceSuggestion, error) {
	results, err := s.photon.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	items := make([]placemodel.PlaceSuggestion, 0, len(results))
	for _, item := range results {
		payload := signedPlacePayload{
			Name:        normalizePlaceName(item),
			AddressLine: buildAddressLine(item),
			CityName:    normalizeCityName(item),
			CountryName: normalizeCountryName(item, s.defaultCountry),
			Timezone:    normalizeTimezone(s.defaultTimezone),
			Latitude:    item.Latitude,
			Longitude:   item.Longitude,
			Postcode:    item.Postcode,
			District:    item.District,
			OSMID:       item.OSMID,
			OSMType:     item.OSMType,
			OSMKey:      item.OSMKey,
			OSMValue:    item.OSMValue,
		}
		if payload.Name == "" || payload.AddressLine == "" || payload.CityName == "" || payload.CountryName == "" {
			continue
		}

		token, err := s.signPayload(payload)
		if err != nil {
			return nil, err
		}

		items = append(items, placemodel.PlaceSuggestion{
			Token:       token,
			Label:       item.DisplayLabel,
			Name:        payload.Name,
			AddressLine: payload.AddressLine,
			CityName:    payload.CityName,
			CountryName: payload.CountryName,
			Timezone:    payload.Timezone,
			Latitude:    payload.Latitude,
			Longitude:   payload.Longitude,
			Postcode:    payload.Postcode,
			District:    payload.District,
			PhotonSource: placemodel.PlaceSuggestionSource{
				OSMID:    payload.OSMID,
				OSMType:  payload.OSMType,
				OSMKey:   payload.OSMKey,
				OSMValue: payload.OSMValue,
			},
		})
	}

	return items, nil
}

func (s *PlaceLookupService) Resolve(ctx context.Context, input placemodel.PlaceResolveInput) (placemodel.PlaceResolved, error) {
	payload, err := s.verifyPayload(input.Token)
	if err != nil {
		return placemodel.PlaceResolved{}, err
	}

	return s.repo.EnsurePlace(ctx, placerepo.PlaceLookupWriteInput{
		CityName:    payload.CityName,
		CountryName: payload.CountryName,
		Timezone:    payload.Timezone,
		Name:        payload.Name,
		AddressLine: payload.AddressLine,
		Latitude:    payload.Latitude,
		Longitude:   payload.Longitude,
	})
}

func (s *PlaceLookupService) signPayload(payload signedPlacePayload) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	encodedBody := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, s.signingKey)
	_, _ = mac.Write([]byte(encodedBody))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedBody + "." + signature, nil
}

func (s *PlaceLookupService) verifyPayload(token string) (signedPlacePayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return signedPlacePayload{}, ErrInvalidPlaceSuggestion
	}

	mac := hmac.New(sha256.New, s.signingKey)
	_, _ = mac.Write([]byte(parts[0]))
	expected := mac.Sum(nil)

	got, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(got, expected) {
		return signedPlacePayload{}, ErrInvalidPlaceSuggestion
	}

	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return signedPlacePayload{}, ErrInvalidPlaceSuggestion
	}

	var payload signedPlacePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return signedPlacePayload{}, ErrInvalidPlaceSuggestion
	}
	if payload.Name == "" || payload.AddressLine == "" || payload.CityName == "" || payload.CountryName == "" {
		return signedPlacePayload{}, ErrInvalidPlaceSuggestion
	}

	return payload, nil
}

func normalizePlaceName(item photonintegration.SearchResult) string {
	return firstNonEmpty(strings.TrimSpace(item.Name), buildAddressLine(item))
}

func buildAddressLine(item photonintegration.SearchResult) string {
	street := strings.TrimSpace(item.Street)
	house := strings.TrimSpace(item.HouseNumber)
	switch {
	case street != "" && house != "":
		return street + " " + house
	case street != "":
		return street
	default:
		return firstNonEmpty(strings.TrimSpace(item.Name), strings.TrimSpace(item.District))
	}
}

func normalizeCityName(item photonintegration.SearchResult) string {
	return firstNonEmpty(strings.TrimSpace(item.City), strings.TrimSpace(item.State), strings.TrimSpace(item.District))
}

func normalizeCountryName(item photonintegration.SearchResult, fallback string) string {
	return firstNonEmpty(strings.TrimSpace(item.Country), strings.TrimSpace(fallback))
}

func normalizeTimezone(fallback string) string {
	return firstNonEmpty(strings.TrimSpace(fallback), "UTC")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
