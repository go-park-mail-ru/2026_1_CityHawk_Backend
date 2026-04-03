package usecase

import (
	"context"
	"time"

	authmodel "cityhawk/backend/internal/auth/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
)

type UserReader interface {
	GetByID(ctx context.Context, id string) (usermodel.User, bool)
}

type RefreshSessionRepository interface {
	Store(ctx context.Context, refreshToken, userID string, expiresAt time.Time) error
	Consume(ctx context.Context, refreshToken string) (string, error)
	Revoke(ctx context.Context, refreshToken string) error
}

type TokenService interface {
	SignAccessToken(u usermodel.User, ttl time.Duration) (string, error)
	GenerateOpaqueToken(size int) (string, error)
}

type Service struct {
	accessTTL  time.Duration
	refreshTTL time.Duration

	users   UserReader
	refresh RefreshSessionRepository
	tokens  TokenService
}

func NewService(
	accessTTL time.Duration,
	refreshTTL time.Duration,
	users UserReader,
	refresh RefreshSessionRepository,
	tokens TokenService,
) *Service {
	return &Service{
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		users:      users,
		refresh:    refresh,
		tokens:     tokens,
	}
}

func (s *Service) IssueTokenPair(ctx context.Context, u usermodel.User) (authmodel.TokenPair, error) {
	accessToken, err := s.tokens.SignAccessToken(u, s.accessTTL)
	if err != nil {
		return authmodel.TokenPair{}, err
	}

	refreshToken, err := s.tokens.GenerateOpaqueToken(32)
	if err != nil {
		return authmodel.TokenPair{}, err
	}

	if err := s.refresh.Store(ctx, refreshToken, u.ID, time.Now().UTC().Add(s.refreshTTL)); err != nil {
		return authmodel.TokenPair{}, platformerrors.ErrInternal
	}

	return authmodel.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *Service) IssueTokenPairForUserID(ctx context.Context, userID string) (authmodel.TokenPair, error) {
	u, ok := s.users.GetByID(ctx, userID)
	if !ok {
		return authmodel.TokenPair{}, platformerrors.ErrUserNotFound
	}

	return s.IssueTokenPair(ctx, u)
}

func (s *Service) RotateRefresh(ctx context.Context, oldRefreshToken string) (authmodel.TokenPair, error) {
	userID, err := s.refresh.Consume(ctx, oldRefreshToken)
	if err != nil {
		return authmodel.TokenPair{}, platformerrors.ErrTokenRevoked
	}

	return s.IssueTokenPairForUserID(ctx, userID)
}

func (s *Service) RevokeRefresh(ctx context.Context, token string) error {
	return s.refresh.Revoke(ctx, token)
}
