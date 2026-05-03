package grpcconv

import (
	"errors"

	platformerrors "cityhawk/backend/internal/platform/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Error(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, platformerrors.ErrInvalidCredentials),
		errors.Is(err, platformerrors.ErrMissingAccess),
		errors.Is(err, platformerrors.ErrInvalidAccess),
		errors.Is(err, platformerrors.ErrMissingRefresh),
		errors.Is(err, platformerrors.ErrInvalidRefresh),
		errors.Is(err, platformerrors.ErrTokenExpired),
		errors.Is(err, platformerrors.ErrTokenRevoked),
		errors.Is(err, platformerrors.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, platformerrors.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, platformerrors.ErrUserNotFound),
		errors.Is(err, platformerrors.ErrPlaceNotFound),
		errors.Is(err, platformerrors.ErrEventNotFound),
		errors.Is(err, platformerrors.ErrCategoryNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, platformerrors.ErrEmailExists),
		errors.Is(err, platformerrors.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, platformerrors.ErrInvalidCity),
		errors.Is(err, platformerrors.ErrInvalidReference):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func InvalidArgument(message string) error {
	return status.Error(codes.InvalidArgument, message)
}

func NotFound(message string) error {
	return status.Error(codes.NotFound, message)
}
