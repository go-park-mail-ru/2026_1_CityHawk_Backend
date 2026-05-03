package http

type patchMeRequest struct {
	Email          *string  `json:"email"`
	Username       *string  `json:"username"`
	UserSurname    *string  `json:"userSurname"`
	Birthday       *string  `json:"birthday"`
	CityID         *string  `json:"cityId"`
	Bio            *string  `json:"bio"`
	InterestTagIDs []string `json:"interestTagIds"`
	AvatarURL      *string  `json:"avatarUrl"`
}
