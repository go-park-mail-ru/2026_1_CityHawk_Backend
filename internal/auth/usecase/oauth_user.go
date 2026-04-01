package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	authmodel "cityhawk/backend/internal/auth/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
)

type OAuthUserService struct {
	users     AuthUserRepository
	passwords PasswordService
	ids       UserIDProvider
}

func NewOAuthUserService(users AuthUserRepository, passwords PasswordService, ids UserIDProvider) *OAuthUserService {
	return &OAuthUserService{
		users:     users,
		passwords: passwords,
		ids:       ids,
	}
}

func (s *OAuthUserService) FindOrCreateFromOAuth(identity authmodel.OAuthIdentity) (usermodel.User, error) {
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

	if u, ok := s.users.GetByEmail(email); ok {
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

	return s.createOAuthUser(email, username)
}

func (s *OAuthUserService) createOAuthUser(email, username string) (usermodel.User, error) {
	password, err := randomHex(16)
	if err != nil {
		return usermodel.User{}, err
	}

	passwordHash, err := s.passwords.Hash(password)
	if err != nil {
		return usermodel.User{}, err
	}

	u := usermodel.User{
		ID:           s.ids.New(),
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
	}

	persistedUser, err := s.users.Create(u)
	if err != nil {
		if errors.Is(err, platformerrors.ErrEmailExists) {
			if existing, ok := s.users.GetByEmail(email); ok {
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

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
