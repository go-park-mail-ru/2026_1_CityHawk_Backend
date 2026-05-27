package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
)

func TestInMemoryRepositoryReadOperations(t *testing.T) {
	repo := NewInMemoryRepository([]placemodel.EventDetailsView{
		testEventDetails("event-1", "author-1", "Jazz Night", "music", "live"),
		testEventDetails("event-2", "author-2", "Art Expo", "art", "expo"),
	})

	home := repo.HomePayload(context.Background(), placemodel.HomeFilter{})
	if len(home.FeaturedEvents) != 2 || len(home.Categories) != 2 {
		t.Fatalf("unexpected home payload: %+v", home)
	}

	filteredHome := repo.HomePayload(context.Background(), placemodel.HomeFilter{City: "Moscow"})
	if len(filteredHome.FeaturedEvents) != 2 {
		t.Fatalf("unexpected city-filtered home payload: %+v", filteredHome)
	}
	emptyHome := repo.HomePayload(context.Background(), placemodel.HomeFilter{City: "Unknown"})
	if len(emptyHome.FeaturedEvents) != 0 || len(emptyHome.Categories) != 0 || len(emptyHome.Collections) != 0 {
		t.Fatalf("unexpected empty city-filtered home payload: %+v", emptyHome)
	}

	categories := repo.ListCategories(context.Background())
	if len(categories) != 2 {
		t.Fatalf("ListCategories() len = %d, want 2", len(categories))
	}

	tags := repo.ListTags(context.Background())
	if len(tags) != 2 {
		t.Fatalf("ListTags() len = %d, want 2", len(tags))
	}

	collections, err := repo.ListCollections(context.Background())
	if err != nil || len(collections) != 1 {
		t.Fatalf("ListCollections() = (%+v, %v), want single collection", collections, err)
	}

	collection, ok, err := repo.GetCollectionByID(context.Background(), "weekend-picks")
	if err != nil || !ok || len(collection.Events) == 0 {
		t.Fatalf("GetCollectionByID() = (%+v, %v, %v), want existing collection", collection, ok, err)
	}

	suggestions, err := repo.SearchSuggestions(context.Background(), "ja", 5)
	if err != nil {
		t.Fatalf("SearchSuggestions() error = %v", err)
	}
	if len(suggestions) == 0 || suggestions[0].Label != "Jazz Night" || suggestions[0].Type != "event" {
		t.Fatalf("SearchSuggestions() = %+v, want Jazz Night suggestion", suggestions)
	}

	filtered, total, err := repo.ListEvents(context.Background(), placemodel.EventListFilter{
		Query:      "jazz",
		CategoryID: "music",
		TagID:      "live",
		AuthorID:   "author-1",
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != "event-1" {
		t.Fatalf("ListEvents() = (%+v, %d), want only event-1", filtered, total)
	}
	filtered, total, err = repo.ListEvents(context.Background(), placemodel.EventListFilter{
		Query: "expo",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListEvents() by taxonomy query error = %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].ID != "event-2" {
		t.Fatalf("ListEvents() by taxonomy query = (%+v, %d), want event-2", filtered, total)
	}

	event, ok, err := repo.GetByID(context.Background(), "event-1", "author-1")
	if err != nil || !ok {
		t.Fatalf("GetByID() = (%+v, %v, %v), want existing event", event, ok, err)
	}
	if !event.IsOwner || len(event.Sessions) != 1 || event.Sessions[0].Place.Name != "place-1" {
		t.Fatalf("unexpected event details: %+v", event)
	}
}

func TestInMemoryRepositoryWriteOperations(t *testing.T) {
	repo := NewInMemoryRepository(nil)
	title := "Created Event"
	shortDescription := "Short description"
	fullDescription := "Full description"
	ageLimit := 18
	sourceURL := "https://events.example.com"
	categoryIDs := []string{"music"}
	tagIDs := []string{"live"}
	imageURLs := []string{"/uploads/created.png"}
	sessions := []placemodel.EventSessionInput{{
		PlaceID: "place-1",
		StartAt: time.Date(2026, time.April, 20, 19, 0, 0, 0, time.UTC),
		EndAt:   time.Date(2026, time.April, 20, 21, 0, 0, 0, time.UTC),
		Price:   1200,
	}}

	eventID, err := repo.CreateEvent(context.Background(), placemodel.EventWriteInput{
		AuthorUserID:     "author-1",
		Title:            &title,
		ShortDescription: &shortDescription,
		FullDescription:  &fullDescription,
		AgeLimit:         &ageLimit,
		SourceURL:        &sourceURL,
		CategoryIDs:      &categoryIDs,
		TagIDs:           &tagIDs,
		ImageURLs:        &imageURLs,
		Sessions:         &sessions,
	})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	created, ok, err := repo.GetByID(context.Background(), eventID, "author-1")
	if err != nil || !ok {
		t.Fatalf("GetByID() after CreateEvent = (%+v, %v, %v)", created, ok, err)
	}
	if created.Title != title || len(created.Images) != 1 || len(created.Sessions) != 1 {
		t.Fatalf("unexpected created event: %+v", created)
	}

	updatedTitle := "Updated Event"
	updatedImages := []string{"/uploads/updated.png"}
	updatedSessions := []placemodel.EventSessionInput{{
		PlaceID: "place-2",
		StartAt: time.Date(2026, time.April, 21, 18, 0, 0, 0, time.UTC),
		EndAt:   time.Date(2026, time.April, 21, 20, 0, 0, 0, time.UTC),
		Price:   2000,
	}}

	updated, err := repo.UpdateEvent(context.Background(), placemodel.EventWriteInput{
		ID:             eventID,
		AuthorUserID:   "author-1",
		Title:          &updatedTitle,
		ImageURLs:      &updatedImages,
		Sessions:       &updatedSessions,
		ClearSourceURL: true,
	})
	if err != nil || !updated {
		t.Fatalf("UpdateEvent() = (%v, %v), want success", updated, err)
	}

	afterUpdate, ok, err := repo.GetByID(context.Background(), eventID, "author-1")
	if err != nil || !ok {
		t.Fatalf("GetByID() after UpdateEvent = (%+v, %v, %v)", afterUpdate, ok, err)
	}
	if afterUpdate.Title != updatedTitle || afterUpdate.SourceURL != nil || afterUpdate.Sessions[0].Place.ID != "place-2" {
		t.Fatalf("unexpected updated event: %+v", afterUpdate)
	}

	updated, err = repo.UpdateEvent(context.Background(), placemodel.EventWriteInput{
		ID:           eventID,
		AuthorUserID: "another-user",
		Title:        &updatedTitle,
	})
	if updated || !errors.Is(err, platformerrors.ErrForbidden) {
		t.Fatalf("unauthorized UpdateEvent() = (%v, %v), want ErrForbidden", updated, err)
	}

	deleted, err := repo.DeleteEvent(context.Background(), eventID, "another-user")
	if deleted || !errors.Is(err, platformerrors.ErrForbidden) {
		t.Fatalf("unauthorized DeleteEvent() = (%v, %v), want ErrForbidden", deleted, err)
	}

	deleted, err = repo.DeleteEvent(context.Background(), eventID, "author-1")
	if err != nil || !deleted {
		t.Fatalf("DeleteEvent() = (%v, %v), want success", deleted, err)
	}

	_, ok, err = repo.GetByID(context.Background(), eventID, "author-1")
	if err != nil || ok {
		t.Fatalf("GetByID() after delete = (%v, %v), want not found", ok, err)
	}
}

func TestInMemoryRepositoryHelpers(t *testing.T) {
	event := testEventDetails("event-1", "author-1", "Jazz Night", "music", "live")

	if got := firstImageURL(event); got != "/uploads/event-1.png" {
		t.Fatalf("firstImageURL() = %q, want /uploads/event-1.png", got)
	}
	if got := firstImageURL(placemodel.EventDetailsView{}); got != "" {
		t.Fatalf("firstImageURL() on empty event = %q, want empty string", got)
	}

	homeTags := toHomeTags(event.Tags)
	if len(homeTags) != 1 || homeTags[0].ID != "live" {
		t.Fatalf("toHomeTags() = %+v, want single mapped tag", homeTags)
	}

	card := toCard(event)
	if card.ID != "event-1" || card.NextSession == nil || card.NextSession.Place.Name != "place-1" {
		t.Fatalf("toCard() = %+v, want populated card", card)
	}

	if !matchesFilter(event, placemodel.EventListFilter{Query: "jazz", CategoryID: "music", TagID: "live", AuthorID: "author-1"}) {
		t.Fatal("matchesFilter() = false, want true")
	}
	if matchesFilter(event, placemodel.EventListFilter{Query: "opera"}) {
		t.Fatal("matchesFilter() = true for mismatched query, want false")
	}
	if !hasTaxonomyID(event.Tags, "live") || hasTaxonomyID(event.Tags, "expo") {
		t.Fatal("hasTaxonomyID() returned unexpected result")
	}

	if got := derefStringSlice(nil); got != nil {
		t.Fatalf("derefStringSlice(nil) = %+v, want nil", got)
	}
	if got := derefSessions(nil); got != nil {
		t.Fatalf("derefSessions(nil) = %+v, want nil", got)
	}

	next := 99
	if id := generateID("event", &next); id != "event-100" {
		t.Fatalf("generateID() = %q, want event-100", id)
	}
	ids := reserveIDs("image", &next, 2)
	if len(ids) != 2 || ids[0] != "image-101" || ids[1] != "image-102" {
		t.Fatalf("reserveIDs() = %+v, want sequential ids", ids)
	}

	taxonomyItems := buildTaxonomyItems([]string{"music", "live"})
	if len(taxonomyItems) != 2 || taxonomyItems[0].Slug != "music" || taxonomyItems[1].Name != "live" {
		t.Fatalf("buildTaxonomyItems() = %+v, want mapped taxonomy items", taxonomyItems)
	}

	imageItems := buildImageItems([]string{"/uploads/one.png"}, &next)
	if len(imageItems) != 1 || imageItems[0].ImageURL != "/uploads/one.png" {
		t.Fatalf("buildImageItems() = %+v, want mapped image items", imageItems)
	}

	sessionItems := buildSessionItems([]placemodel.EventSessionInput{{
		PlaceID: "place-9",
		StartAt: time.Date(2026, time.April, 30, 18, 0, 0, 0, time.UTC),
		EndAt:   time.Date(2026, time.April, 30, 20, 0, 0, 0, time.UTC),
		Price:   900,
	}}, &next)
	if len(sessionItems) != 1 || sessionItems[0].Place.ID != "place-9" || sessionItems[0].Price != 900 {
		t.Fatalf("buildSessionItems() = %+v, want mapped session items", sessionItems)
	}

	if got := derefStringLocal(nil); got != "" {
		t.Fatalf("derefStringLocal(nil) = %q, want empty string", got)
	}
	if got := derefIntLocal(nil); got != 0 {
		t.Fatalf("derefIntLocal(nil) = %d, want 0", got)
	}
}

func testEventDetails(id, authorID, title, categoryID, tagID string) placemodel.EventDetailsView {
	startAt := time.Date(2026, time.April, 15, 19, 0, 0, 0, time.UTC)
	return placemodel.EventDetailsView{
		ID:               id,
		Title:            title,
		ShortDescription: title + " short",
		FullDescription:  title + " full",
		Author: placemodel.EventAuthorView{
			ID:       authorID,
			Username: "author",
		},
		Categories: []placemodel.EventTaxonomyItem{{
			ID:   categoryID,
			Name: categoryID,
			Slug: categoryID,
		}},
		Tags: []placemodel.EventTaxonomyItem{{
			ID:   tagID,
			Name: tagID,
			Slug: tagID,
		}},
		Images: []placemodel.EventImageView{{
			ID:       "image-1",
			ImageURL: "/uploads/" + id + ".png",
		}},
		Sessions: []placemodel.EventSessionView{{
			ID:      "session-1",
			StartAt: startAt,
			EndAt:   startAt.Add(2 * time.Hour),
			Price:   1000,
			Place: placemodel.EventSessionPlaceView{
				ID:          "place-1",
				Name:        "place-1",
				AddressLine: "Address for place-1",
				City: placemodel.EventSessionPlaceCityView{
					ID:          "city-1",
					Name:        "Moscow",
					CountryName: "Russia",
					Timezone:    "Europe/Moscow",
				},
			},
		}},
		CreatedAt: startAt,
		UpdatedAt: startAt,
	}
}
