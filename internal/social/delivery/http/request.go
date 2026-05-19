package http

type createInvitationsRequest struct {
	RecipientIDs   []string `json:"recipientIds"`
	Message        string   `json:"message"`
	EventSessionID string   `json:"eventSessionId"`
}

type updateInvitationRequest struct {
	Status string `json:"status"`
}
