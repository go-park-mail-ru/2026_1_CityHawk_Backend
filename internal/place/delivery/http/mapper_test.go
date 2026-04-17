package http

import (
	"encoding/json"
	"testing"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
)

func TestOptionalStringUnmarshalJSON(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		var value optionalString
		if err := json.Unmarshal([]byte(`"https://example.com"`), &value); err != nil {
			t.Fatalf("UnmarshalJSON() error = %v", err)
		}
		if !value.Set || value.Value == nil || *value.Value != "https://example.com" {
			t.Fatalf("unexpected optionalString: %+v", value)
		}
	})

	t.Run("null value", func(t *testing.T) {
		var value optionalString
		if err := json.Unmarshal([]byte(`null`), &value); err != nil {
			t.Fatalf("UnmarshalJSON() error = %v", err)
		}
		if !value.Set || value.Value != nil {
			t.Fatalf("unexpected optionalString: %+v", value)
		}
	})
}

func TestResponseMappers(t *testing.T) {
	now := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)
	sourceURL := "https://events.example.com"
	avatarURL := "/uploads/avatar.png"
	publicAvatarURL := "http://cityhawk.ru:8080/uploads/avatar.png"

	details := placemodel.EventDetailsView{
		ID:               "event-1",
		Title:            `<script>alert("x")</script>Concert`,
		ShortDescription: "Short <b>desc</b>",
		FullDescription:  "Full <i>desc</i>",
		AgeLimit:         18,
		SourceURL:        &sourceURL,
		Author: placemodel.EventAuthorView{
			ID:        "user-1",
			Username:  `Alice <admin>`,
			AvatarURL: &avatarURL,
		},
		Categories: []placemodel.EventTaxonomyItem{
			{ID: "cat-1", Name: `Music & Fun`, Slug: "music"},
		},
		Tags: []placemodel.EventTaxonomyItem{
			{ID: "tag-1", Name: `Live <show>`, Slug: "live"},
		},
		Images: []placemodel.EventImageView{
			{ID: "img-1", ImageURL: "/uploads/event.png"},
		},
		Sessions: []placemodel.EventSessionView{
			{
				ID:      "sess-1",
				StartAt: now,
				EndAt:   now.Add(2 * time.Hour),
				Price:   1500,
				Place: placemodel.EventSessionPlaceView{
					ID:          "place-1",
					Name:        `Main <Hall>`,
					AddressLine: `Lenina <1>`,
					Latitude:    55.75,
					Longitude:   37.61,
					City: placemodel.EventSessionPlaceCityView{
						ID:          "city-1",
						Name:        `Moscow <RU>`,
						CountryName: `Russia <RU>`,
						Timezone:    "Europe/Moscow",
					},
				},
			},
		},
		CreatedAt:  now,
		UpdatedAt:  now.Add(time.Hour),
		IsFavorite: true,
		IsOwner:    true,
	}

	t.Run("event details", func(t *testing.T) {
		resp := toEventDetailsResponse(details)
		if resp.Title == details.Title {
			t.Fatalf("Title was not escaped: %q", resp.Title)
		}
		if resp.Author.AvatarURL == nil || *resp.Author.AvatarURL != publicAvatarURL {
			t.Fatalf("unexpected author avatar: %+v", resp.Author.AvatarURL)
		}
		if len(resp.Images) != 1 || resp.Images[0].ImageURL != "http://cityhawk.ru:8080/uploads/event.png" {
			t.Fatalf("unexpected image URLs: %+v", resp.Images)
		}
		if len(resp.Sessions) != 1 || resp.Sessions[0].Place.City.Timezone != "Europe/Moscow" {
			t.Fatalf("unexpected sessions response: %+v", resp.Sessions)
		}
		if resp.CreatedAt != "2026-04-15T10:00:00Z" || resp.UpdatedAt != "2026-04-15T11:00:00Z" {
			t.Fatalf("unexpected timestamps: created=%q updated=%q", resp.CreatedAt, resp.UpdatedAt)
		}
	})

	t.Run("event list and collection details", func(t *testing.T) {
		card := placemodel.EventCardView{
			ID:               "event-1",
			Title:            `Concert <main>`,
			ShortDescription: `Loud <fun>`,
			CoverImageURL:    "/uploads/cover.png",
			Tags: []placemodel.EventTaxonomyItem{
				{ID: "tag-1", Name: `Live <show>`, Slug: "live"},
			},
			NextSession: &placemodel.EventCardNextSession{
				StartAt: now,
				Place: placemodel.EventCardNextSessionPlace{
					Name:        `Main <Hall>`,
					AddressLine: `Lenina <1>`,
				},
			},
		}

		listResp := toEventListResponse([]placemodel.EventCardView{card}, 1, 20, 0)
		if listResp.Total != 1 || len(listResp.Items) != 1 {
			t.Fatalf("unexpected event list response: %+v", listResp)
		}
		if listResp.Items[0].Title == card.Title || listResp.Items[0].NextSession == nil {
			t.Fatalf("unexpected event card response: %+v", listResp.Items[0])
		}
		if listResp.Items[0].CoverImageURL != "http://cityhawk.ru:8080/uploads/cover.png" {
			t.Fatalf("unexpected event cover URL: %q", listResp.Items[0].CoverImageURL)
		}

		collectionResp := toCollectionDetailsResponse(placemodel.CollectionDetailsView{
			ID:          "collection-1",
			Title:       `Weekend <Picks>`,
			Description: `Best <events>`,
			ImageURL:    "/uploads/collection.png",
			IsPublic:    true,
			Events:      []placemodel.EventCardView{card},
		})
		if collectionResp.Title == `Weekend <Picks>` || len(collectionResp.Events) != 1 {
			t.Fatalf("unexpected collection response: %+v", collectionResp)
		}
		if collectionResp.ImageURL != "http://cityhawk.ru:8080/uploads/collection.png" {
			t.Fatalf("unexpected collection image URL: %q", collectionResp.ImageURL)
		}
	})

	t.Run("home, taxonomies and search suggestions", func(t *testing.T) {
		categoriesResp := toCategoriesResponse([]placemodel.HomeCategory{{ID: "cat-1", Name: `Music <fest>`, Slug: "music"}})
		tagsResp := toTagsResponse([]placemodel.HomeTag{{ID: "tag-1", Name: `Live <show>`, Slug: "live"}})
		collectionsResp := toCollectionsResponse([]placemodel.CollectionCardView{{
			ID:          "collection-1",
			Title:       `Weekend <Picks>`,
			Description: `Best <events>`,
			ImageURL:    "/uploads/collection.png",
			IsPublic:    true,
		}})
		searchResp := toSearchSuggestionsResponse([]placemodel.SearchSuggestion{{Name: `Jazz <Night>`}})
		homeResp := toHomePayloadResponse(placemodel.HomePayload{
			FeaturedEvents: []placemodel.HomeFeaturedEvent{{
				ID:            "event-1",
				Title:         `Concert <main>`,
				CoverImageURL: "/uploads/cover.png",
				Tags:          []placemodel.HomeTag{{ID: "tag-1", Name: `Live <show>`, Slug: "live"}},
				NextSession: placemodel.HomeNextSession{
					StartAt: now,
					Place: placemodel.HomeNextSessionPlace{
						Name:        `Main <Hall>`,
						AddressLine: `Lenina <1>`,
					},
				},
			}},
			Categories: []placemodel.HomeCategory{{ID: "cat-1", Name: `Music <fest>`, Slug: "music"}},
			Collections: []placemodel.HomeCollection{{
				ID:          "collection-1",
				Title:       `Weekend <Picks>`,
				Description: `Best <events>`,
				ImageURL:    "/uploads/collection.png",
			}},
		})

		if categoriesResp.Items[0].Name == `Music <fest>` {
			t.Fatalf("category was not escaped: %+v", categoriesResp.Items[0])
		}
		if tagsResp.Items[0].Name == `Live <show>` {
			t.Fatalf("tag was not escaped: %+v", tagsResp.Items[0])
		}
		if collectionsResp.Items[0].Title == `Weekend <Picks>` {
			t.Fatalf("collection was not escaped: %+v", collectionsResp.Items[0])
		}
		if searchResp.Items[0] == `Jazz <Night>` {
			t.Fatalf("search suggestion was not escaped: %+v", searchResp.Items)
		}
		if homeResp.FeaturedEvents[0].Title == `Concert <main>` || homeResp.Collections[0].Title == `Weekend <Picks>` {
			t.Fatalf("home payload was not escaped: %+v", homeResp)
		}
		if homeResp.FeaturedEvents[0].CoverImageURL != "http://cityhawk.ru:8080/uploads/cover.png" || homeResp.Collections[0].ImageURL != "http://cityhawk.ru:8080/uploads/collection.png" {
			t.Fatalf("unexpected home image URLs: %+v", homeResp)
		}
	})
}
