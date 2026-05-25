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

type tokenPairIssuer interface {
	IssueTokenPair(ctx context.Context, u usermodel.User) (authmodel.TokenPair, error)
}

type AuthFlowService struct {
	users     AuthUserRepository
	issuer    tokenPairIssuer
	passwords PasswordService
}

func NewAuthFlowService(
	users AuthUserRepository,
	issuer tokenPairIssuer,
	passwords PasswordService,
) *AuthFlowService {
	return &AuthFlowService{
		users:     users,
		issuer:    issuer,
		passwords: passwords,
	}
}

func (s *AuthFlowService) Register(ctx context.Context, input authmodel.RegisterInput) (authmodel.RegistrationResult, error) {
	normalizedEmail, normalizedPassword, err := authvalidation.ValidateRegister(
		input.Email,
		input.Password,
	)
	if err != nil {
		return authmodel.RegistrationResult{}, err
	}

	passwordHash, err := s.passwords.Hash(normalizedPassword)
	if err != nil {
		return authmodel.RegistrationResult{}, platformerrors.ErrInternal
	}

	u := usermodel.User{
		Email:        normalizedEmail,
		Username:     "",
		UserSurname:  "",
		PasswordHash: passwordHash,
	}

	persistedUser, err := s.users.Create(ctx, u)
	if err != nil {
		if errors.Is(err, platformerrors.ErrEmailExists) {
			return authmodel.RegistrationResult{}, platformerrors.ErrEmailExists
		}
		return authmodel.RegistrationResult{}, err
	}

	resp, err := s.issuer.IssueTokenPair(ctx, persistedUser)
	if err != nil {
		return authmodel.RegistrationResult{}, platformerrors.ErrIssueTokens
	}
	return authmodel.RegistrationResult{
		User:   persistedUser,
		Tokens: resp,
	}, nil
}

func (s *AuthFlowService) Login(ctx context.Context, input authmodel.LoginInput) (authmodel.SessionResult, error) {
	normalizedEmail, normalizedPassword, err := authvalidation.ValidateLogin(input.Email, input.Password)
	if err != nil {
		return authmodel.SessionResult{}, err
	}

	u, ok := s.users.GetByEmail(ctx, normalizedEmail)
	if !ok || !s.passwords.Verify(normalizedPassword, u.PasswordHash) {
		return authmodel.SessionResult{}, platformerrors.ErrInvalidCredentials
	}

	resp, err := s.issuer.IssueTokenPair(ctx, u)
	if err != nil {
		return authmodel.SessionResult{}, platformerrors.ErrIssueTokens
	}
	return authmodel.SessionResult{
		User:   u,
		Tokens: resp,
	}, nil
}
