package usecase

import (
	"context"
	"errors"

	authmodel "cityhawk/backend/internal/auth/model"
	authvalidation "cityhawk/backend/internal/auth/validation"
	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
)

type AuthUserRepository interface {
	Create(ctx context.Context, u usermodel.User) (usermodel.User, error)
	GetByEmail(ctx context.Context, email string) (usermodel.User, bool)
}

type PasswordService interface {
	Hash(password string) (string, error)
	Verify(password, stored string) bool
}

type UserIDProvider interface {
	New() string
}

type tokenPairIssuer interface {
	IssueTokenPair(ctx context.Context, u usermodel.User) (authmodel.TokenPair, error)
}

type AuthFlowService struct {
	users     AuthUserRepository
	issuer    tokenPairIssuer
	passwords PasswordService
	ids       UserIDProvider
}

func NewAuthFlowService(
	users AuthUserRepository,
	issuer tokenPairIssuer,
	passwords PasswordService,
	ids UserIDProvider,
) *AuthFlowService {
	return &AuthFlowService{
		users:     users,
		issuer:    issuer,
		passwords: passwords,
		ids:       ids,
	}
}

func (s *AuthFlowService) Register(ctx context.Context, email, password, username string) (authmodel.TokenPair, error) {
	normalizedEmail, normalizedPassword, normalizedUsername, err := authvalidation.ValidateRegister(email, password, username)
	if err != nil {
		return authmodel.TokenPair{}, err
	}

	passwordHash, err := s.passwords.Hash(normalizedPassword)
	if err != nil {
		return authmodel.TokenPair{}, platformerrors.ErrInternal
	}

	u := usermodel.User{
		ID:           s.ids.New(),
		Email:        normalizedEmail,
		Username:     normalizedUsername,
		PasswordHash: passwordHash,
	}

	persistedUser, err := s.users.Create(ctx, u)
	if err != nil {
		if errors.Is(err, platformerrors.ErrEmailExists) {
			return authmodel.TokenPair{}, platformerrors.ErrEmailExists
		}
		return authmodel.TokenPair{}, err
	}

	resp, err := s.issuer.IssueTokenPair(ctx, persistedUser)
	if err != nil {
		return authmodel.TokenPair{}, platformerrors.ErrIssueTokens
	}
	return resp, nil
}

func (s *AuthFlowService) Login(ctx context.Context, email, password string) (authmodel.TokenPair, error) {
	normalizedEmail, normalizedPassword, err := authvalidation.ValidateLogin(email, password)
	if err != nil {
		return authmodel.TokenPair{}, err
	}

	u, ok := s.users.GetByEmail(ctx, normalizedEmail)
	if !ok || !s.passwords.Verify(normalizedPassword, u.PasswordHash) {
		return authmodel.TokenPair{}, platformerrors.ErrInvalidCredentials
	}

	resp, err := s.issuer.IssueTokenPair(ctx, u)
	if err != nil {
		return authmodel.TokenPair{}, platformerrors.ErrIssueTokens
	}
	return resp, nil
}
