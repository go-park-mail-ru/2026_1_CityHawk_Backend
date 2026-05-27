package validation

import (
	"errors"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"cityhawk/backend/internal/platform/media"
)

const (
	maxEmailLen    = 254
	minPasswordLen = 8
	maxPasswordLen = 72
	minUsernameLen = 3
	maxUsernameLen = 32
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9_.-]+$`)
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type ValidationError struct {
	Details map[string]string
}

func (e ValidationError) Error() string {
	return "validation failed"
}

func ValidateRegister(email, username, password string) (string, string, string, error) {
	details := make(map[string]string)

	normalizedEmail, err := normalizeAndValidateEmail(email)
	if err != nil {
		details["email"] = err.Error()
	}

	normalizedUsername, err := normalizeAndValidateUsername(username)
	if err != nil {
		details["username"] = err.Error()
	}

	normalizedPassword, err := normalizeAndValidatePassword(password)
	if err != nil {
		details["password"] = err.Error()
	}

	if len(details) > 0 {
		return "", "", "", ValidationError{Details: details}
	}

	return normalizedEmail, normalizedUsername, normalizedPassword, nil
}

func ValidateLogin(email, password string) (string, string, error) {
	details := make(map[string]string)

	normalizedEmail, err := normalizeAndValidateEmail(email)
	if err != nil {
		details["email"] = err.Error()
	}

	normalizedPassword, err := normalizeAndValidatePassword(password)
	if err != nil {
		details["password"] = err.Error()
	}

	if len(details) > 0 {
		return "", "", ValidationError{Details: details}
	}

	return normalizedEmail, normalizedPassword, nil
}

func ValidateProfilePatch(email, username, birthday, cityID, avatarURL, bio *string, interestTagIDs []string) (string, string, *time.Time, string, string, string, []string, error) {
	details := make(map[string]string)

	var normalizedEmail string
	if email != nil {
		value, err := normalizeAndValidateEmail(*email)
		if err != nil {
			details["email"] = err.Error()
		} else {
			normalizedEmail = value
		}
	}

	var normalizedUsername string
	if username != nil {
		value, err := normalizeAndValidateUsername(*username)
		if err != nil {
			details["username"] = err.Error()
		} else {
			normalizedUsername = value
		}
	}

	var normalizedBirthday *time.Time
	if birthday != nil {
		value, err := normalizeAndValidateBirthday(*birthday)
		if err != nil {
			details["birthday"] = err.Error()
		} else {
			normalizedBirthday = &value
		}
	}

	var normalizedCityID string
	if cityID != nil {
		value, err := normalizeAndValidateUUID(*cityID, "cityId")
		if err != nil {
			details["cityId"] = err.Error()
		} else {
			normalizedCityID = value
		}
	}

	var normalizedAvatarURL string
	if avatarURL != nil {
		value, err := normalizeAndValidateAvatarURL(*avatarURL)
		if err != nil {
			details["avatarUrl"] = err.Error()
		} else {
			normalizedAvatarURL = value
		}
	}

	var normalizedBio string
	if bio != nil {
		normalizedBio = strings.TrimSpace(*bio)
		if len([]rune(normalizedBio)) > 1000 {
			details["bio"] = "bio is too long"
		}
	}

	normalizedInterestTagIDs := make([]string, 0, len(interestTagIDs))
	seenInterestTagIDs := make(map[string]struct{}, len(interestTagIDs))
	for i, raw := range interestTagIDs {
		value, err := normalizeAndValidateUUID(raw, "interestTagIds")
		if err != nil {
			details["interestTagIds"] = "interestTagIds[" + strconv.Itoa(i) + "] must be uuid"
			continue
		}
		if _, ok := seenInterestTagIDs[value]; ok {
			continue
		}
		seenInterestTagIDs[value] = struct{}{}
		normalizedInterestTagIDs = append(normalizedInterestTagIDs, value)
	}

	if len(details) > 0 {
		return "", "", nil, "", "", "", nil, ValidationError{Details: details}
	}

	return normalizedEmail, normalizedUsername, normalizedBirthday, normalizedCityID, normalizedAvatarURL, normalizedBio, normalizedInterestTagIDs, nil
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

func normalizeAndValidateBirthday(raw string) (time.Time, error) {
	birthday := strings.TrimSpace(raw)
	if birthday == "" {
		return time.Time{}, errors.New("birthday is required")
	}
	parsed, err := time.Parse("2006-01-02", birthday)
	if err != nil {
		return time.Time{}, errors.New("birthday must be in YYYY-MM-DD format")
	}
	return parsed.UTC(), nil
}

func normalizeAndValidateUUID(raw, field string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New(field + " is required")
	}
	if !uuidPattern.MatchString(value) {
		return "", errors.New(field + " must be a valid UUID")
	}
	return strings.ToLower(value), nil
}

func normalizeAndValidateAvatarURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("avatarUrl is required")
	}
	if utf8.RuneCountInString(value) > 2048 {
		return "", errors.New("avatarUrl is too long")
	}
	if !media.IsFileReference(value) {
		return "", errors.New("avatarUrl must be an http(s) URL or /uploads path")
	}
	return value, nil
}
