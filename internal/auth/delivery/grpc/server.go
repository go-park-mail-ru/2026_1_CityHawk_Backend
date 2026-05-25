package grpc

import (
	"context"
	"errors"
	"strings"

	authmodel "cityhawk/backend/internal/auth/model"
	authvalidation "cityhawk/backend/internal/auth/validation"
	"cityhawk/backend/internal/grpcconv"
	platformerrors "cityhawk/backend/internal/platform/errors"
	platformsecurity "cityhawk/backend/internal/platform/security"
	usermodel "cityhawk/backend/internal/user/model"
	authv1 "cityhawk/backend/pkg/pb/auth/v1"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
)

type UserReader interface {
	GetByID(ctx context.Context, id string) (usermodel.User, bool)
}

type AuthFlow interface {
	Register(ctx context.Context, input authmodel.RegisterInput) (authmodel.RegistrationResult, error)
	Login(ctx context.Context, input authmodel.LoginInput) (authmodel.SessionResult, error)
}

type RefreshTokenManager interface {
	RotateRefresh(ctx context.Context, oldRefreshToken string) (authmodel.TokenPair, error)
	RevokeRefresh(ctx context.Context, token string) error
}

type OAuthLogin interface {
	LoginWithYandex(ctx context.Context, code string) (authmodel.TokenPair, error)
	LoginWithVK(ctx context.Context, code string) (authmodel.TokenPair, error)
}

type TokenParser interface {
	Parse(token string) (platformsecurity.Claims, error)
}

type Server struct {
	authv1.UnimplementedAuthServiceServer

	flow   AuthFlow
	tokens RefreshTokenManager
	oauth  OAuthLogin
	users  UserReader
	parser TokenParser
}

func NewServer(
	flow AuthFlow,
	tokens RefreshTokenManager,
	oauth OAuthLogin,
	users UserReader,
	parser TokenParser,
) *Server {
	return &Server{
		flow:   flow,
		tokens: tokens,
		oauth:  oauth,
		users:  users,
		parser: parser,
	}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.SessionResponse, error) {
	result, err := s.flow.Register(ctx, authmodel.RegisterInput{
		Email:       req.GetEmail(),
		Username:    req.GetUsername(),
		UserSurname: req.GetUserSurname(),
		Password:    req.GetPassword(),
		Birthday:    req.GetBirthday(),
		CityID:      req.GetCityId(),
	})
	if err != nil {
		return nil, authError(err)
	}

	return sessionResponse(result.User, result.Tokens), nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.SessionResponse, error) {
	result, err := s.flow.Login(ctx, authmodel.LoginInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, authError(err)
	}

	return sessionResponse(result.User, result.Tokens), nil
}

func (s *Server) OAuthLogin(ctx context.Context, req *authv1.OAuthLoginRequest) (*authv1.SessionResponse, error) {
	var (
		tokens authmodel.TokenPair
		err    error
	)

	switch req.GetProvider() {
	case authv1.OAuthProvider_OAUTH_PROVIDER_YANDEX:
		tokens, err = s.oauth.LoginWithYandex(ctx, req.GetCode())
	case authv1.OAuthProvider_OAUTH_PROVIDER_VK:
		tokens, err = s.oauth.LoginWithVK(ctx, req.GetCode())
	default:
		return nil, grpcconv.InvalidArgument("oauth provider is required")
	}
	if err != nil {
		return nil, authError(err)
	}

	user, err := s.userFromAccessToken(ctx, tokens.AccessToken)
	if err != nil {
		return nil, err
	}
	return sessionResponse(user, tokens), nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.SessionResponse, error) {
	if strings.TrimSpace(req.GetRefreshToken()) == "" {
		return nil, grpcconv.Error(platformerrors.ErrMissingRefresh)
	}

	tokens, err := s.tokens.RotateRefresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, authError(err)
	}

	user, err := s.userFromAccessToken(ctx, tokens.AccessToken)
	if err != nil {
		return nil, err
	}
	return sessionResponse(user, tokens), nil
}

func (s *Server) Logout(ctx context.Context, req *authv1.LogoutRequest) (*commonv1.BoolResponse, error) {
	if strings.TrimSpace(req.GetRefreshToken()) == "" {
		return nil, grpcconv.Error(platformerrors.ErrMissingRefresh)
	}
	if err := s.tokens.RevokeRefresh(ctx, req.GetRefreshToken()); err != nil {
		return nil, authError(err)
	}
	return &commonv1.BoolResponse{Ok: true}, nil
}

func (s *Server) ValidateAccessToken(ctx context.Context, req *authv1.ValidateAccessTokenRequest) (*authv1.ValidateAccessTokenResponse, error) {
	claims, err := s.parser.Parse(req.GetAccessToken())
	if err != nil || claims.Type != "access" || claims.Subject == "" {
		return &authv1.ValidateAccessTokenResponse{Valid: false}, nil
	}

	user, ok := s.users.GetByID(ctx, claims.Subject)
	if !ok {
		return &authv1.ValidateAccessTokenResponse{Valid: false}, nil
	}

	return &authv1.ValidateAccessTokenResponse{
		Valid: true,
		Context: &commonv1.UserContext{
			UserId:        user.ID,
			Role:          grpcconv.UserRoleToProto(user.Role),
			Authenticated: true,
		},
	}, nil
}

func (s *Server) userFromAccessToken(ctx context.Context, accessToken string) (usermodel.User, error) {
	claims, err := s.parser.Parse(accessToken)
	if err != nil || claims.Type != "access" || claims.Subject == "" {
		return usermodel.User{}, grpcconv.Error(platformerrors.ErrInvalidAccess)
	}
	user, ok := s.users.GetByID(ctx, claims.Subject)
	if !ok {
		return usermodel.User{}, grpcconv.Error(platformerrors.ErrUserNotFound)
	}
	return user, nil
}

func sessionResponse(user usermodel.User, tokens authmodel.TokenPair) *authv1.SessionResponse {
	csrfToken, _ := platformsecurity.RandomHex(32)
	return &authv1.SessionResponse{
		UserId:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     grpcconv.UserRoleToProto(user.Role),
		Tokens: &authv1.TokenPair{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			CsrfToken:    csrfToken,
			ExpiresIn:    tokens.ExpiresIn,
		},
	}
}

func authError(err error) error {
	var validationErr authvalidation.ValidationError
	if errors.As(err, &validationErr) {
		return grpcconv.InvalidArgument("validation failed")
	}
	return grpcconv.Error(err)
}
