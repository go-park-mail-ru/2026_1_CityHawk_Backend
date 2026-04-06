package http

type messageResponse struct {
	Message string `json:"message"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}
