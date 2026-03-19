package usecase

import (
	"errors"
	"time"

	authmodel "cityhawk/backend/internal/auth/model"
	usermodel "cityhawk/backend/internal/user/model"
)

var (
	ErrTokenRevoked = errors.New("token revoked")
	ErrUserNotFound = errors.New("user not found")
)

type Service struct {
	accessTTL  time.Duration
	refreshTTL time.Duration

	users    UserReader
	refresh  RefreshSessionRepository
	tokens   TokenService
	nowUTCFn func() time.Time
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
		nowUTCFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *Service) IssueTokenPair(u usermodel.User) (authmodel.TokenPair, error) {
	accessToken, err := s.tokens.SignAccessToken(u, s.accessTTL)
	if err != nil {
		return authmodel.TokenPair{}, err
	}

	refreshToken, err := s.tokens.GenerateOpaqueToken(32)
	if err != nil {
		return authmodel.TokenPair{}, err
	}

	s.refresh.Store(refreshToken, u.ID, s.nowUTCFn().Add(s.refreshTTL))

	return authmodel.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *Service) IssueTokenPairForUserID(userID string) (authmodel.TokenPair, error) {
	u, ok := s.users.GetByID(userID)
	if !ok {
		return authmodel.TokenPair{}, ErrUserNotFound
	}

	return s.IssueTokenPair(u)
}

func (s *Service) RotateRefresh(oldRefreshToken string) (authmodel.TokenPair, error) {
	userID, err := s.refresh.Consume(oldRefreshToken)
	if err != nil {
		return authmodel.TokenPair{}, ErrTokenRevoked
	}

	return s.IssueTokenPairForUserID(userID)
}

func (s *Service) RevokeRefresh(token string) {
	s.refresh.Revoke(token)
}

