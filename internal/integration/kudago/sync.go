package kudago

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const externalSource = "kudago"

type SyncConfig struct {
	Enabled      bool
	Location     string
	SyncInterval time.Duration
}

type SyncStats struct {
	Processed int
	Imported  int
	Skipped   int
}

type Syncer struct {
	cfg    SyncConfig
	client *Client
	pool   *pgxpool.Pool
}

func NewSyncer(cfg SyncConfig, client *Client, pool *pgxpool.Pool) *Syncer {
	interval := cfg.SyncInterval
	if interval <= 0 {
		interval = 15 * time.Minute
	}

	return &Syncer{
		cfg: SyncConfig{
			Enabled:      cfg.Enabled,
			Location:     strings.TrimSpace(cfg.Location),
			SyncInterval: interval,
		},
		client: client,
		pool:   pool,
	}
}

func (s *Syncer) Start(ctx context.Context) {
	s.runAndLog(ctx)

	ticker := time.NewTicker(s.cfg.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runAndLog(ctx)
		}
	}
}

func (s *Syncer) runAndLog(ctx context.Context) {
	stats, err := s.RunOnce(ctx)
	if err != nil {
		log.Printf("kudago sync failed: %v", err)
		return
	}

	log.Printf(
		"kudago sync done location=%s processed=%d imported=%d skipped=%d",
		s.cfg.Location, stats.Processed, stats.Imported, stats.Skipped,
	)
}

func (s *Syncer) RunOnce(ctx context.Context) (SyncStats, error) {
	events, err := s.client.FetchEvents(ctx)
	if err != nil {
		return SyncStats{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return SyncStats{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	location := locationFromSlug(s.cfg.Location)
	cityID, err := ensureCity(ctx, tx, location.City, location.Country, location.Timezone)
	if err != nil {
		return SyncStats{}, err
	}
	authorUserID, err := ensureSystemUser(ctx, tx, cityID)
	if err != nil {
		return SyncStats{}, err
	}

	stats := SyncStats{Processed: len(events)}
	for i, event := range events {
		savepoint := fmt.Sprintf("kudago_event_%d", i+1)
		if _, err := tx.Exec(ctx, "SAVEPOINT "+savepoint); err != nil {
			return SyncStats{}, err
		}

		imported, err := importEvent(ctx, tx, authorUserID, cityID, event)
		if err != nil {
			stats.Skipped++
			_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+savepoint)
			_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+savepoint)
			log.Printf("kudago event skip external_id=%s reason=%v", event.ExternalID, err)
			continue
		}
		if _, err := tx.Exec(ctx, "RELEASE SAVEPOINT "+savepoint); err != nil {
			return SyncStats{}, err
		}
		if imported {
			stats.Imported++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return SyncStats{}, err
	}

	stats.Skipped += stats.Processed - stats.Imported - stats.Skipped
	if stats.Skipped < 0 {
		stats.Skipped = 0
	}
	return stats, nil
}

type locationInfo struct {
	City     string
	Country  string
	Timezone string
}

func locationFromSlug(slug string) locationInfo {
	switch strings.ToLower(strings.TrimSpace(slug)) {
	case "spb":
		return locationInfo{City: "Санкт-Петербург", Country: "Россия", Timezone: "Europe/Moscow"}
	default:
		return locationInfo{City: "Москва", Country: "Россия", Timezone: "Europe/Moscow"}
	}
}

func importEvent(ctx context.Context, tx pgx.Tx, authorUserID, cityID string, item Event) (bool, error) {
	title := truncate(item.Title, 200)
	if title == "" {
		return false, fmt.Errorf("empty title")
	}

	locationDescription := truncate(buildLocationDescription(item.PlaceName, item.PlaceAddress), 500)
	if locationDescription == "" {
		locationDescription = "Москва"
	}

	var fullDescription *string
	if text := truncate(firstNonEmpty(item.FullDescription, item.ShortDescription), 5000); text != "" {
		fullDescription = &text
	}

	var sourceURL *string
	if parsed := normalizeURL(item.SourceURL); parsed != "" {
		sourceURL = &parsed
	}

	eventID, inserted, err := upsertEvent(ctx, tx, upsertEventInput{
		ExternalSource:      externalSource,
		ExternalID:          truncate(item.ExternalID, 128),
		AuthorUserID:        authorUserID,
		Title:               title,
		LocationDescription: locationDescription,
		FullDescription:     fullDescription,
		AgeLimit:            clampAge(item.AgeLimit),
		SourceURL:           sourceURL,
	})
	if err != nil {
		return false, err
	}

	placeID, err := ensurePlace(ctx, tx, ensurePlaceInput{
		CityID:      cityID,
		Name:        truncate(firstNonEmpty(item.PlaceName, "Неизвестная площадка"), 200),
		AddressLine: truncate(firstNonEmpty(item.PlaceAddress, "Москва"), 300),
		Latitude:    normalizeLatitude(item.Latitude),
		Longitude:   normalizeLongitude(item.Longitude),
	})
	if err != nil {
		return false, err
	}

	for _, imageURL := range item.Images {
		normalized := normalizeURL(imageURL)
		if normalized == "" {
			continue
		}
		if err := ensureEventImage(ctx, tx, eventID, normalized); err != nil {
			return false, err
		}
	}

	for _, categoryName := range item.Categories {
		normalized := normalizeTaxonomyName(categoryName)
		if normalized == "" {
			continue
		}
		categoryID, err := ensureCategory(ctx, tx, normalized)
		if err != nil {
			return false, err
		}
		if err := ensureEventCategory(ctx, tx, eventID, categoryID); err != nil {
			return false, err
		}
	}

	for _, tagName := range item.Tags {
		normalized := normalizeTaxonomyName(tagName)
		if normalized == "" {
			continue
		}
		tagID, err := ensureTag(ctx, tx, normalized)
		if err != nil {
			return false, err
		}
		if err := ensureEventTag(ctx, tx, eventID, tagID); err != nil {
			return false, err
		}
	}

	for _, session := range item.Sessions {
		if !session.EndAt.After(session.StartAt) {
			continue
		}
		if err := ensureEventSession(ctx, tx, eventID, placeID, session.StartAt, session.EndAt); err != nil {
			return false, err
		}
	}

	return inserted, nil
}

type upsertEventInput struct {
	ExternalSource      string
	ExternalID          string
	AuthorUserID        string
	Title               string
	LocationDescription string
	FullDescription     *string
	AgeLimit            int
	SourceURL           *string
}

func upsertEvent(ctx context.Context, tx pgx.Tx, input upsertEventInput) (string, bool, error) {
	var (
		eventID  string
		inserted bool
	)

	err := tx.QueryRow(ctx, `
		INSERT INTO event (
			external_source,
			external_id,
			author_user_id,
			title,
			location_description,
			full_description,
			age_limit,
			source_url
		)
		VALUES ($1, $2, $3::uuid, $4, $5, $6, $7, $8)
		ON CONFLICT (external_source, external_id)
		DO UPDATE SET
			title = CASE WHEN EXCLUDED.title <> '' THEN EXCLUDED.title ELSE event.title END,
			location_description = CASE
				WHEN EXCLUDED.location_description <> '' THEN EXCLUDED.location_description
				ELSE event.location_description
			END,
			full_description = COALESCE(EXCLUDED.full_description, event.full_description),
			age_limit = CASE
				WHEN EXCLUDED.age_limit > 0 THEN EXCLUDED.age_limit
				ELSE event.age_limit
			END,
			source_url = COALESCE(EXCLUDED.source_url, event.source_url),
			updated_at = now()
		RETURNING id::text, (xmax = 0)
	`,
		input.ExternalSource,
		input.ExternalID,
		input.AuthorUserID,
		input.Title,
		input.LocationDescription,
		input.FullDescription,
		input.AgeLimit,
		input.SourceURL,
	).Scan(&eventID, &inserted)
	if err != nil {
		return "", false, err
	}

	return eventID, inserted, nil
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

func ensureSystemUser(ctx context.Context, tx pgx.Tx, cityID string) (string, error) {
	var userID string
	err := tx.QueryRow(ctx, `
		INSERT INTO user_account (email, username, user_surname, password_hash, city_id, avatar_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email)
		DO UPDATE SET city_id = EXCLUDED.city_id, updated_at = now()
		RETURNING id::text
	`,
		"system@cityhawk.local",
		"system_importer",
		"CityHawk",
		"disabled",
		cityID,
		"https://cityhawk.local/static/system-user.png",
	).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

type ensurePlaceInput struct {
	CityID      string
	Name        string
	AddressLine string
	Latitude    float64
	Longitude   float64
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
	`,
		input.CityID,
		input.Name,
		input.AddressLine,
		input.Latitude,
		input.Longitude,
	).Scan(&placeID)
	if err != nil {
		return "", err
	}
	return placeID, nil
}

func ensureEventImage(ctx context.Context, tx pgx.Tx, eventID, imageURL string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO event_image (event_id, image_url)
		SELECT $1::uuid, $2
		WHERE NOT EXISTS (
			SELECT 1 FROM event_image
			WHERE event_id = $1::uuid AND image_url = $2
		)
	`, eventID, imageURL)
	return err
}

func ensureCategory(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO category (name)
		VALUES ($1)
		ON CONFLICT (name)
		DO UPDATE SET updated_at = now()
		RETURNING id::text
	`, name).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureTag(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO tag (name)
		VALUES ($1)
		ON CONFLICT (name)
		DO UPDATE SET updated_at = now()
		RETURNING id::text
	`, name).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func ensureEventCategory(ctx context.Context, tx pgx.Tx, eventID, categoryID string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO event_category (event_id, category_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (event_id, category_id) DO NOTHING
	`, eventID, categoryID)
	return err
}

func ensureEventTag(ctx context.Context, tx pgx.Tx, eventID, tagID string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO event_tag (event_id, tag_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (event_id, tag_id) DO NOTHING
	`, eventID, tagID)
	return err
}

func ensureEventSession(ctx context.Context, tx pgx.Tx, eventID, placeID string, startAt, endAt time.Time) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO event_session (event_id, place_id, start_at, end_at, price)
		SELECT $1::uuid, $2::uuid, $3, $4, 0
		WHERE NOT EXISTS (
			SELECT 1
			FROM event_session es
			WHERE es.event_id = $1::uuid
			  AND es.place_id = $2::uuid
			  AND es.start_at = $3
			  AND es.end_at = $4
		)
		AND NOT EXISTS (
			SELECT 1
			FROM event_session es
			WHERE es.place_id = $2::uuid
			  AND tstzrange(es.start_at, es.end_at, '[)') && tstzrange($3, $4, '[)')
		)
	`, eventID, placeID, startAt, endAt)
	return err
}

func buildLocationDescription(placeName, placeAddress string) string {
	switch {
	case strings.TrimSpace(placeName) != "" && strings.TrimSpace(placeAddress) != "":
		return strings.TrimSpace(placeName) + ", " + strings.TrimSpace(placeAddress)
	case strings.TrimSpace(placeName) != "":
		return strings.TrimSpace(placeName)
	default:
		return strings.TrimSpace(placeAddress)
	}
}

func truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if value == "" || max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return strings.TrimSpace(string(runes[:max]))
}

func normalizeURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return truncate(value, 2048)
}

func normalizeLatitude(v float64) float64 {
	if v < -90 || v > 90 || v == 0 {
		return 55.751244
	}
	return v
}

func normalizeLongitude(v float64) float64 {
	if v < -180 || v > 180 || v == 0 {
		return 37.618423
	}
	return v
}

func normalizeTaxonomyName(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return ""
	}
	runes := make([]rune, 0, len(value))
	prevDash := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			runes = append(runes, r)
			prevDash = false
		case r == '-' || r == '_' || unicode.IsSpace(r):
			if !prevDash {
				runes = append(runes, '-')
				prevDash = true
			}
		}
	}
	value = strings.Trim(string(runes), "-")
	value = truncate(value, 64)
	if value == "" {
		return ""
	}
	// Drop age-like numeric tags from provider taxonomy.
	if isDigitsOnly(value) {
		return ""
	}
	return value
}

func isDigitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
