package main

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	maxEmailLen    = 254
	minPasswordLen = 8
	maxPasswordLen = 72
	minUsernameLen = 3
	maxUsernameLen = 32
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9_.-]+$`)

func validateRegisterInput(req registerRequest) (string, string, string, error) {
	email, err := normalizeAndValidateEmail(req.Email)
	if err != nil {
		return "", "", "", err
	}

	password, err := normalizeAndValidatePassword(req.Password)
	if err != nil {
		return "", "", "", err
	}

	username, err := normalizeAndValidateUsername(req.Username)
	if err != nil {
		return "", "", "", err
	}

	return email, password, username, nil
}

func validateLoginInput(req loginRequest) (string, string, error) {
	email, err := normalizeAndValidateEmail(req.Email)
	if err != nil {
		return "", "", err
	}

	password, err := normalizeAndValidatePassword(req.Password)
	if err != nil {
		return "", "", err
	}

	return email, password, nil
}

func normalizeAndValidateEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", errors.New("email is required")
	}
	if len(email) > maxEmailLen {
		return "", errors.New("email is too long")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", errors.New("invalid email format")
	}
	return email, nil
}

func normalizeAndValidatePassword(raw string) (string, error) {
	password := strings.TrimSpace(raw)
	if password == "" {
		return "", errors.New("password is required")
	}
	if utf8.RuneCountInString(password) < minPasswordLen {
		return "", errors.New("password must be at least 8 characters")
	}
	if len(password) > maxPasswordLen {
		return "", errors.New("password is too long")
	}
	return password, nil
}

func normalizeAndValidateUsername(raw string) (string, error) {
	username := strings.TrimSpace(raw)
	if username == "" {
		return "", errors.New("username is required")
	}
	usernameLen := utf8.RuneCountInString(username)
	if usernameLen < minUsernameLen || usernameLen > maxUsernameLen {
		return "", errors.New("username must be 3-32 characters")
	}
	if !usernamePattern.MatchString(username) {
		return "", errors.New("username contains invalid characters")
	}
	return username, nil
}
