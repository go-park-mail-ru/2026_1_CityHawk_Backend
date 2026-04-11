package errors

import "errors"

var (
	ErrMethodNotAllowed   = errors.New("method not allowed")
	ErrInvalidJSON        = errors.New("invalid json")
	ErrInternal           = errors.New("internal error")
	ErrIssueTokens        = errors.New("failed to issue tokens")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrMissingRefresh     = errors.New("missing refresh token")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidOAuthCode   = errors.New("invalid oauth code")
	ErrOAuthProfileFetch  = errors.New("failed to fetch oauth profile")
	ErrMissingAccess      = errors.New("missing access token")
	ErrInvalidAccess      = errors.New("invalid access token")
	ErrPlaceNotFound      = errors.New("place not found")
	ErrEventNotFound      = errors.New("event not found")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCity        = errors.New("invalid city")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidReference   = errors.New("invalid reference")
)
