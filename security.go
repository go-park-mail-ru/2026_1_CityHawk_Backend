package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

func createUser(email, password string) (user, error) {
	salt, err := randomHex(16)
	if err != nil {
		return user{}, err
	}
	return user{
		ID:           time.Now().UTC().Format("20060102150405.000000000"),
		Email:        email,
		PasswordHash: hashPassword(password, salt),
	}, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return salt + ":" + hex.EncodeToString(sum[:])
}

func verifyPassword(password, stored string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return false
	}
	return hashPassword(password, parts[0]) == stored
}
