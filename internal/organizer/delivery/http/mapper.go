package http

import (
	organizermodel "cityhawk/backend/internal/organizer/model"
	"cityhawk/backend/internal/platform/safety"
)

func toApplicationCreateResponse(item organizermodel.Application) applicationCreateResponse {
	return applicationCreateResponse{
		ID:        item.ID,
		Status:    string(item.Status),
		CreatedAt: item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func toApplicationResponse(item organizermodel.Application) applicationResponse {
	return applicationResponse{
		ID:            item.ID,
		Status:        string(item.Status),
		Name:          safety.EscapeText(item.Name),
		Email:         safety.EscapeText(item.Email),
		Phone:         safety.EscapeText(item.Phone),
		City:          safety.EscapeText(item.City),
		ProjectName:   safety.EscapeText(item.ProjectName),
		Categories:    safety.EscapeText(item.Categories),
		Links:         safety.EscapeText(item.Links),
		About:         safety.EscapeText(item.About),
		ReviewComment: safety.EscapeText(item.ReviewComment),
		CreatedAt:     item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func toApplicationStatusResponse(item organizermodel.Application) applicationStatusResponse {
	return applicationStatusResponse{
		ID:            item.ID,
		Status:        string(item.Status),
		ReviewComment: safety.EscapeText(item.ReviewComment),
		UpdatedAt:     item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
