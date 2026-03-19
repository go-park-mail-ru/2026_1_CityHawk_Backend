package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var (
	ErrRefreshTokenRevoked = errors.New("token revoked")
	ErrRefreshTokenExpired = errors.New("token expired")
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

func (r *InMemoryRefreshRepository) Store(refreshToken, userID string, expiresAt time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessions[tokenHash(refreshToken)] = refreshSession{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}
}

func (r *InMemoryRefreshRepository) Consume(refreshToken string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	hash := tokenHash(refreshToken)
	s, ok := r.sessions[hash]
	if !ok {
		return "", ErrRefreshTokenRevoked
	}

	if time.Now().UTC().After(s.ExpiresAt) {
		delete(r.sessions, hash)
		return "", ErrRefreshTokenExpired
	}

	delete(r.sessions, hash)
	return s.UserID, nil
}

func (r *InMemoryRefreshRepository) Revoke(refreshToken string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, tokenHash(refreshToken))
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

