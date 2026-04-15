package repository

import (
	"context"

	placemodel "cityhawk/backend/internal/place/model"
	"github.com/jackc/pgx/v5"
)

type PlaceLookupWriteInput struct {
	CityName    string
	CountryName string
	Timezone    string
	Name        string
	AddressLine string
	Latitude    float64
	Longitude   float64
	Description *string
}

func (r *PostgresRepository) EnsurePlace(ctx context.Context, input PlaceLookupWriteInput) (placemodel.PlaceResolved, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return placemodel.PlaceResolved{}, err
	}
	defer rollbackTx(ctx, tx)

	cityID, err := ensureCity(ctx, tx, input.CityName, input.CountryName, input.Timezone)
	if err != nil {
		return placemodel.PlaceResolved{}, err
	}

	placeID, err := ensurePlace(ctx, tx, ensurePlaceInput{
		CityID:      cityID,
		Name:        input.Name,
		AddressLine: input.AddressLine,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
	})
	if err != nil {
		return placemodel.PlaceResolved{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return placemodel.PlaceResolved{}, err
	}

	return placemodel.PlaceResolved{
		ID:          placeID,
		CityID:      cityID,
		Name:        input.Name,
		AddressLine: input.AddressLine,
		CityName:    input.CityName,
		CountryName: input.CountryName,
		Timezone:    input.Timezone,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
	}, nil
}

type ensurePlaceInput struct {
	CityID      string
	Name        string
	AddressLine string
	Latitude    float64
	Longitude   float64
}

func ensureCity(ctx context.Context, tx pgx.Tx, name, country, timezone string) (string, error) {
	var cityID string
	err := tx.QueryRow(ctx, `
		INSERT INTO city (name, country_name, timezone)
		VALUES ($1, $2, $3)
		ON CONFLICT (country_name, name)
		DO UPDATE SET timezone = EXCLUDED.timezone, updated_at = now()
		RETURNING id::text
	`, name, country, timezone).Scan(&cityID)
	if err != nil {
		return "", err
	}
	return cityID, nil
}

func ensurePlace(ctx context.Context, tx pgx.Tx, input ensurePlaceInput) (string, error) {
	var placeID string
	err := tx.QueryRow(ctx, `
		WITH existing AS (
			SELECT p.id::text AS id
			FROM place p
			WHERE p.city_id::text = $1
			  AND p.name = $2
			  AND p.address_line = $3
			LIMIT 1
		),
		inserted AS (
			INSERT INTO place (city_id, name, address_line, latitude, longitude, description)
			SELECT $1::uuid, $2, $3, $4, $5, NULL
			WHERE NOT EXISTS (SELECT 1 FROM existing)
			RETURNING id::text AS id
		)
		SELECT id FROM inserted
		UNION ALL
		SELECT id FROM existing
		LIMIT 1
	`, input.CityID, input.Name, input.AddressLine, input.Latitude, input.Longitude).Scan(&placeID)
	if err != nil {
		return "", err
	}
	return placeID, nil
}
