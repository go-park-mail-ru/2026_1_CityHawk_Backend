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
	ErrMissingAccess      = errors.New("missing access token")
	ErrInvalidAccess      = errors.New("invalid access token")
	ErrPlaceNotFound      = errors.New("place not found")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrEmailExists        = errors.New("email already exists")
)

