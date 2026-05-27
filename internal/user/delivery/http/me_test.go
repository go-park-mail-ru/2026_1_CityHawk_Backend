package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cityhawk/backend/internal/platform/httpx"
	usermodel "cityhawk/backend/internal/user/model"
)

func TestMeHandlerFlow(t *testing.T) {
	users := &fakeMeUsers{user: meTestUser()}
	handler := NewMeHandler(users, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, "user-1"))
	rec := httptest.NewRecorder()
	handler.Me(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status=%d body=%s", rec.Code, rec.Body.String())
	}

	body := patchMeRequest{
		Email:          stringPtr("new@example.com"),
		Username:       stringPtr("new_user"),
		Birthday:       stringPtr("2001-02-03"),
		CityID:         stringPtr("11111111-1111-1111-1111-111111111111"),
		Bio:            stringPtr("Bio"),
		InterestTagIDs: []string{"11111111-1111-1111-1111-111111111111"},
		AvatarURL:      stringPtr("/uploads/avatar.png"),
	}
	raw, _ := json.Marshal(body)
	req = httptest.NewRequest(http.MethodPatch, "/api/me", bytes.NewReader(raw))
	req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, "user-1"))
	rec = httptest.NewRecorder()
	handler.Me(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("PATCH status=%d body=%s", rec.Code, rec.Body.String())
	}
	if users.patch.Email == nil || *users.patch.Email != "new@example.com" || users.patch.InterestTagIDs == nil {
		t.Fatalf("patch was not passed: %+v", users.patch)
	}
}

func TestMeHandlerErrors(t *testing.T) {
	handler := NewMeHandler(&fakeMeUsers{}, nil)
	rec := httptest.NewRecorder()
	handler.Me(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want unauthorized", rec.Code)
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/me", bytes.NewBufferString(`{"email":"bad"}`))
	req = req.WithContext(context.WithValue(req.Context(), httpx.UserIDContextKey, "user-1"))
	rec = httptest.NewRecorder()
	handler.Me(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want bad request", rec.Code)
	}
}

type fakeMeUsers struct {
	user  usermodel.User
	patch usermodel.ProfilePatch
}

func (f *fakeMeUsers) GetByID(_ context.Context, id string) (usermodel.User, bool) {
	if f.user.ID == id {
		return f.user, true
	}
	return usermodel.User{}, false
}

func (f *fakeMeUsers) UpdateProfile(_ context.Context, id string, patch usermodel.ProfilePatch) (usermodel.User, bool, error) {
	f.patch = patch
	if f.user.ID != id {
		return usermodel.User{}, false, nil
	}
	if patch.Email != nil {
		f.user.Email = *patch.Email
	}
	return f.user, true, nil
}

func meTestUser() usermodel.User {
	birthday := time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)
	avatar := "/uploads/avatar.png"
	bio := "Bio"
	return usermodel.User{
		ID: "user-1", Email: "user@example.com", Username: "user",
		Role: usermodel.RoleUser, Birthday: &birthday, AvatarURL: &avatar, Bio: &bio,
		InterestTagIDs: []string{"tag-1"},
		City:           &usermodel.City{ID: "city-1", Name: "Moscow", CountryName: "Russia", Timezone: "Europe/Moscow"},
		CreatedAt:      time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, time.May, 4, 11, 0, 0, 0, time.UTC),
	}
}

func stringPtr(value string) *string { return &value }
