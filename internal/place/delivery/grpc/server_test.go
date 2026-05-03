package grpc

import (
	"context"
	"testing"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
	eventsv1 "cityhawk/backend/pkg/pb/events/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServerEventsFlow(t *testing.T) {
	events := newFakeEvents()
	server := NewServer(events, fakeLookup{})
	actor := &commonv1.UserContext{UserId: "user-1", Authenticated: true}

	if home, err := server.GetHome(context.Background(), &eventsv1.GetHomeRequest{City: stringPtr("Moscow")}); err != nil || len(home.GetFeaturedEvents()) != 1 {
		t.Fatalf("GetHome() = (%+v, %v)", home, err)
	}
	if list, err := server.ListEvents(context.Background(), &eventsv1.ListEventsRequest{Page: &commonv1.PageRequest{Limit: 10}, ViewerContext: actor}); err != nil || len(list.GetItems()) != 1 {
		t.Fatalf("ListEvents() = (%+v, %v)", list, err)
	}
	if got, err := server.GetEvent(context.Background(), &eventsv1.GetEventRequest{EventId: "event-1", ViewerContext: actor}); err != nil || got.GetId() != "event-1" {
		t.Fatalf("GetEvent() = (%+v, %v)", got, err)
	}

	sourceURL := "https://events.example.com"
	created, err := server.CreateEvent(context.Background(), &eventsv1.CreateEventRequest{
		Actor:            actor,
		Title:            " Event ",
		ShortDescription: " Short ",
		FullDescription:  " Full ",
		AgeLimit:         18,
		SourceUrl:        &sourceURL,
		CategoryIds:      []string{" cat-1 "},
		TagIds:           []string{" tag-1 "},
		ImageUrls:        []string{" /uploads/event.png "},
		Sessions: []*eventsv1.EventSessionInput{{
			PlaceId: " place-1 ",
			StartAt: timestamppb.New(events.now),
			EndAt:   timestamppb.New(events.now.Add(time.Hour)),
			Price:   100,
		}},
	})
	if err != nil || created.GetId() == "" || events.created.Title == nil || *events.created.Title != "Event" {
		t.Fatalf("CreateEvent() = (%+v, %v), input=%+v", created, err, events.created)
	}

	updated, err := server.UpdateEvent(context.Background(), &eventsv1.UpdateEventRequest{
		Actor:              actor,
		EventId:            "event-1",
		Title:              stringPtr(" Updated "),
		ReplaceCategoryIds: true,
		CategoryIds:        []string{"cat-1"},
		ReplaceSessions:    true,
		Sessions: []*eventsv1.EventSessionInput{{
			PlaceId: "place-1",
			StartAt: timestamppb.New(events.now),
			EndAt:   timestamppb.New(events.now.Add(time.Hour)),
		}},
	})
	if err != nil || updated.GetId() != "event-1" || events.updated.Title == nil || *events.updated.Title != "Updated" {
		t.Fatalf("UpdateEvent() = (%+v, %v), input=%+v", updated, err, events.updated)
	}

	if deleted, err := server.DeleteEvent(context.Background(), &eventsv1.DeleteEventRequest{Actor: actor, EventId: "event-1"}); err != nil || !deleted.GetOk() {
		t.Fatalf("DeleteEvent() = (%+v, %v)", deleted, err)
	}
	if cats, err := server.ListCategories(context.Background(), &eventsv1.ListTaxonomyRequest{}); err != nil || len(cats.GetItems()) != 1 {
		t.Fatalf("ListCategories() = (%+v, %v)", cats, err)
	}
	if tags, err := server.ListTags(context.Background(), &eventsv1.ListTaxonomyRequest{}); err != nil || len(tags.GetItems()) != 1 {
		t.Fatalf("ListTags() = (%+v, %v)", tags, err)
	}
	if cities, err := server.ListCities(context.Background(), &eventsv1.ListCitiesRequest{}); err != nil || len(cities.GetItems()) != 1 {
		t.Fatalf("ListCities() = (%+v, %v)", cities, err)
	}
	if collections, err := server.ListCollections(context.Background(), &eventsv1.ListCollectionsRequest{Page: &commonv1.PageRequest{Limit: 10}}); err != nil || len(collections.GetItems()) != 1 {
		t.Fatalf("ListCollections() = (%+v, %v)", collections, err)
	}
	if collection, err := server.GetCollection(context.Background(), &eventsv1.GetCollectionRequest{CollectionId: "collection-1"}); err != nil || collection.GetId() != "collection-1" {
		t.Fatalf("GetCollection() = (%+v, %v)", collection, err)
	}
	if suggestions, err := server.SearchSuggestions(context.Background(), &eventsv1.SearchSuggestionsRequest{Query: "ja", Limit: 5}); err != nil || len(suggestions.GetItems()) != 1 {
		t.Fatalf("SearchSuggestions() = (%+v, %v)", suggestions, err)
	}
	if places, err := server.SuggestPlaces(context.Background(), &eventsv1.SuggestPlacesRequest{Query: "hall", Limit: 3}); err != nil || len(places.GetItems()) != 1 {
		t.Fatalf("SuggestPlaces() = (%+v, %v)", places, err)
	}
	if place, err := server.ResolvePlace(context.Background(), &eventsv1.ResolvePlaceRequest{Token: "token"}); err != nil || place.GetId() != "place-1" {
		t.Fatalf("ResolvePlace() = (%+v, %v)", place, err)
	}
}

type fakeEvents struct {
	now     time.Time
	details placemodel.EventDetailsView
	card    placemodel.EventCardView
	created placemodel.EventWriteInput
	updated placemodel.EventWriteInput
}

func newFakeEvents() *fakeEvents {
	now := time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC)
	details := placemodel.EventDetailsView{
		ID:               "event-1",
		Title:            "Jazz night",
		ShortDescription: "Short",
		FullDescription:  "Full",
		AgeLimit:         18,
		Author:           placemodel.EventAuthorView{ID: "user-1", Username: "alice"},
		Categories:       []placemodel.EventTaxonomyItem{{ID: "cat-1", Name: "Music", Slug: "music"}},
		Tags:             []placemodel.EventTaxonomyItem{{ID: "tag-1", Name: "Jazz", Slug: "jazz"}},
		Images:           []placemodel.EventImageView{{ID: "img-1", ImageURL: "/uploads/event.png"}},
		Sessions: []placemodel.EventSessionView{{
			ID:      "session-1",
			StartAt: now,
			EndAt:   now.Add(time.Hour),
			Place: placemodel.EventSessionPlaceView{
				ID: "place-1", Name: "Hall", AddressLine: "Lenina 1", Latitude: 55.7, Longitude: 37.6,
				City: placemodel.EventSessionPlaceCityView{ID: "city-1", Name: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow"},
			},
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
	card := placemodel.EventCardView{ID: "event-1", Title: "Jazz night", ShortDescription: "Short", CoverImageURL: "/uploads/event.png", Tags: details.Tags, NextSession: &placemodel.EventCardNextSession{StartAt: now, Place: placemodel.EventCardNextSessionPlace{Name: "Hall", AddressLine: "Lenina 1"}}}
	return &fakeEvents{now: now, details: details, card: card}
}

func (f *fakeEvents) HomePayload(context.Context, placemodel.HomeFilter) placemodel.HomePayload {
	return placemodel.HomePayload{FeaturedEvents: []placemodel.HomeFeaturedEvent{{ID: "event-1", Title: "Jazz night", CoverImageURL: "/uploads/event.png", Tags: []placemodel.HomeTag{{ID: "tag-1", Name: "Jazz", Slug: "jazz"}}, NextSession: placemodel.HomeNextSession{StartAt: f.now, Place: placemodel.HomeNextSessionPlace{Name: "Hall"}}}}, Categories: []placemodel.HomeCategory{{ID: "cat-1", Name: "Music", Slug: "music"}}, Collections: []placemodel.HomeCollection{{ID: "collection-1", Title: "Weekend", ImageURL: "/uploads/collection.png"}}}
}
func (f *fakeEvents) ListCategories(context.Context) []placemodel.HomeCategory {
	return []placemodel.HomeCategory{{ID: "cat-1", Name: "Music", Slug: "music"}}
}
func (f *fakeEvents) ListTags(context.Context) []placemodel.HomeTag {
	return []placemodel.HomeTag{{ID: "tag-1", Name: "Jazz", Slug: "jazz"}}
}
func (f *fakeEvents) ListCities(context.Context) []placemodel.City {
	return []placemodel.City{{ID: "city-1", Name: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow"}}
}
func (f *fakeEvents) ListCollections(context.Context) ([]placemodel.CollectionCardView, error) {
	return []placemodel.CollectionCardView{{ID: "collection-1", Title: "Weekend", ImageURL: "/uploads/collection.png", IsPublic: true}}, nil
}
func (f *fakeEvents) GetCollectionByID(context.Context, string) (placemodel.CollectionDetailsView, bool, error) {
	return placemodel.CollectionDetailsView{ID: "collection-1", Title: "Weekend", ImageURL: "/uploads/collection.png", IsPublic: true, Events: []placemodel.EventCardView{f.card}}, true, nil
}
func (f *fakeEvents) SearchSuggestions(context.Context, string, int) ([]placemodel.SearchSuggestion, error) {
	return []placemodel.SearchSuggestion{{ID: "event-1", Type: "event", Label: "Jazz night"}}, nil
}
func (f *fakeEvents) ListEvents(context.Context, placemodel.EventListFilter) ([]placemodel.EventCardView, int, error) {
	return []placemodel.EventCardView{f.card}, 1, nil
}
func (f *fakeEvents) GetByID(context.Context, string, string) (placemodel.EventDetailsView, bool, error) {
	return f.details, true, nil
}
func (f *fakeEvents) CreateEvent(_ context.Context, input placemodel.EventWriteInput) (string, error) {
	f.created = input
	return "event-created", nil
}
func (f *fakeEvents) UpdateEvent(_ context.Context, input placemodel.EventWriteInput) (bool, error) {
	f.updated = input
	return true, nil
}
func (f *fakeEvents) DeleteEvent(context.Context, string, string) (bool, error) { return true, nil }

type fakeLookup struct{}

func (fakeLookup) Suggest(context.Context, string, int) ([]placemodel.PlaceSuggestion, error) {
	return []placemodel.PlaceSuggestion{{Token: "token", Label: "Hall", Name: "Hall", AddressLine: "Lenina 1", CityName: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow", Latitude: 55.7, Longitude: 37.6, PhotonSource: placemodel.PlaceSuggestionSource{OSMID: 1, OSMType: "N"}}}, nil
}
func (fakeLookup) Resolve(context.Context, placemodel.PlaceResolveInput) (placemodel.PlaceResolved, error) {
	return placemodel.PlaceResolved{ID: "place-1", CityID: "city-1", Name: "Hall", AddressLine: "Lenina 1", CityName: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow", Latitude: 55.7, Longitude: 37.6}, nil
}

func stringPtr(value string) *string { return &value }
