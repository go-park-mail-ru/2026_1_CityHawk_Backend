package model

import "time"

type FavoriteEvent struct {
	EventID   string
	CreatedAt time.Time
}

type UserFollow struct {
	UserID    string
	CreatedAt time.Time
}

type City struct {
	ID          string
	Name        string
	CountryName string
	Timezone    string
}

type UserProfile struct {
	ID          string
	Username    string
	UserSurname string
	AvatarURL   *string
	City        *City
	IsFollowing bool
}

type CollectionCard struct {
	ID          string
	Title       string
	Description string
	ImageURL    string
	IsPublic    bool
}

type FavoriteFlag struct {
	EventID    string
	IsFavorite bool
	CreatedAt  *time.Time
}

type FollowingFlag struct {
	UserID      string
	IsFollowing bool
	CreatedAt   *time.Time
}
