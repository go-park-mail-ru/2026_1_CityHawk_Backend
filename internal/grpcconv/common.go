package grpcconv

import (
	usermodel "cityhawk/backend/internal/user/model"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
)

func UserRoleToProto(role usermodel.Role) commonv1.UserRole {
	switch role {
	case usermodel.RoleAdmin:
		return commonv1.UserRole_USER_ROLE_ADMIN
	case usermodel.RoleUser:
		return commonv1.UserRole_USER_ROLE_USER
	case usermodel.RoleOrganizer:
		return commonv1.UserRole_USER_ROLE_USER
	default:
		return commonv1.UserRole_USER_ROLE_UNSPECIFIED
	}
}

func UserRoleFromProto(role commonv1.UserRole) usermodel.Role {
	switch role {
	case commonv1.UserRole_USER_ROLE_ADMIN:
		return usermodel.RoleAdmin
	case commonv1.UserRole_USER_ROLE_USER:
		return usermodel.RoleUser
	default:
		return ""
	}
}

func UserIDFromContext(ctx *commonv1.UserContext) string {
	if ctx == nil || !ctx.Authenticated {
		return ""
	}
	return ctx.UserId
}
