package usecase

import (
	"errors"

	authmodel "cityhawk/backend/internal/auth/model"
	usermodel "cityhawk/backend/internal/user/model"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrIssueTokens        = errors.New("failed to issue tokens")
	ErrInternal           = errors.New("internal error")
)

type AuthUserRepository interface {
	Create(u usermodel.User) error
	GetByEmail(email string) (usermodel.User, bool)
}

type CredentialsValidator interface {
	ValidateRegister(email, password, username string) (string, string, string, error)
	ValidateLogin(email, password string) (string, string, error)
}

type PasswordService interface {
	Hash(password string) (string, error)
	Verify(password, stored string) bool
}

type UserIDProvider interface {
	New() string
}

type tokenPairIssuer interface {
	IssueTokenPair(u usermodel.User) (authmodel.TokenPair, error)
}

type AuthFlowService struct {
	users     AuthUserRepository
	issuer    tokenPairIssuer
	validator CredentialsValidator
	passwords PasswordService
	ids       UserIDProvider
}

func NewAuthFlowService(
	users AuthUserRepository,
	issuer tokenPairIssuer,
	validator CredentialsValidator,
	passwords PasswordService,
	ids UserIDProvider,
) *AuthFlowService {
	return &AuthFlowService{
		users:     users,
		issuer:    issuer,
		validator: validator,
		passwords: passwords,
		ids:       ids,
	}
}

func (s *AuthFlowService) Register(email, password, username string) (authmodel.TokenPair, error) {
	normalizedEmail, normalizedPassword, normalizedUsername, err := s.validator.ValidateRegister(email, password, username)
	if err != nil {
		return authmodel.TokenPair{}, err
	}

	passwordHash, err := s.passwords.Hash(normalizedPassword)
	if err != nil {
		return authmodel.TokenPair{}, ErrInternal
	}

	u := usermodel.User{
		ID:           s.ids.New(),
		Email:        normalizedEmail,
		Username:     normalizedUsername,
		PasswordHash: passwordHash,
	}

	if err := s.users.Create(u); err != nil {
		if errors.Is(err, usermodel.ErrEmailExists) {
			return authmodel.TokenPair{}, ErrEmailExists
		}
		return authmodel.TokenPair{}, err
	}

	resp, err := s.issuer.IssueTokenPair(u)
	if err != nil {
		return authmodel.TokenPair{}, ErrIssueTokens
	}
	return resp, nil
}

func (s *AuthFlowService) Login(email, password string) (authmodel.TokenPair, error) {
	normalizedEmail, normalizedPassword, err := s.validator.ValidateLogin(email, password)
	if err != nil {
		return authmodel.TokenPair{}, err
	}

	u, ok := s.users.GetByEmail(normalizedEmail)
	if !ok || !s.passwords.Verify(normalizedPassword, u.PasswordHash) {
		return authmodel.TokenPair{}, ErrInvalidCredentials
	}

	resp, err := s.issuer.IssueTokenPair(u)
	if err != nil {
		return authmodel.TokenPair{}, ErrIssueTokens
	}
	return resp, nil
}
