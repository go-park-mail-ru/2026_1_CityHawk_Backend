package http

type registerRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Birthday string `json:"birthday"`
	CityID   string `json:"cityId"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
