package http

type meResponse struct {
	ID             string        `json:"id"`
	Email          string        `json:"email"`
	Username       string        `json:"username"`
	Role           string        `json:"role"`
	Birthday       *string       `json:"birthday"`
	Bio            *string       `json:"bio"`
	InterestTagIDs []string      `json:"interestTagIds"`
	AvatarURL      *string       `json:"avatarUrl"`
	City           *cityResponse `json:"city"`
	CreatedAt      string        `json:"createdAt"`
}

type patchMeResponse struct {
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	Username       string   `json:"username"`
	Role           string   `json:"role"`
	Birthday       *string  `json:"birthday"`
	Bio            *string  `json:"bio"`
	InterestTagIDs []string `json:"interestTagIds"`
	AvatarURL      *string  `json:"avatarUrl"`
	UpdatedAt      string   `json:"updatedAt"`
}

type cityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CountryName string `json:"countryName"`
	Timezone    string `json:"timezone"`
}
