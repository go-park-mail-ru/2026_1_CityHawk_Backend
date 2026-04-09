package http

type registerRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	UserSurname string `json:"userSurname"`
	Password    string `json:"password"`
	Birthday    string `json:"birthday"`
	CityID      string `json:"cityId"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
