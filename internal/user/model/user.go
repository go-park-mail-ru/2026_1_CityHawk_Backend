package model

import "time"

type City struct {
	ID          string
	Name        string
	CountryName string
	Timezone    string
}

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func (r Role) Valid() bool {
	switch r {
	case RoleUser, RoleAdmin:
		return true
	default:
		return false
	}
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
	Role         Role
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
