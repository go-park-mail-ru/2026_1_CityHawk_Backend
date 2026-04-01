package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	platformerrors "cityhawk/backend/internal/platform/errors"
)

type refreshSession struct {
	UserID    string
	ExpiresAt time.Time
}

type InMemoryRefreshRepository struct {
	mu       sync.Mutex
	sessions map[string]refreshSession
}

func NewInMemoryRefreshRepository() *InMemoryRefreshRepository {
	return &InMemoryRefreshRepository{
		sessions: make(map[string]refreshSession),
	}
}

func (r *InMemoryRefreshRepository) Store(refreshToken, userID string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[tokenHash(refreshToken)] = refreshSession{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	return nil
}

func (r *InMemoryRefreshRepository) Consume(refreshToken string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hash := tokenHash(refreshToken)
	s, ok := r.sessions[hash]
	if !ok {
		return "", platformerrors.ErrTokenRevoked
	}

	if time.Now().UTC().After(s.ExpiresAt) {
		delete(r.sessions, hash)
		return "", platformerrors.ErrTokenExpired
	}

	delete(r.sessions, hash)
	return s.UserID, nil
}

func (r *InMemoryRefreshRepository) Revoke(refreshToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, tokenHash(refreshToken))
	return nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
