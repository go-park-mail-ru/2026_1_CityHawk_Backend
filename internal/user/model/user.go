package model

import "time"

type City struct {
	ID          string
	Name        string
	CountryName string
	Timezone    string
}

type User struct {
	ID           string
	Email        string
	Username     string
	UserSurname  string
	PasswordHash string
	Birthday     *time.Time
	CityID       *string
	City         *City
	AvatarURL    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ProfilePatch struct {
	Email       *string
	Username    *string
	UserSurname *string
	Birthday    *time.Time
	CityID      *string
	AvatarURL   *string
}
