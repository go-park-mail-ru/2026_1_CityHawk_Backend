package http

import "time"

type messageResponse struct {
	Message string `json:"message"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

type okResponse struct {
	OK bool `json:"ok"`
}

type registerResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatarUrl"`
	CreatedAt time.Time `json:"createdAt"`
}

type loginResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
