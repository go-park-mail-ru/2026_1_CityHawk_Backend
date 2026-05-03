package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	"cityhawk/backend/internal/platform/httpx"
	socialmodel "cityhawk/backend/internal/social/model"
)

func TestHandlerSocialHTTPFlow(t *testing.T) {
	now := time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC)
	repo := &fakeSocialHTTPRepo{now: now}
	handler := NewHandler(repo, fakeSocialEvents{now: now})

	tests := []struct {
		name   string
		method string
		path   string
		call   func(http.ResponseWriter, *http.Request)
	}{
		{"add favorite", http.MethodPost, "/api/me/favorites/event-1", handler.FavoriteByID},
		{"remove favorite", http.MethodDelete, "/api/me/favorites/event-1", handler.FavoriteByID},
		{"favorites", http.MethodGet, "/api/me/favorites?limit=5&offset=1", handler.Favorites},
		{"followers", http.MethodGet, "/api/me/followers?limit=200&offset=0", handler.Followers},
		{"following", http.MethodGet, "/api/me/following", handler.Following},
		{"follow", http.MethodPost, "/api/users/user-2/follow", handler.FollowByID},
		{"unfollow", http.MethodDelete, "/api/users/user-2/follow", handler.FollowByID},
		{"collections", http.MethodGet, "/api/me/collections?limit=3", handler.Collections},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, "user-1"))
			rec := httptest.NewRecorder()
			tc.call(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandlerSocialValidation(t *testing.T) {
	handler := NewHandler(&fakeSocialHTTPRepo{}, fakeSocialEvents{})
	req := httptest.NewRequest(http.MethodPost, "/api/me/favorites/event-1", nil)
	rec := httptest.NewRecorder()
	handler.FavoriteByID(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want unauthorized", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/me/favorites?limit=bad", nil)
	req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, "user-1"))
	rec = httptest.NewRecorder()
	handler.Favorites(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want bad request", rec.Code)
	}
}

type fakeSocialHTTPRepo struct{ now time.Time }

func (f *fakeSocialHTTPRepo) AddFavorite(context.Context, string, string) error    { return nil }
func (f *fakeSocialHTTPRepo) RemoveFavorite(context.Context, string, string) error { return nil }
func (f *fakeSocialHTTPRepo) ListFavoriteEvents(context.Context, string, int, int) ([]socialmodel.FavoriteEvent, error) {
	return []socialmodel.FavoriteEvent{{EventID: "event-1", CreatedAt: f.now}}, nil
}
func (f *fakeSocialHTTPRepo) CountFavoriteEvents(context.Context, string) (int, error) { return 1, nil }
func (f *fakeSocialHTTPRepo) FollowUser(context.Context, string, string) error         { return nil }
func (f *fakeSocialHTTPRepo) UnfollowUser(context.Context, string, string) error       { return nil }
func (f *fakeSocialHTTPRepo) ListFollowerProfiles(context.Context, string, string, int, int) ([]socialmodel.UserProfile, int, error) {
	return []socialmodel.UserProfile{testProfile()}, 1, nil
}
func (f *fakeSocialHTTPRepo) ListFollowingProfiles(context.Context, string, string, int, int) ([]socialmodel.UserProfile, int, error) {
	return []socialmodel.UserProfile{testProfile()}, 1, nil
}
func (f *fakeSocialHTTPRepo) UserCollections(context.Context, string, int, int) ([]socialmodel.CollectionCard, int, error) {
	return []socialmodel.CollectionCard{{ID: "collection-1", Title: "Weekend", Description: "Best", ImageURL: "/uploads/collection.png", IsPublic: true}}, 1, nil
}

func testProfile() socialmodel.UserProfile {
	avatar := "/uploads/avatar.png"
	return socialmodel.UserProfile{ID: "user-2", Username: "bob", UserSurname: "smith", AvatarURL: &avatar, City: &socialmodel.City{ID: "city-1", Name: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow"}, IsFollowing: true}
}

type fakeSocialEvents struct{ now time.Time }

func (f fakeSocialEvents) GetByID(context.Context, string, string) (placemodel.EventDetailsView, bool, error) {
	return placemodel.EventDetailsView{
		ID:               "event-1",
		Title:            "Jazz",
		ShortDescription: "Short",
		Images:           []placemodel.EventImageView{{ID: "img-1", ImageURL: "/uploads/event.png"}},
		Tags:             []placemodel.EventTaxonomyItem{{ID: "tag-1", Name: "Jazz", Slug: "jazz"}},
		Sessions: []placemodel.EventSessionView{{
			StartAt: f.now,
			Place:   placemodel.EventSessionPlaceView{Name: "Hall", AddressLine: "Lenina 1"},
		}},
		IsFavorite: true,
	}, true, nil
}
