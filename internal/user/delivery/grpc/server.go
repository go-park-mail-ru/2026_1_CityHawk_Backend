package grpc

import (
	"context"
	"errors"

	authvalidation "cityhawk/backend/internal/auth/validation"
	"cityhawk/backend/internal/grpcconv"
	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
	profilev1 "cityhawk/backend/pkg/pb/profile/v1"
)

type UserStore interface {
	GetByID(ctx context.Context, id string) (usermodel.User, bool)
	UpdateProfile(ctx context.Context, id string, patch usermodel.ProfilePatch) (usermodel.User, bool, error)
}

type Server struct {
	profilev1.UnimplementedProfileServiceServer

	users UserStore
}

func NewServer(users UserStore) *Server {
	return &Server{users: users}
}

func (s *Server) GetMe(ctx context.Context, req *profilev1.GetMeRequest) (*profilev1.UserProfile, error) {
	userID := grpcconv.UserIDFromContext(req.GetContext())
	if userID == "" {
		return nil, grpcconv.Error(platformerrors.ErrUnauthorized)
	}

	user, ok := s.users.GetByID(ctx, userID)
	if !ok {
		return nil, grpcconv.Error(platformerrors.ErrUserNotFound)
	}
	return userProfile(user), nil
}

func (s *Server) UpdateMe(ctx context.Context, req *profilev1.UpdateMeRequest) (*profilev1.UserProfile, error) {
	userID := grpcconv.UserIDFromContext(req.GetContext())
	if userID == "" {
		return nil, grpcconv.Error(platformerrors.ErrUnauthorized)
	}

	email, username, userSurname, birthday, cityID, avatarURL, err := authvalidation.ValidateProfilePatch(
		req.Email,
		req.Username,
		req.UserSurname,
		req.Birthday,
		req.CityId,
		req.AvatarUrl,
	)
	if err != nil {
		var validationErr authvalidation.ValidationError
		if errors.As(err, &validationErr) {
			return nil, grpcconv.InvalidArgument("validation failed")
		}
		return nil, grpcconv.Error(err)
	}

	patch := usermodel.ProfilePatch{}
	if req.Email != nil {
		patch.Email = &email
	}
	if req.Username != nil {
		patch.Username = &username
	}
	if req.UserSurname != nil {
		patch.UserSurname = &userSurname
	}
	if req.Birthday != nil {
		patch.Birthday = birthday
	}
	if req.CityId != nil {
		patch.CityID = &cityID
	}
	if req.AvatarUrl != nil {
		patch.AvatarURL = &avatarURL
	}

	user, ok, err := s.users.UpdateProfile(ctx, userID, patch)
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	if !ok {
		return nil, grpcconv.Error(platformerrors.ErrUserNotFound)
	}
	return userProfile(user), nil
}

func (s *Server) GetUser(ctx context.Context, req *profilev1.GetUserRequest) (*profilev1.UserProfile, error) {
	user, ok := s.users.GetByID(ctx, req.GetUserId())
	if !ok {
		return nil, grpcconv.Error(platformerrors.ErrUserNotFound)
	}
	return userProfile(user), nil
}

func userProfile(user usermodel.User) *profilev1.UserProfile {
	var birthday *string
	if user.Birthday != nil {
		value := user.Birthday.UTC().Format("2006-01-02")
		birthday = &value
	}

	var city *profilev1.ProfileCity
	if user.City != nil {
		city = &profilev1.ProfileCity{
			Id:          user.City.ID,
			Name:        user.City.Name,
			CountryName: user.City.CountryName,
			Timezone:    user.City.Timezone,
		}
	}

	return &profilev1.UserProfile{
		Id:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		UserSurname: user.UserSurname,
		Role:        grpcconv.UserRoleToProto(user.Role),
		Birthday:    birthday,
		AvatarUrl:   user.AvatarURL,
		City:        city,
		CreatedAt:   grpcconv.TimeToProto(user.CreatedAt),
		UpdatedAt:   grpcconv.TimeToProto(user.UpdatedAt),
	}
}
