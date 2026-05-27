package repository

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestPostgresRepositoryReadMethodsWithFakeDB(t *testing.T) {
	now := time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC)
	db := &fakePlaceDB{
		rows: map[string][][]any{
			"FROM category c ORDER BY":                                  {{"cat-1", "Music"}},
			"FROM tag t ORDER BY":                                       {{"tag-1", "Jazz", "genre"}},
			"FROM city c ORDER BY":                                      {{"city-1", "Moscow", "Russia", "Europe/Moscow"}},
			"FROM collection c\n\t\tLEFT JOIN":                          {{"collection-1", "Weekend", "Best", "/uploads/collection.png", true}},
			"FROM candidates":                                           {{"event-1", "event", "Jazz night", "Jazz night", nil, false}},
			"FROM filtered_events filtered":                             {{"event-1", "Jazz night", "Short", "/uploads/event.png", []string{"tag-1"}, []string{"Jazz"}, &now, "place-1", "Hall", "Lenina 1", 55.7, 37.6, "city-1", "Moscow", "Russia", "Europe/Moscow", 1, true, 3}},
			"FROM event e\n\t\tLEFT JOIN first_image":                   {{"event-1", "Jazz night", "/uploads/event.png", []string{"tag-1"}, []string{"Jazz"}, &now, "Hall", "Lenina 1"}},
			"SELECT c.id::text, c.name\n\t\tFROM category c\n\t\tWHERE": {{"cat-1", "Music"}},
			"SELECT\n\t\t\tc.id::text":                                  {{"collection-1", "Weekend", "Best", "/uploads/collection.png"}},
			"FROM event_category ec":                                    {{"cat-1", "Music"}},
			"FROM event_tag et":                                         {{"tag-1", "Jazz"}},
			"FROM event_image WHERE":                                    {{"img-1", "/uploads/event.png"}},
			"FETCH event sessions":                                      {{"session-1", now, now.Add(time.Hour), 1200, "place-1", "Hall", "Lenina 1", 55.7, 37.6, "city-1", "Moscow", "Russia", "Europe/Moscow"}},
		},
		row: map[string][]any{
			"FROM collection c\n\t\tLEFT JOIN first_image": {"collection-1", "Weekend", "Best", "/uploads/collection.png", true},
			"FROM event e\n\t\tJOIN user_account":          {"event-1", "Jazz night", "Short", "Full", 18, stringPtr("https://events.example.com"), "user-1", "alice", stringPtr("/uploads/avatar.png"), now, now.Add(time.Hour), true, true},
			"FROM event_place ep":                          {"place-1", "Hall", "Lenina 1", 55.7, 37.6, "city-1", "Moscow", "Russia", "Europe/Moscow"},
		},
	}
	repo := &PostgresRepository{pool: db}

	if got := repo.ListCategories(context.Background()); len(got) != 1 || got[0].Slug != "music" {
		t.Fatalf("ListCategories() = %+v", got)
	}
	if got := repo.ListTags(context.Background()); len(got) != 1 || got[0].Slug != "jazz" || got[0].Group != "genre" {
		t.Fatalf("ListTags() = %+v", got)
	}
	if got := repo.ListCities(context.Background()); len(got) != 1 || got[0].Name != "Moscow" {
		t.Fatalf("ListCities() = %+v", got)
	}
	if got := repo.ListFeaturedEvents(context.Background(), 1, "Moscow"); len(got) != 1 || got[0].NextSession.Place.Name != "Hall" {
		t.Fatalf("ListFeaturedEvents() = %+v", got)
	}
	if got := repo.ListHomeCategories(context.Background(), 1, "Moscow"); len(got) != 1 {
		t.Fatalf("ListHomeCategories() = %+v", got)
	}
	if got := repo.ListHomeCollections(context.Background(), 1, "Moscow"); len(got) != 1 {
		t.Fatalf("ListHomeCollections() = %+v", got)
	}
	if got, err := repo.ListCollections(context.Background()); err != nil || len(got) != 1 {
		t.Fatalf("ListCollections() = (%+v, %v)", got, err)
	}
	if got, ok, err := repo.GetCollectionByID(context.Background(), "collection-1"); err != nil || !ok || got.ID != "collection-1" {
		t.Fatalf("GetCollectionByID() = (%+v, %v, %v)", got, ok, err)
	}
	if got, err := repo.SearchSuggestions(context.Background(), "ja", 5); err != nil || len(got) != 1 {
		t.Fatalf("SearchSuggestions() = (%+v, %v)", got, err)
	}
	if got, total, err := repo.ListEvents(context.Background(), placemodel.EventListFilter{Limit: 10}); err != nil || total != 1 || len(got) != 1 {
		t.Fatalf("ListEvents() = (%+v, %d, %v)", got, total, err)
	}
	if got, ok, err := repo.GetByID(context.Background(), "event-1", "user-1"); err != nil || !ok || len(got.Sessions) != 1 || len(got.Images) != 1 {
		t.Fatalf("GetByID() = (%+v, %v, %v)", got, ok, err)
	}
}

func TestPostgresRepositoryWriteMethodsWithFakeTx(t *testing.T) {
	now := time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC)
	tx := &fakePlaceTx{}
	repo := &PostgresRepository{pool: &fakePlaceDB{tx: tx}}
	title := "Event"
	shortDescription := "Short"
	fullDescription := "Full"
	ageLimit := 18
	sourceURL := "https://events.example.com"
	categories := []string{"cat-1"}
	tags := []string{"tag-1"}
	images := []string{"/uploads/event.png"}
	placeID := "place-1"
	sessions := []placemodel.EventSessionInput{{PlaceID: "place-1", StartAt: now, EndAt: now.Add(time.Hour), Price: 100}}

	id, err := repo.CreateEvent(context.Background(), placemodel.EventWriteInput{
		AuthorUserID: "user-1", Title: &title, ShortDescription: &shortDescription, FullDescription: &fullDescription,
		AgeLimit: &ageLimit, SourceURL: &sourceURL, CategoryIDs: &categories, TagIDs: &tags, ImageURLs: &images, PlaceID: &placeID, Sessions: &sessions,
	})
	if err != nil || id != "event-created" || !tx.committed {
		t.Fatalf("CreateEvent() = (%q, %v), committed=%v", id, err, tx.committed)
	}

	tx.committed = false
	ok, err := repo.UpdateEvent(context.Background(), placemodel.EventWriteInput{
		ID: "event-created", AuthorUserID: "user-1", Title: &title, ShortDescription: &shortDescription, FullDescription: &fullDescription,
		AgeLimit: &ageLimit, SourceURL: &sourceURL, ClearSourceURL: true, CategoryIDs: &categories, TagIDs: &tags, ImageURLs: &images, PlaceID: &placeID, Sessions: &sessions,
	})
	if err != nil || !ok || !tx.committed {
		t.Fatalf("UpdateEvent() = (%v, %v), committed=%v", ok, err, tx.committed)
	}

	tx.committed = false
	ok, err = repo.DeleteEvent(context.Background(), "event-created", "user-1")
	if err != nil || !ok || !tx.committed {
		t.Fatalf("DeleteEvent() = (%v, %v), committed=%v", ok, err, tx.committed)
	}
}

type fakePlaceDB struct {
	rows map[string][][]any
	row  map[string][]any
	tx   pgx.Tx
}

func (f *fakePlaceDB) Begin(context.Context) (pgx.Tx, error) { return f.tx, nil }
func (f *fakePlaceDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (f *fakePlaceDB) Query(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
	if strings.Contains(sql, "FROM filtered_events filtered") {
		return &fakePlaceRows{rows: f.rows["FROM filtered_events filtered"], index: -1}, nil
	}
	if strings.Contains(sql, "FROM event e\n\t\tLEFT JOIN first_image fi") {
		return &fakePlaceRows{rows: f.rows["FROM event e\n\t\tLEFT JOIN first_image"], index: -1}, nil
	}
	if strings.Contains(sql, "FROM event_session es") && strings.Contains(sql, "WHERE es.event_id = $1") {
		return &fakePlaceRows{rows: f.rows["FETCH event sessions"], index: -1}, nil
	}
	if strings.Contains(sql, "FROM event_category ec") && strings.Contains(sql, "WHERE ec.event_id = $1") {
		return &fakePlaceRows{rows: f.rows["FROM event_category ec"], index: -1}, nil
	}
	if strings.Contains(sql, "FROM event_tag et") && strings.Contains(sql, "WHERE et.event_id = $1") {
		return &fakePlaceRows{rows: f.rows["FROM event_tag et"], index: -1}, nil
	}
	for pattern, rows := range f.rows {
		if strings.Contains(sql, pattern) {
			return &fakePlaceRows{rows: rows, index: -1}, nil
		}
	}
	return &fakePlaceRows{index: -1}, nil
}
func (f *fakePlaceDB) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	for pattern, row := range f.row {
		if strings.Contains(sql, pattern) {
			return fakePlaceRow(row)
		}
	}
	return fakePlaceRow(nil)
}

type fakePlaceRow []any

func (r fakePlaceRow) Scan(dest ...any) error {
	if r == nil {
		return pgx.ErrNoRows
	}
	return scanFakeValues([]any(r), dest...)
}

type fakePlaceRows struct {
	rows   [][]any
	index  int
	closed bool
}

func (r *fakePlaceRows) Close()                                       { r.closed = true }
func (r *fakePlaceRows) Err() error                                   { return nil }
func (r *fakePlaceRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakePlaceRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakePlaceRows) Next() bool {
	r.index++
	if r.index >= len(r.rows) {
		r.Close()
		return false
	}
	return true
}
func (r *fakePlaceRows) Scan(dest ...any) error { return scanFakeValues(r.rows[r.index], dest...) }
func (r *fakePlaceRows) Values() ([]any, error) { return r.rows[r.index], nil }
func (r *fakePlaceRows) RawValues() [][]byte    { return nil }
func (r *fakePlaceRows) Conn() *pgx.Conn        { return nil }

func scanFakeValues(values []any, dest ...any) error {
	for i := range dest {
		if i >= len(values) {
			break
		}
		target := reflect.ValueOf(dest[i])
		if target.Kind() != reflect.Pointer || target.IsNil() {
			continue
		}
		value := reflect.ValueOf(values[i])
		if !value.IsValid() {
			target.Elem().SetZero()
			continue
		}
		if value.Type().AssignableTo(target.Elem().Type()) {
			target.Elem().Set(value)
			continue
		}
		if value.Type().ConvertibleTo(target.Elem().Type()) {
			target.Elem().Set(value.Convert(target.Elem().Type()))
		}
	}
	return nil
}

func stringPtr(value string) *string {
	return &value
}

type fakePlaceTx struct {
	committed bool
}

func (f *fakePlaceTx) Begin(context.Context) (pgx.Tx, error) { return f, nil }
func (f *fakePlaceTx) Commit(context.Context) error {
	f.committed = true
	return nil
}
func (f *fakePlaceTx) Rollback(context.Context) error { return nil }
func (f *fakePlaceTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakePlaceTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakePlaceTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (f *fakePlaceTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakePlaceTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("DELETE 1"), nil
}
func (f *fakePlaceTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return &fakePlaceRows{index: -1}, nil
}
func (f *fakePlaceTx) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	if strings.Contains(sql, "INSERT INTO event") {
		return fakePlaceRow([]any{"event-created"})
	}
	if strings.Contains(sql, "SELECT author_user_id") {
		return fakePlaceRow([]any{"user-1"})
	}
	return fakePlaceRow(nil)
}
func (f *fakePlaceTx) Conn() *pgx.Conn { return nil }
