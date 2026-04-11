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

func (s *AuthFlowService) Register(ctx context.Context, email, username, userSurname, password, birthday, cityID string) (authmodel.RegistrationResult, error) {
	normalizedEmail, normalizedUsername, normalizedSurname, normalizedPassword, normalizedBirthday, normalizedCityID, err := authvalidation.ValidateRegister(email, username, userSurname, password, birthday, cityID)
	if err != nil {
		return authmodel.RegistrationResult{}, err
	}

	passwordHash, err := s.passwords.Hash(normalizedPassword)
	if err != nil {
		return authmodel.RegistrationResult{}, platformerrors.ErrInternal
	}

	u := usermodel.User{
		Email:        normalizedEmail,
		Username:     normalizedUsername,
		UserSurname:  normalizedSurname,
		PasswordHash: passwordHash,
		Birthday:     normalizedBirthday,
		CityID:       normalizedCityID,
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

func (s *AuthFlowService) Login(ctx context.Context, email, password string) (authmodel.SessionResult, error) {
	normalizedEmail, normalizedPassword, err := authvalidation.ValidateLogin(email, password)
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
