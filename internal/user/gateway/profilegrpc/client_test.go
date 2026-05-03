package profilegrpc

import (
	"context"
	"testing"
	"time"

	commonv1 "cityhawk/backend/pkg/pb/common/v1"
	profilev1 "cityhawk/backend/pkg/pb/profile/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestClientGetByIDMapsProfile(t *testing.T) {
	createdAt := time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC)
	avatarURL := "/uploads/avatars/user.png"
	birthday := "2000-01-02"
	client := NewClient(fakeProfileClient{
		profile: &profilev1.UserProfile{
			Id:          "user-1",
			Email:       "user@example.com",
			Username:    "user",
			UserSurname: "surname",
			Role:        commonv1.UserRole_USER_ROLE_ADMIN,
			Birthday:    &birthday,
			AvatarUrl:   &avatarURL,
			City: &profilev1.ProfileCity{
				Id:          "city-1",
				Name:        "Moscow",
				CountryName: "Russia",
				Timezone:    "Europe/Moscow",
			},
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: timestamppb.New(createdAt.Add(time.Hour)),
		},
	})

	user, ok := client.GetByID(context.Background(), " user-1 ")
	if !ok {
		t.Fatal("GetByID() ok = false, want true")
	}
	if user.ID != "user-1" || user.Role != "admin" || user.City == nil || user.City.ID != "city-1" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if user.Birthday == nil || user.Birthday.Format("2006-01-02") != birthday {
		t.Fatalf("unexpected birthday: %+v", user.Birthday)
	}
	if user.AvatarURL == nil || *user.AvatarURL != avatarURL {
		t.Fatalf("unexpected avatar: %+v", user.AvatarURL)
	}
}

func TestClientGetByIDHandlesNotFound(t *testing.T) {
	client := NewClient(fakeProfileClient{err: status.Error(codes.NotFound, "user not found")})

	_, ok := client.GetByID(context.Background(), "missing-user")
	if ok {
		t.Fatal("GetByID() ok = true, want false")
	}
}

type fakeProfileClient struct {
	profile *profilev1.UserProfile
	err     error
}

func (f fakeProfileClient) GetMe(context.Context, *profilev1.GetMeRequest, ...grpc.CallOption) (*profilev1.UserProfile, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (f fakeProfileClient) UpdateMe(context.Context, *profilev1.UpdateMeRequest, ...grpc.CallOption) (*profilev1.UserProfile, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (f fakeProfileClient) GetUser(context.Context, *profilev1.GetUserRequest, ...grpc.CallOption) (*profilev1.UserProfile, error) {
	return f.profile, f.err
}
