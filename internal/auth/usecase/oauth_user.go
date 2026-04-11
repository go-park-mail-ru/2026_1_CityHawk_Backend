package usecase

import (
	"context"
	"errors"
	"strings"

	authmodel "cityhawk/backend/internal/auth/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	platformsecurity "cityhawk/backend/internal/platform/security"
	usermodel "cityhawk/backend/internal/user/model"
)

type OAuthUserService struct {
	users     AuthUserRepository
	passwords PasswordService
}

func NewOAuthUserService(users AuthUserRepository, passwords PasswordService) *OAuthUserService {
	return &OAuthUserService{
		users:     users,
		passwords: passwords,
	}
}

func (s *OAuthUserService) FindOrCreateFromOAuth(ctx context.Context, identity authmodel.OAuthIdentity) (usermodel.User, error) {
	provider := strings.ToLower(strings.TrimSpace(identity.Provider))
	if provider == "" {
		return usermodel.User{}, errors.New("oauth provider is empty")
	}

	subjectID := sanitizePart(identity.SubjectID)
	email := strings.ToLower(strings.TrimSpace(identity.Email))
	if email == "" {
		if subjectID == "" {
			return usermodel.User{}, errors.New("oauth user identity is empty")
		}
		email = provider + "_" + subjectID + "@" + provider + ".local"
	}

	if u, ok := s.users.GetByEmail(ctx, email); ok {
		return u, nil
	}

	username := sanitizePart(identity.Username)
	if username == "" {
		if subjectID != "" {
			username = provider + "_" + subjectID
		} else {
			username = provider + "_user"
		}
	}

	if len(username) < 3 {
		username = provider + "_user"
	}
	if len(username) > 32 {
		username = username[:32]
	}

	return s.createOAuthUser(ctx, email, username)
}

func (s *OAuthUserService) createOAuthUser(ctx context.Context, email, username string) (usermodel.User, error) {
	password, err := platformsecurity.RandomHex(16)
	if err != nil {
		return usermodel.User{}, err
	}

	passwordHash, err := s.passwords.Hash(password)
	if err != nil {
		return usermodel.User{}, err
	}

	u := usermodel.User{
		Email:        email,
		Username:     username,
		UserSurname:  username,
		PasswordHash: passwordHash,
	}

	persistedUser, err := s.users.Create(ctx, u)
	if err != nil {
		if errors.Is(err, platformerrors.ErrEmailExists) {
			if existing, ok := s.users.GetByEmail(ctx, email); ok {
				return existing, nil
			}
		}
		return usermodel.User{}, err
	}

	return persistedUser, nil
}

func sanitizePart(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
