package profilegrpc

import (
	"context"
	"strings"
	"time"

	"cityhawk/backend/internal/grpcconv"
	usermodel "cityhawk/backend/internal/user/model"
	profilev1 "cityhawk/backend/pkg/pb/profile/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	profiles profilev1.ProfileServiceClient
}

func NewClient(profiles profilev1.ProfileServiceClient) *Client {
	return &Client{profiles: profiles}
}

func (c *Client) GetByID(ctx context.Context, id string) (usermodel.User, bool) {
	id = strings.TrimSpace(id)
	if c == nil || c.profiles == nil || id == "" {
		return usermodel.User{}, false
	}

	profile, err := c.profiles.GetUser(ctx, &profilev1.GetUserRequest{UserId: id})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return usermodel.User{}, false
		}
		return usermodel.User{}, false
	}
	if profile == nil {
		return usermodel.User{}, false
	}

	user := usermodel.User{
		ID:          profile.GetId(),
		Email:       profile.GetEmail(),
		Username:    profile.GetUsername(),
		UserSurname: profile.GetUserSurname(),
		Role:        grpcconv.UserRoleFromProto(profile.GetRole()),
	}
	if profile.GetCreatedAt() != nil {
		user.CreatedAt = profile.GetCreatedAt().AsTime()
	}
	if profile.GetUpdatedAt() != nil {
		user.UpdatedAt = profile.GetUpdatedAt().AsTime()
	}
	if user.Role == "" {
		user.Role = usermodel.RoleUser
	}
	if profile.Birthday != nil {
		value := profile.GetBirthday()
		user.Birthday = parseProfileBirthday(value)
	}
	if profile.AvatarUrl != nil {
		value := profile.GetAvatarUrl()
		user.AvatarURL = &value
	}
	if profile.City != nil {
		cityID := profile.City.GetId()
		user.CityID = &cityID
		user.City = &usermodel.City{
			ID:          profile.City.GetId(),
			Name:        profile.City.GetName(),
			CountryName: profile.City.GetCountryName(),
			Timezone:    profile.City.GetTimezone(),
		}
	}

	return user, user.ID != ""
}

func parseProfileBirthday(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	parsed = parsed.UTC()
	return &parsed
}
