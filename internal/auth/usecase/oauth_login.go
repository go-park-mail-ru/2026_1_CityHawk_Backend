package usecase

import (
	"context"

	authmodel "cityhawk/backend/internal/auth/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
)

type OAuthGateway interface {
	FetchIdentity(ctx context.Context, code string) (authmodel.OAuthIdentity, error)
}

type OAuthUserFinder interface {
	FindOrCreateFromOAuth(identity authmodel.OAuthIdentity) (usermodel.User, error)
}

type OAuthTokenIssuer interface {
	IssueTokenPair(u usermodel.User) (authmodel.TokenPair, error)
}

type OAuthLoginService struct {
	google OAuthGateway
	yandex OAuthGateway
	vk     OAuthGateway
	users  OAuthUserFinder
	issuer OAuthTokenIssuer
}

func NewOAuthLoginService(
	google OAuthGateway,
	yandex OAuthGateway,
	vk OAuthGateway,
	users OAuthUserFinder,
	issuer OAuthTokenIssuer,
) *OAuthLoginService {
	return &OAuthLoginService{
		google: google,
		yandex: yandex,
		vk:     vk,
		users:  users,
		issuer: issuer,
	}
}

func (s *OAuthLoginService) LoginWithGoogle(ctx context.Context, code string) (authmodel.TokenPair, error) {
	return s.loginWithIdentity(ctx, code, s.google)
}

func (s *OAuthLoginService) LoginWithYandex(ctx context.Context, code string) (authmodel.TokenPair, error) {
	return s.loginWithIdentity(ctx, code, s.yandex)
}

func (s *OAuthLoginService) LoginWithVK(ctx context.Context, code string) (authmodel.TokenPair, error) {
	return s.loginWithIdentity(ctx, code, s.vk)
}

type oauthIdentityFetcher interface {
	FetchIdentity(ctx context.Context, code string) (authmodel.OAuthIdentity, error)
}

func (s *OAuthLoginService) loginWithIdentity(ctx context.Context, code string, gateway oauthIdentityFetcher) (authmodel.TokenPair, error) {
	identity, err := gateway.FetchIdentity(ctx, code)
	if err != nil {
		return authmodel.TokenPair{}, err
	}

	u, err := s.users.FindOrCreateFromOAuth(identity)
	if err != nil {
		return authmodel.TokenPair{}, platformerrors.ErrInternal
	}

	resp, err := s.issuer.IssueTokenPair(u)
	if err != nil {
		return authmodel.TokenPair{}, platformerrors.ErrIssueTokens
	}

	return resp, nil
}
