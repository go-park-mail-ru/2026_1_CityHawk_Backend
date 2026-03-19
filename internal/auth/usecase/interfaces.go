package usecase

import (
	"time"

	authmodel "cityhawk/backend/internal/auth/model"
	usermodel "cityhawk/backend/internal/user/model"
)

type UserReader interface {
	GetByID(id string) (usermodel.User, bool)
}

type RefreshSessionRepository interface {
	Store(refreshToken, userID string, expiresAt time.Time)
	Consume(refreshToken string) (string, error)
	Revoke(refreshToken string)
}

type TokenService interface {
	SignAccessToken(u usermodel.User, ttl time.Duration) (string, error)
	GenerateOpaqueToken(size int) (string, error)
}

type AuthUsecase interface {
	IssueTokenPair(u usermodel.User) (authmodel.TokenPair, error)
	IssueTokenPairForUserID(userID string) (authmodel.TokenPair, error)
	RotateRefresh(oldRefreshToken string) (authmodel.TokenPair, error)
	RevokeRefresh(token string)
}

type AuthFlowUsecase interface {
	Register(email, password, username string) (authmodel.TokenPair, error)
	Login(email, password string) (authmodel.TokenPair, error)
}
