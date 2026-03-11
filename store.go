package main

import (
	"errors"
	"sync"
	"time"
)

type user struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
}

type userStore struct {
	mu      sync.RWMutex
	byEmail map[string]user
}

type refreshSession struct {
	UserID    string
	ExpiresAt time.Time
}

type refreshStore struct {
	mu      sync.RWMutex
	session map[string]refreshSession
}

func (s *userStore) create(u user) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byEmail[u.Email]; exists {
		return errors.New("email already exists")
	}
	s.byEmail[u.Email] = u
	return nil
}

func (s *userStore) getByEmail(email string) (user, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byEmail[email]
	return u, ok
}

func (s *userStore) getByID(id string) (user, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.byEmail {
		if u.ID == id {
			return u, true
		}
	}
	return user{}, false
}
