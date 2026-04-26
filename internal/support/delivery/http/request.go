package http

type createTicketRequest struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}

type updateTicketRequest struct {
	Category *string `json:"category"`
	Title    *string `json:"title"`
	Message  *string `json:"message"`
}

type updateTicketStatusRequest struct {
	Status string `json:"status"`
}

type createTicketMessageRequest struct {
	Body string `json:"body"`
}
