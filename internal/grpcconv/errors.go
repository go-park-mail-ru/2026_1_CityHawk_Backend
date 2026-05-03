package grpcconv

import (
	"errors"

	platformerrors "cityhawk/backend/internal/platform/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errorCodeMapping struct {
	code    codes.Code
	targets []error
}

var errorCodeMappings = []errorCodeMapping{
	{
		code: codes.Unauthenticated,
		targets: []error{
			platformerrors.ErrInvalidCredentials,
			platformerrors.ErrMissingAccess,
			platformerrors.ErrInvalidAccess,
			platformerrors.ErrMissingRefresh,
			platformerrors.ErrInvalidRefresh,
			platformerrors.ErrTokenExpired,
			platformerrors.ErrTokenRevoked,
			platformerrors.ErrUnauthorized,
		},
	},
	{
		code: codes.PermissionDenied,
		targets: []error{
			platformerrors.ErrForbidden,
		},
	},
	{
		code: codes.NotFound,
		targets: []error{
			platformerrors.ErrUserNotFound,
			platformerrors.ErrPlaceNotFound,
			platformerrors.ErrEventNotFound,
			platformerrors.ErrCategoryNotFound,
		},
	},
	{
		code: codes.AlreadyExists,
		targets: []error{
			platformerrors.ErrEmailExists,
			platformerrors.ErrAlreadyExists,
		},
	},
	{
		code: codes.InvalidArgument,
		targets: []error{
			platformerrors.ErrInvalidCity,
			platformerrors.ErrInvalidReference,
		},
	},
}

func Error(err error) error {
	if err == nil {
		return nil
	}

	return status.Error(codeForError(err), err.Error())
}

func codeForError(err error) codes.Code {
	for _, mapping := range errorCodeMappings {
		if isAny(err, mapping.targets...) {
			return mapping.code
		}
	}
	return codes.Internal
}

func isAny(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func InvalidArgument(message string) error {
	return status.Error(codes.InvalidArgument, message)
}

func NotFound(message string) error {
	return status.Error(codes.NotFound, message)
}
