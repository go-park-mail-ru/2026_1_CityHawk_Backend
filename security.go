package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const userIDTimeLayout = "20060102150405.000000000"

func createUser(email, password, username string) (user, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return user{}, err
	}
	return user{
		ID:           time.Now().UTC().Format(userIDTimeLayout),
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
	}, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifyPassword(password, stored string) bool {
	return bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) == nil
}
