package repository

import (
	"errors"
	"strings"
	"testing"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	platformpostgres "cityhawk/backend/internal/platform/postgres"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestBuildListEventsQueryUsesJoinFilters(t *testing.T) {
	dateFrom := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2026, time.May, 31, 23, 59, 59, 0, time.UTC)

	query, args := buildListEventsQuery(placemodel.EventListFilter{
		Query:        "jazz",
		CategoryID:   "category-1",
		TagID:        "tag-1",
		CityID:       "city-1",
		DateFrom:     &dateFrom,
		DateTo:       &dateTo,
		AuthorID:     "author-1",
		CollectionID: "collection-1",
		UserID:       "user-1",
		Limit:        20,
		Offset:       40,
	})

	for _, unwanted := range []string{
		"SELECT DISTINCT ON (es.event_id)",
		"EXISTS (SELECT 1 FROM event_category ec_filter",
		"EXISTS (SELECT 1 FROM event_tag et_filter",
		"EXISTS (SELECT 1 FROM event_session",
		"EXISTS (SELECT 1 FROM collection_event",
		"EXISTS (SELECT 1 FROM favorite_event",
	} {
		if strings.Contains(query, unwanted) {
			t.Fatalf("buildListEventsQuery() contains %q in query:\n%s", unwanted, query)
		}
	}

	for _, wanted := range []string{
		"JOIN event_category ec_filter",
		"JOIN event_tag et_filter",
		"JOIN event_session es_filter",
		"FROM event_category ec_query",
		"FROM event_tag et_query",
		"FROM event_place ep_city",
		"FROM event_session es_city",
		"JOIN collection_event ce_filter",
		"LEFT JOIN favorite_event fe_filter",
		"ROW_NUMBER() OVER (PARTITION BY es.event_id",
	} {
		if !strings.Contains(query, wanted) {
			t.Fatalf("buildListEventsQuery() does not contain %q in query:\n%s", wanted, query)
		}
	}

	if len(args) != 11 {
		t.Fatalf("buildListEventsQuery() args len = %d, want 11", len(args))
	}
}

func TestToEventCardIncludesPlaceWithoutNextSession(t *testing.T) {
	card := toEventCard(eventListRow{
		ID:          "event-1",
		Title:       "Secret place",
		ShortDesc:   "Hidden yard",
		PlaceID:     "place-1",
		PlaceName:   "Hidden Yard",
		AddressLine: "Moscow, Tverskaya 1",
		Latitude:    55.7,
		Longitude:   37.6,
		CityID:      "city-1",
		CityName:    "Moscow",
		CountryName: "Russia",
		Timezone:    "Europe/Moscow",
	})

	if card.NextSession != nil {
		t.Fatalf("toEventCard() nextSession = %+v, want nil", card.NextSession)
	}
	if card.Place == nil || card.Place.Name != "Hidden Yard" || card.Place.AddressLine == "" {
		t.Fatalf("toEventCard() place = %+v, want populated place", card.Place)
	}
}

func TestMapPGErrorReturnsReferenceDetails(t *testing.T) {
	err := mapPGError(&pgconn.PgError{
		Code:           platformpostgres.CodeForeignKeyViolation,
		ConstraintName: "event_author_user_id_fkey",
	})

	if !errors.Is(err, platformerrors.ErrInvalidReference) {
		t.Fatalf("mapPGError() = %v, want ErrInvalidReference", err)
	}
	field, message, ok := platformerrors.InvalidReferenceDetails(err)
	if !ok || field != "actor" || message != "author user was not found" {
		t.Fatalf("InvalidReferenceDetails() = (%q, %q, %v)", field, message, ok)
	}
}
