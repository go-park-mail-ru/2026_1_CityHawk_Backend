package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
)

func TestInMemoryUserRepositoryFlow(t *testing.T) {
	repo := NewInMemoryUserRepository()
	cityID := "11111111-1111-1111-1111-111111111111"
	user, err := repo.Create(context.Background(), usermodel.User{
		ID: "user-1", Email: "user@example.com", Username: "user", CityID: &cityID,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if user.Role != usermodel.RoleUser || user.City == nil {
		t.Fatalf("unexpected created user: %+v", user)
	}
	if _, err := repo.Create(context.Background(), usermodel.User{ID: "user-2", Email: "user@example.com"}); !errors.Is(err, platformerrors.ErrEmailExists) {
		t.Fatalf("duplicate Create() err = %v", err)
	}
	if got, ok := repo.GetByEmail(context.Background(), "user@example.com"); !ok || got.ID != "user-1" {
		t.Fatalf("GetByEmail() = (%+v, %v)", got, ok)
	}
	if got, ok := repo.GetByID(context.Background(), "user-1"); !ok || got.Email != "user@example.com" {
		t.Fatalf("GetByID() = (%+v, %v)", got, ok)
	}

	newEmail := "new@example.com"
	username := "new"
	birthday := time.Date(2001, time.February, 3, 0, 0, 0, 0, time.UTC)
	avatar := "/uploads/avatar.png"
	bio := "Bio"
	tags := []string{"tag-1", "tag-2"}
	updated, ok, err := repo.UpdateProfile(context.Background(), "user-1", usermodel.ProfilePatch{
		Email: &newEmail, Username: &username, Birthday: &birthday,
		CityID: &cityID, AvatarURL: &avatar, Bio: &bio, InterestTagIDs: &tags,
	})
	if err != nil || !ok {
		t.Fatalf("UpdateProfile() = (%+v, %v, %v)", updated, ok, err)
	}
	if updated.Email != newEmail || updated.Birthday == nil || len(updated.InterestTagIDs) != 2 {
		t.Fatalf("unexpected updated user: %+v", updated)
	}
	if _, ok, err := repo.UpdateProfile(context.Background(), "missing", usermodel.ProfilePatch{}); err != nil || ok {
		t.Fatalf("missing UpdateProfile() ok=%v err=%v", ok, err)
	}
}
