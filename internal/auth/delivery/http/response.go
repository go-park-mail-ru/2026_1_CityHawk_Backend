package http

type messageResponse struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}
