package model

import "time"

type City struct {
	ID          string
	Name        string
	CountryName string
	Timezone    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Place struct {
	ID          string
	CityID      string
	Name        string
	AddressLine string
	Latitude    float64
	Longitude   float64
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Category struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Tag struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Event struct {
	ID                  string
	AuthorUserID        string
	Title               string
	LocationDescription string
	FullDescription     string
	AgeLimit            int
	SourceURL           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type EventSession struct {
	ID        string
	EventID   string
	PlaceID   string
	StartAt   time.Time
	EndAt     time.Time
	Price     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EventImage struct {
	ID        string
	EventID   string
	ImageURL  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EventCategory struct {
	EventID    string
	CategoryID string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type EventTag struct {
	EventID   string
	TagID     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Collection struct {
	ID           string
	AuthorUserID string
	Title        string
	Description  string
	IsPublic     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CollectionImage struct {
	ID           string
	CollectionID string
	ImageURL     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CollectionEvent struct {
	CollectionID string
	EventID      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type FavoriteEvent struct {
	UserID    string
	EventID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserFollow struct {
	FollowerUserID string
	FollowedUserID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type EventInvitation struct {
	ID              string
	SenderUserID    string
	RecipientUserID string
	MessageText     string
	RespondedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type EventInvitationEvent struct {
	InvitationID string
	EventID      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type EventInvitationSession struct {
	InvitationID   string
	EventSessionID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ShareLink struct {
	ID            string
	CreatorUserID string
	ShareToken    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ShareLinkEvent struct {
	ShareLinkID string
	EventID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ShareLinkCollection struct {
	ShareLinkID  string
	CollectionID string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Notification struct {
	ID               string
	RecipientUserID  string
	NotificationType string
	IsRead           bool
	ReadAt           *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type NotificationActor struct {
	NotificationID string
	AuthorUserID   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type NotificationEvent struct {
	NotificationID string
	EventID        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type NotificationEventSession struct {
	NotificationID string
	EventSessionID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type NotificationInvitation struct {
	NotificationID string
	InvitationID   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type NotificationCollection struct {
	NotificationID string
	CollectionID   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
