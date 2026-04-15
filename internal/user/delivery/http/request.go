package http

type patchMeRequest struct {
	Username    *string `json:"username"`
	UserSurname *string `json:"userSurname"`
	Birthday    *string `json:"birthday"`
	CityID      *string `json:"cityId"`
	AvatarURL   *string `json:"avatarUrl"`
}
