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
