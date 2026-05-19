package model

import "time"

type EventTaxonomyItem struct {
	ID   string
	Name string
	Slug string
}

type EventCardNextSessionPlace struct {
	Name        string
	AddressLine string
	ID          string
	Latitude    float64
	Longitude   float64
	City        EventSessionPlaceCityView
}

type EventCardNextSession struct {
	StartAt time.Time
	Place   EventCardNextSessionPlace
}

type EventCardView struct {
	ID               string
	Title            string
	ShortDescription string
	CoverImageURL    string
	Tags             []EventTaxonomyItem
	NextSession      *EventCardNextSession
	IsFavorite       bool
	Popularity       int
}

type EventAuthorView struct {
	ID        string
	Username  string
	AvatarURL *string
}

type EventImageView struct {
	ID       string
	ImageURL string
}

type EventSessionPlaceCityView struct {
	ID          string
	Name        string
	CountryName string
	Timezone    string
}

type EventSessionPlaceView struct {
	ID          string
	Name        string
	AddressLine string
	Latitude    float64
	Longitude   float64
	City        EventSessionPlaceCityView
}

type EventSessionView struct {
	ID      string
	StartAt time.Time
	EndAt   time.Time
	Price   int
	Place   EventSessionPlaceView
}

type EventDetailsView struct {
	ID               string
	Title            string
	ShortDescription string
	FullDescription  string
	AgeLimit         int
	SourceURL        *string
	Author           EventAuthorView
	Categories       []EventTaxonomyItem
	Tags             []EventTaxonomyItem
	Images           []EventImageView
	Sessions         []EventSessionView
	CreatedAt        time.Time
	UpdatedAt        time.Time
	IsFavorite       bool
	IsOwner          bool
}

type EventListFilter struct {
	Query        string
	CategoryID   string
	TagID        string
	CityID       string
	DateFrom     *time.Time
	DateTo       *time.Time
	AuthorID     string
	CollectionID string
	Sort         string
	Limit        int
	Offset       int
	UserID       string
}

type HomeFilter struct {
	City string
}

type EventSessionInput struct {
	PlaceID string
	StartAt time.Time
	EndAt   time.Time
	Price   int
}

type EventWriteInput struct {
	ID               string
	AuthorUserID     string
	Title            *string
	ShortDescription *string
	FullDescription  *string
	AgeLimit         *int
	SourceURL        *string
	ClearSourceURL   bool
	CategoryIDs      *[]string
	TagIDs           *[]string
	ImageURLs        *[]string
	Sessions         *[]EventSessionInput
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

type CollectionCardView struct {
	ID          string
	Title       string
	Description string
	ImageURL    string
	IsPublic    bool
}

type CollectionDetailsView struct {
	ID          string
	Title       string
	Description string
	ImageURL    string
	IsPublic    bool
	Events      []EventCardView
}

type SearchSuggestion struct {
	ID          string
	Type        string
	Title       string
	Label       string
	AvatarURL   *string
	IsFollowing bool
}

type HomePayload struct {
	FeaturedEvents []HomeFeaturedEvent
	Categories     []HomeCategory
	Collections    []HomeCollection
}
