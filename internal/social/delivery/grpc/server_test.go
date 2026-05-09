package grpc

import (
	"context"
	"testing"
	"time"

	socialmodel "cityhawk/backend/internal/social/model"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
	socialv1 "cityhawk/backend/pkg/pb/social/v1"
)

func TestServerSocialFlow(t *testing.T) {
	repo := &fakeSocialRepo{now: time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC)}
	server := NewServer(repo)
	actor := &commonv1.UserContext{UserId: "user-1", Authenticated: true}

	if resp, err := server.AddFavorite(context.Background(), &socialv1.AddFavoriteRequest{Actor: actor, EventId: "event-1"}); err != nil || !resp.GetOk() {
		t.Fatalf("AddFavorite() = (%+v, %v)", resp, err)
	}
	if resp, err := server.RemoveFavorite(context.Background(), &socialv1.RemoveFavoriteRequest{Actor: actor, EventId: "event-1"}); err != nil || !resp.GetOk() {
		t.Fatalf("RemoveFavorite() = (%+v, %v)", resp, err)
	}
	if resp, err := server.FollowUser(context.Background(), &socialv1.FollowUserRequest{Actor: actor, FollowedUserId: "user-2"}); err != nil || !resp.GetOk() {
		t.Fatalf("FollowUser() = (%+v, %v)", resp, err)
	}
	if resp, err := server.UnfollowUser(context.Background(), &socialv1.UnfollowUserRequest{Actor: actor, FollowedUserId: "user-2"}); err != nil || !resp.GetOk() {
		t.Fatalf("UnfollowUser() = (%+v, %v)", resp, err)
	}

	favorites, err := server.ListFavoriteEvents(context.Background(), &socialv1.ListFavoriteEventsRequest{
		Actor: actor,
		Page:  &commonv1.PageRequest{Limit: 5, Offset: -1},
	})
	if err != nil || len(favorites.GetItems()) != 1 || favorites.GetPage().GetOffset() != 0 {
		t.Fatalf("ListFavoriteEvents() = (%+v, %v)", favorites, err)
	}

	favorite, err := server.IsFavorite(context.Background(), &socialv1.IsFavoriteRequest{UserId: "user-1", EventId: "event-1"})
	if err != nil || !favorite.GetIsFavorite() {
		t.Fatalf("IsFavorite() = (%+v, %v)", favorite, err)
	}

	followers, err := server.ListFollowers(context.Background(), &socialv1.ListFollowersRequest{
		UserId:        "user-2",
		ViewerContext: actor,
		Page:          &commonv1.PageRequest{Limit: 1},
	})
	if err != nil || len(followers.GetItems()) != 1 || !followers.GetItems()[0].GetIsFollowing() {
		t.Fatalf("ListFollowers() = (%+v, %v)", followers, err)
	}

	following, err := server.ListFollowing(context.Background(), &socialv1.ListFollowingRequest{UserId: "user-1", ViewerContext: actor})
	if err != nil || len(following.GetItems()) != 1 {
		t.Fatalf("ListFollowing() = (%+v, %v)", following, err)
	}

	flag, err := server.IsFollowing(context.Background(), &socialv1.IsFollowingRequest{FollowerUserId: "user-1", FollowedUserId: "user-2"})
	if err != nil || !flag.GetIsFollowing() {
		t.Fatalf("IsFollowing() = (%+v, %v)", flag, err)
	}

	flags, err := server.GetSocialFlags(context.Background(), &socialv1.GetSocialFlagsRequest{
		ViewerContext: actor,
		EventIds:      []string{"event-1"},
		AuthorUserIds: []string{"user-2"},
	})
	if err != nil || len(flags.GetFavoriteEvents()) != 1 || len(flags.GetFollowingUsers()) != 1 {
		t.Fatalf("GetSocialFlags() = (%+v, %v)", flags, err)
	}
}

func TestServerRequiresActor(t *testing.T) {
	server := NewServer(&fakeSocialRepo{})
	if _, err := server.AddFavorite(context.Background(), &socialv1.AddFavoriteRequest{EventId: "event-1"}); err == nil {
		t.Fatal("AddFavorite() accepted missing actor")
	}
}

type fakeSocialRepo struct {
	now time.Time
}

func (f *fakeSocialRepo) AddFavorite(context.Context, string, string) error    { return nil }
func (f *fakeSocialRepo) RemoveFavorite(context.Context, string, string) error { return nil }

func (f *fakeSocialRepo) ListFavoriteEvents(context.Context, string, int, int) ([]socialmodel.FavoriteEvent, error) {
	return []socialmodel.FavoriteEvent{{EventID: "event-1", CreatedAt: f.now}}, nil
}

func (f *fakeSocialRepo) IsFavorite(context.Context, string, string) (socialmodel.FavoriteFlag, error) {
	createdAt := f.now
	return socialmodel.FavoriteFlag{EventID: "event-1", IsFavorite: true, CreatedAt: &createdAt}, nil
}

func (f *fakeSocialRepo) FollowUser(context.Context, string, string) error   { return nil }
func (f *fakeSocialRepo) UnfollowUser(context.Context, string, string) error { return nil }

func (f *fakeSocialRepo) ListFollowers(context.Context, string, int, int) ([]socialmodel.UserFollow, error) {
	return []socialmodel.UserFollow{{UserID: "user-2", CreatedAt: f.now}}, nil
}

func (f *fakeSocialRepo) ListFollowing(context.Context, string, int, int) ([]socialmodel.UserFollow, error) {
	return []socialmodel.UserFollow{{UserID: "user-2", CreatedAt: f.now}}, nil
}

func (f *fakeSocialRepo) IsFollowing(context.Context, string, string) (socialmodel.FollowingFlag, error) {
	createdAt := f.now
	return socialmodel.FollowingFlag{UserID: "user-2", IsFollowing: true, CreatedAt: &createdAt}, nil
}

func (f *fakeSocialRepo) FavoriteFlags(_ context.Context, _ string, eventIDs []string) (map[string]socialmodel.FavoriteFlag, error) {
	flags := make(map[string]socialmodel.FavoriteFlag, len(eventIDs))
	for _, id := range eventIDs {
		createdAt := f.now
		flags[id] = socialmodel.FavoriteFlag{EventID: id, IsFavorite: true, CreatedAt: &createdAt}
	}
	return flags, nil
}

func (f *fakeSocialRepo) FollowingFlags(_ context.Context, _ string, userIDs []string) (map[string]socialmodel.FollowingFlag, error) {
	flags := make(map[string]socialmodel.FollowingFlag, len(userIDs))
	for _, id := range userIDs {
		createdAt := f.now
		flags[id] = socialmodel.FollowingFlag{UserID: id, IsFollowing: true, CreatedAt: &createdAt}
	}
	return flags, nil
}
