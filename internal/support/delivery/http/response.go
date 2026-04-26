package http

type ticketResponse struct {
	ID        string  `json:"id"`
	Category  string  `json:"category"`
	Status    string  `json:"status"`
	Title     string  `json:"title"`
	Message   string  `json:"message"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	ClosedAt  *string `json:"closedAt"`
}

type ticketListResponse struct {
	Items  []ticketResponse `json:"items"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

type ticketStatsResponse struct {
	Total           int            `json:"total"`
	ByStatus        map[string]int `json:"byStatus"`
	ByCategory      map[string]int `json:"byCategory"`
	OpenTotal       int            `json:"openTotal"`
	InProgressTotal int            `json:"inProgressTotal"`
	ClosedTotal     int            `json:"closedTotal"`
}

type ticketMessageResponse struct {
	ID           string `json:"id"`
	TicketID     string `json:"ticketId"`
	AuthorUserID string `json:"authorUserId"`
	AuthorRole   string `json:"authorRole"`
	Body         string `json:"body"`
	CreatedAt    string `json:"createdAt"`
}

type ticketMessageListResponse struct {
	Items []ticketMessageResponse `json:"items"`
}
