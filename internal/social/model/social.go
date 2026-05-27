package model

import "time"

type FavoriteEvent struct {
	EventID   string
	CreatedAt time.Time
}

type NotificationEventRef struct {
	EventID    string
	CreatedAt  time.Time
	InvitedBy  *NotificationEventInviter
	Invitation *NotificationInvitation
}

type NotificationEventInviter struct {
	ID        string
	Username  string
	AvatarURL *string
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

type InviteeCandidate struct {
	ID               string
	Username         string
	AvatarURL        *string
	City             *City
	IsFollowing      bool
	IsFriend         bool
	InvitationStatus *string
}

type Invitation struct {
	ID             string
	EventID        string
	EventSessionID *string
	SenderID       string
	RecipientID    string
	Status         string
	Message        string
	RespondedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Notification struct {
	ID           string
	Type         string
	Message      string
	IsRead       bool
	ReadAt       *time.Time
	CreatedAt    time.Time
	Actor        *NotificationActor
	Event        *NotificationEvent
	EventSession *NotificationEventSession
	Invitation   *NotificationInvitation
	Collection   *NotificationCollection
}

type NotificationActor struct {
	ID          string
	DisplayName string
	AvatarURL   *string
}

type NotificationEvent struct {
	ID            string
	Title         string
	CoverImageURL string
	DateText      string
	PlaceText     string
}

type NotificationEventSession struct {
	ID      string
	StartAt time.Time
	EndAt   time.Time
}

type NotificationInvitation struct {
	ID     string
	Status string
}

type NotificationCollection struct {
	ID       string
	Title    string
	ImageURL string
}

type ShareLink struct {
	ID           string
	Token        string
	URL          string
	EventID      *string
	CollectionID *string
	CreatedAt    time.Time
}
