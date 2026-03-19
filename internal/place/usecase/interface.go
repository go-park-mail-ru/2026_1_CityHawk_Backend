package usecase

import placemodel "cityhawk/backend/internal/place/model"

type PlaceUsecase interface {
	HomePayload() placemodel.HomePayload
	ListCards() []placemodel.PlaceCard
	GetByID(id string) (placemodel.Place, bool)
	ListByCategory(category string) ([]placemodel.PlaceCard, bool)
	Best(limit int) []placemodel.PlaceCard
}

