package model

import "time"

type EventDetailsView struct {
	ID                  string
	Title               string
	Categories          []string
	LikeCount           int
	LocationDescription string
	FullDescription     string
	AddressLine         string
	ImageURL            string
	SessionLabel        string
	PriceLevel          string
}

type EventCardView struct {
	ID                  string
	Title               string
	Categories          []string
	LikeCount           int
	LocationDescription string
	AddressLine         string
	ImageURL            string
}

type HomeTag struct {
	ID   string
	Name string
	Slug string
}

type HomeNextSessionPlace struct {
	Name        string
	AddressLine string
}

type HomeNextSession struct {
	StartAt time.Time
	Place   HomeNextSessionPlace
}

type HomeFeaturedEvent struct {
	ID            string
	Title         string
	CoverImageURL string
	Tags          []HomeTag
	NextSession   HomeNextSession
}

type HomeCategory struct {
	ID   string
	Name string
	Slug string
}

type HomeCollection struct {
	ID          string
	Title       string
	Description string
	ImageURL    string
}

type HomePayload struct {
	FeaturedEvents []HomeFeaturedEvent
	Categories     []HomeCategory
	Collections    []HomeCollection
}
