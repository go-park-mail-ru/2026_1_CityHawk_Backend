package repository

import (
	"context"
	"sync"

	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
)

type InMemoryUserRepository struct {
	mu      sync.RWMutex
	byID    map[string]usermodel.User
	byEmail map[string]usermodel.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		byID:    make(map[string]usermodel.User),
		byEmail: make(map[string]usermodel.User),
	}
}

func (r *InMemoryUserRepository) Create(_ context.Context, u usermodel.User) (usermodel.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byEmail[u.Email]; exists {
		return usermodel.User{}, platformerrors.ErrEmailExists
	}

	r.byEmail[u.Email] = u
	r.byID[u.ID] = u
	return u, nil
}

func (r *InMemoryUserRepository) GetByEmail(_ context.Context, email string) (usermodel.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byEmail[email]
	return u, ok
}

func (r *InMemoryUserRepository) GetByID(_ context.Context, id string) (usermodel.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	return u, ok
}
