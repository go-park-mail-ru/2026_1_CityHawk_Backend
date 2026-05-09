package grpcconv

import (
	"errors"
	"testing"
	"time"

	platformerrors "cityhawk/backend/internal/platform/errors"
	usermodel "cityhawk/backend/internal/user/model"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestErrorMapsKnownPlatformErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "nil", err: nil, code: codes.OK},
		{name: "unauthenticated", err: platformerrors.ErrInvalidCredentials, code: codes.Unauthenticated},
		{name: "permission denied", err: platformerrors.ErrForbidden, code: codes.PermissionDenied},
		{name: "not found", err: platformerrors.ErrUserNotFound, code: codes.NotFound},
		{name: "already exists", err: platformerrors.ErrEmailExists, code: codes.AlreadyExists},
		{name: "invalid argument", err: platformerrors.ErrInvalidCity, code: codes.InvalidArgument},
		{name: "wrapped", err: errors.Join(errors.New("context"), platformerrors.ErrEventNotFound), code: codes.NotFound},
		{name: "unknown", err: errors.New("boom"), code: codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Error(tt.err)
			if tt.err == nil {
				if err != nil {
					t.Fatalf("Error(nil) = %v, want nil", err)
				}
				return
			}
			if got := status.Code(err); got != tt.code {
				t.Fatalf("status code = %s, want %s", got, tt.code)
			}
		})
	}
}

func TestStatusHelpers(t *testing.T) {
	if got := status.Code(InvalidArgument("bad input")); got != codes.InvalidArgument {
		t.Fatalf("InvalidArgument code = %s", got)
	}
	if got := status.Code(NotFound("missing")); got != codes.NotFound {
		t.Fatalf("NotFound code = %s", got)
	}
}

func TestTimeConversions(t *testing.T) {
	if TimeToProto(time.Time{}) != nil {
		t.Fatal("zero time converted to non-nil timestamp")
	}
	if OptionalTimeToProto(nil) != nil {
		t.Fatal("nil optional time converted to non-nil timestamp")
	}

	value := time.Date(2026, 5, 4, 12, 30, 0, 0, time.FixedZone("MSK", 3*60*60))
	protoValue := TimeToProto(value)
	if protoValue == nil {
		t.Fatal("TimeToProto returned nil for non-zero time")
	}
	if got := protoValue.AsTime(); !got.Equal(value.UTC()) {
		t.Fatalf("proto time = %v, want %v", got, value.UTC())
	}

	fromProto := TimeFromProto(timestamppb.New(value))
	if fromProto == nil || !fromProto.Equal(value.UTC()) {
		t.Fatalf("TimeFromProto = %v, want %v", fromProto, value.UTC())
	}
	if TimeFromProto(nil) != nil {
		t.Fatal("nil proto time converted to non-nil time")
	}
}

func TestUserRoleConversions(t *testing.T) {
	tests := []struct {
		role usermodel.Role
		want commonv1.UserRole
	}{
		{role: usermodel.RoleAdmin, want: commonv1.UserRole_USER_ROLE_ADMIN},
		{role: usermodel.RoleUser, want: commonv1.UserRole_USER_ROLE_USER},
		{role: usermodel.RoleOrganizer, want: commonv1.UserRole_USER_ROLE_USER},
		{role: "unknown", want: commonv1.UserRole_USER_ROLE_UNSPECIFIED},
	}

	for _, tt := range tests {
		if got := UserRoleToProto(tt.role); got != tt.want {
			t.Fatalf("UserRoleToProto(%q) = %s, want %s", tt.role, got, tt.want)
		}
	}

	if got := UserRoleFromProto(commonv1.UserRole_USER_ROLE_ADMIN); got != usermodel.RoleAdmin {
		t.Fatalf("admin role from proto = %q", got)
	}
	if got := UserRoleFromProto(commonv1.UserRole_USER_ROLE_USER); got != usermodel.RoleUser {
		t.Fatalf("user role from proto = %q", got)
	}
	if got := UserRoleFromProto(commonv1.UserRole_USER_ROLE_UNSPECIFIED); got != "" {
		t.Fatalf("unspecified role from proto = %q, want empty", got)
	}
}

func TestUserIDFromContext(t *testing.T) {
	if got := UserIDFromContext(nil); got != "" {
		t.Fatalf("nil context user id = %q", got)
	}
	if got := UserIDFromContext(&commonv1.UserContext{Authenticated: false, UserId: "user-1"}); got != "" {
		t.Fatalf("anonymous context user id = %q", got)
	}
	if got := UserIDFromContext(&commonv1.UserContext{Authenticated: true, UserId: "user-1"}); got != "user-1" {
		t.Fatalf("authenticated context user id = %q", got)
	}
}
