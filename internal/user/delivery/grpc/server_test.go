package grpc

import (
	"context"
	"testing"
	"time"

	usermodel "cityhawk/backend/internal/user/model"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
	profilev1 "cityhawk/backend/pkg/pb/profile/v1"
)

func TestServerProfileFlow(t *testing.T) {
	store := &fakeUserStore{user: testUser()}
	server := NewServer(store)
	ctx := &commonv1.UserContext{UserId: "user-1", Authenticated: true}

	me, err := server.GetMe(context.Background(), &profilev1.GetMeRequest{Context: ctx})
	if err != nil || me.GetId() != "user-1" || me.GetCity().GetId() != "city-1" {
		t.Fatalf("GetMe() = (%+v, %v)", me, err)
	}

	email := "new@example.com"
	username := "new_user"
	surname := "New"
	birthday := "2001-02-03"
	cityID := "11111111-1111-1111-1111-111111111111"
	avatarURL := "/uploads/avatar.png"
	updated, err := server.UpdateMe(context.Background(), &profilev1.UpdateMeRequest{
		Context:     ctx,
		Email:       &email,
		Username:    &username,
		UserSurname: &surname,
		Birthday:    &birthday,
		CityId:      &cityID,
		AvatarUrl:   &avatarURL,
	})
	if err != nil || updated.GetEmail() != email || store.patch.Email == nil {
		t.Fatalf("UpdateMe() = (%+v, %v), patch=%+v", updated, err, store.patch)
	}

	got, err := server.GetUser(context.Background(), &profilev1.GetUserRequest{UserId: "user-1"})
	if err != nil || got.GetRole() != commonv1.UserRole_USER_ROLE_ADMIN {
		t.Fatalf("GetUser() = (%+v, %v)", got, err)
	}
}

func TestServerProfileErrors(t *testing.T) {
	server := NewServer(&fakeUserStore{})
	if _, err := server.GetMe(context.Background(), &profilev1.GetMeRequest{}); err == nil {
		t.Fatal("GetMe() accepted missing context")
	}
	if _, err := server.GetUser(context.Background(), &profilev1.GetUserRequest{UserId: "missing"}); err == nil {
		t.Fatal("GetUser() accepted missing user")
	}
	badEmail := "bad"
	if _, err := server.UpdateMe(context.Background(), &profilev1.UpdateMeRequest{
		Context: &commonv1.UserContext{UserId: "user-1", Authenticated: true},
		Email:   &badEmail,
	}); err == nil {
		t.Fatal("UpdateMe() accepted invalid email")
	}
}

type fakeUserStore struct {
	user  usermodel.User
	patch usermodel.ProfilePatch
}

func (f *fakeUserStore) GetByID(_ context.Context, id string) (usermodel.User, bool) {
	if f.user.ID == id {
		return f.user, true
	}
	return usermodel.User{}, false
}

func (f *fakeUserStore) UpdateProfile(_ context.Context, id string, patch usermodel.ProfilePatch) (usermodel.User, bool, error) {
	f.patch = patch
	if f.user.ID != id {
		return usermodel.User{}, false, nil
	}
	if patch.Email != nil {
		f.user.Email = *patch.Email
	}
	return f.user, true, nil
}

func testUser() usermodel.User {
	birthday := time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)
	avatar := "/uploads/avatar.png"
	return usermodel.User{
		ID: "user-1", Email: "user@example.com", Username: "user", UserSurname: "surname",
		Role: usermodel.RoleAdmin, Birthday: &birthday, AvatarURL: &avatar,
		City:      &usermodel.City{ID: "city-1", Name: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow"},
		CreatedAt: time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.May, 4, 11, 0, 0, 0, time.UTC),
	}
}
