package http

import (
	"time"

	supportmodel "cityhawk/backend/internal/support/model"
)

func toTicketResponse(item supportmodel.Ticket) ticketResponse {
	var closedAt *string
	if item.ClosedAt != nil {
		value := item.ClosedAt.UTC().Format(time.RFC3339)
		closedAt = &value
	}

	return ticketResponse{
		ID:        item.ID,
		Category:  string(item.Category),
		Status:    string(item.Status),
		Title:     item.Title,
		Message:   item.Message,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
		ClosedAt:  closedAt,
	}
}

func toTicketListResponse(items []supportmodel.Ticket, limit int, offset int) ticketListResponse {
	responses := make([]ticketResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toTicketResponse(item))
	}
	return ticketListResponse{
		Items:  responses,
		Limit:  limit,
		Offset: offset,
	}
}

func toTicketStatsResponse(stats supportmodel.Stats) ticketStatsResponse {
	byStatus := make(map[string]int, len(stats.ByStatus))
	for status, count := range stats.ByStatus {
		byStatus[string(status)] = count
	}

	byCategory := make(map[string]int, len(stats.ByCategory))
	for category, count := range stats.ByCategory {
		byCategory[string(category)] = count
	}

	return ticketStatsResponse{
		Total:           stats.Total,
		ByStatus:        byStatus,
		ByCategory:      byCategory,
		OpenTotal:       stats.OpenTotal,
		InProgressTotal: stats.InProgressTotal,
		ClosedTotal:     stats.ClosedTotal,
	}
}

func toTicketMessageResponse(item supportmodel.Message) ticketMessageResponse {
	return ticketMessageResponse{
		ID:           item.ID,
		TicketID:     item.TicketID,
		AuthorUserID: item.AuthorUserID,
		AuthorRole:   string(item.AuthorRole),
		Body:         item.Body,
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toTicketMessageListResponse(items []supportmodel.Message) ticketMessageListResponse {
	responses := make([]ticketMessageResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toTicketMessageResponse(item))
	}
	return ticketMessageListResponse{Items: responses}
}
