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
		{"users search", http.MethodGet, "/api/users/search?query=bo&limit=5", handler.UsersSearch},
		{"follow", http.MethodPost, "/api/users/user-2/follow", handler.FollowByID},
		{"unfollow", http.MethodDelete, "/api/users/user-2/follow", handler.FollowByID},
		{"collections", http.MethodGet, "/api/me/collections?limit=3", handler.Collections},
		{"invitees", http.MethodGet, "/api/events/event-1/invitees", handler.Invitees},
		{"notification events", http.MethodGet, "/api/me/notifications/events?limit=4", handler.NotificationEvents},
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
func (f *fakeSocialHTTPRepo) SearchUsers(context.Context, string, string, int, int) ([]socialmodel.UserProfile, int, error) {
	return []socialmodel.UserProfile{testProfile()}, 1, nil
}
func (f *fakeSocialHTTPRepo) UserCollections(context.Context, string, int, int) ([]socialmodel.CollectionCard, int, error) {
	return []socialmodel.CollectionCard{{ID: "collection-1", Title: "Weekend", Description: "Best", ImageURL: "/uploads/collection.png", IsPublic: true}}, 1, nil
}
func (f *fakeSocialHTTPRepo) SearchInvitees(context.Context, string, string, string, int) ([]socialmodel.InviteeCandidate, error) {
	return []socialmodel.InviteeCandidate{{ID: "user-2", Username: "bob"}}, nil
}
func (f *fakeSocialHTTPRepo) ListInvitees(context.Context, string) ([]socialmodel.InviteeCandidate, error) {
	status := "pending"
	return []socialmodel.InviteeCandidate{{ID: "user-2", Username: "bob", InvitationStatus: &status}}, nil
}
func (f *fakeSocialHTTPRepo) CreateInvitations(context.Context, string, string, []string, string, *string) ([]socialmodel.Invitation, error) {
	return []socialmodel.Invitation{{ID: "invitation-1", EventID: "event-1", SenderID: "user-1", RecipientID: "user-2", Status: "pending", CreatedAt: f.now, UpdatedAt: f.now}}, nil
}
func (f *fakeSocialHTTPRepo) UpdateInvitationStatus(context.Context, string, string, string) (socialmodel.Invitation, error) {
	return socialmodel.Invitation{ID: "invitation-1", EventID: "event-1", SenderID: "user-2", RecipientID: "user-1", Status: "accepted", CreatedAt: f.now, UpdatedAt: f.now}, nil
}
func (f *fakeSocialHTTPRepo) ListNotifications(context.Context, string, string, bool, int, int) ([]socialmodel.Notification, int, int, error) {
	return []socialmodel.Notification{{ID: "notification-1", Type: "system", CreatedAt: f.now}}, 1, 1, nil
}
func (f *fakeSocialHTTPRepo) ListNotificationEvents(context.Context, string, int, int) ([]socialmodel.NotificationEventRef, int, error) {
	avatar := "/uploads/avatar.png"
	return []socialmodel.NotificationEventRef{{
		EventID:   "event-1",
		CreatedAt: f.now,
		InvitedBy: &socialmodel.NotificationEventInviter{
			ID:        "user-2",
			Username:  "alice",
			AvatarURL: &avatar,
		},
		Invitation: &socialmodel.NotificationInvitation{ID: "invitation-1", Status: "accepted"},
	}}, 1, nil
}
func (f *fakeSocialHTTPRepo) MarkNotificationRead(context.Context, string, string) (int, error) {
	return 0, nil
}
func (f *fakeSocialHTTPRepo) MarkAllNotificationsRead(context.Context, string) (int, error) {
	return 0, nil
}
func (f *fakeSocialHTTPRepo) CreateEventShareLink(context.Context, string, string, string) (socialmodel.ShareLink, error) {
	eventID := "event-1"
	return socialmodel.ShareLink{ID: "share-1", Token: "abcdef", EventID: &eventID, CreatedAt: f.now}, nil
}
func (f *fakeSocialHTTPRepo) CreateCollectionShareLink(context.Context, string, string, string) (socialmodel.ShareLink, error) {
	collectionID := "collection-1"
	return socialmodel.ShareLink{ID: "share-1", Token: "abcdef", CollectionID: &collectionID, CreatedAt: f.now}, nil
}
func (f *fakeSocialHTTPRepo) ResolveShareLink(context.Context, string) (socialmodel.ShareLink, bool, error) {
	eventID := "event-1"
	return socialmodel.ShareLink{ID: "share-1", Token: "abcdef", EventID: &eventID, CreatedAt: f.now}, true, nil
}

func testProfile() socialmodel.UserProfile {
	avatar := "/uploads/avatar.png"
	return socialmodel.UserProfile{ID: "user-2", Username: "bob", AvatarURL: &avatar, City: &socialmodel.City{ID: "city-1", Name: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow"}, IsFollowing: true}
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
