package repository

import (
	"context"
	"sync"
	"time"

	platformerrors "cityhawk/backend/internal/platform/errors"
	platformid "cityhawk/backend/internal/platform/id"
	usermodel "cityhawk/backend/internal/user/model"
)

type InMemoryUserRepository struct {
	mu      sync.RWMutex
	byID    map[string]usermodel.User
	byEmail map[string]usermodel.User
	cities  map[string]usermodel.City
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		byID:    make(map[string]usermodel.User),
		byEmail: make(map[string]usermodel.User),
		cities: map[string]usermodel.City{
			"11111111-1111-1111-1111-111111111111": {
				ID:          "11111111-1111-1111-1111-111111111111",
				Name:        "Moscow",
				CountryName: "Russia",
				Timezone:    "Europe/Moscow",
			},
		},
	}
}

func (r *InMemoryUserRepository) Create(_ context.Context, u usermodel.User) (usermodel.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byEmail[u.Email]; exists {
		return usermodel.User{}, platformerrors.ErrEmailExists
	}

	if u.ID == "" {
		u.ID = platformid.NewUUIDUserIDProvider().New()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = u.CreatedAt
	}
	if u.Role == "" {
		u.Role = usermodel.RoleUser
	}
	if u.CityID != nil {
		if city, ok := r.cities[*u.CityID]; ok {
			copied := city
			u.City = &copied
		}
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

func (r *InMemoryUserRepository) UpdateProfile(_ context.Context, id string, patch usermodel.ProfilePatch) (usermodel.User, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.byID[id]
	if !ok {
		return usermodel.User{}, false, nil
	}

	if patch.Email != nil && *patch.Email != u.Email {
		if existing, exists := r.byEmail[*patch.Email]; exists && existing.ID != id {
			return usermodel.User{}, false, platformerrors.ErrEmailExists
		}
		delete(r.byEmail, u.Email)
		u.Email = *patch.Email
	}
	if patch.Username != nil {
		u.Username = *patch.Username
	}
	if patch.UserSurname != nil {
		u.UserSurname = *patch.UserSurname
	}
	if patch.Birthday != nil {
		birthday := patch.Birthday.UTC()
		u.Birthday = &birthday
	}
	if patch.CityID != nil {
		u.CityID = patch.CityID
		if city, exists := r.cities[*patch.CityID]; exists {
			copied := city
			u.City = &copied
		} else {
			u.City = nil
		}
	}
	if patch.AvatarURL != nil {
		u.AvatarURL = patch.AvatarURL
	}
	u.UpdatedAt = time.Now().UTC()

	r.byID[id] = u
	r.byEmail[u.Email] = u
	return u, true, nil
}
