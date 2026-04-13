package repository

import (
	"context"
	"sync"
	"time"

	platformerrors "cityhawk/backend/internal/platform/errors"
	platformsecurity "cityhawk/backend/internal/platform/security"
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

func (r *InMemoryRefreshRepository) Store(_ context.Context, refreshToken, userID string, expiresAt time.Time) error {
	hash := platformsecurity.SHA256Hex(refreshToken)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[hash] = refreshSession{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
	return nil
}

func (r *InMemoryRefreshRepository) Consume(_ context.Context, refreshToken string) (string, error) {
	hash := platformsecurity.SHA256Hex(refreshToken)

	r.mu.Lock()
	defer r.mu.Unlock()

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

func (r *InMemoryRefreshRepository) Revoke(_ context.Context, refreshToken string) error {
	hash := platformsecurity.SHA256Hex(refreshToken)

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, hash)
	return nil
}
