package http

type taxonomyItemResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type eventCardNextSessionPlaceResponse struct {
	Name        string                        `json:"name"`
	AddressLine string                        `json:"addressLine"`
	ID          string                        `json:"id"`
	Latitude    float64                       `json:"latitude"`
	Longitude   float64                       `json:"longitude"`
	City        eventSessionPlaceCityResponse `json:"city"`
}

type eventCardNextSessionResponse struct {
	StartAt string                            `json:"startAt"`
	Place   eventCardNextSessionPlaceResponse `json:"place"`
}

type eventCardResponse struct {
	ID               string                        `json:"id"`
	Title            string                        `json:"title"`
	ShortDescription string                        `json:"shortDescription"`
	CoverImageURL    string                        `json:"coverImageUrl"`
	IsFavorite       bool                          `json:"isFavorite"`
	Tags             []taxonomyItemResponse        `json:"tags"`
	NextSession      *eventCardNextSessionResponse `json:"nextSession"`
}

type eventListResponse struct {
	Items  []eventCardResponse `json:"items"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

type eventAuthorResponse struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatarUrl"`
}

type eventImageResponse struct {
	ID       string `json:"id"`
	ImageURL string `json:"imageUrl"`
}

type eventSessionPlaceCityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CountryName string `json:"countryName"`
	Timezone    string `json:"timezone"`
}

type eventSessionPlaceResponse struct {
	ID          string                        `json:"id"`
	Name        string                        `json:"name"`
	AddressLine string                        `json:"addressLine"`
	Latitude    float64                       `json:"latitude"`
	Longitude   float64                       `json:"longitude"`
	City        eventSessionPlaceCityResponse `json:"city"`
}

type eventSessionResponse struct {
	ID        string                    `json:"id"`
	StartAt   string                    `json:"startAt"`
	EndAt     string                    `json:"endAt"`
	Price     int                       `json:"price"`
	PlaceName string                    `json:"placeName"`
	Place     eventSessionPlaceResponse `json:"place"`
}

type eventDetailsResponse struct {
	ID               string                     `json:"id"`
	Title            string                     `json:"title"`
	ShortDescription string                     `json:"shortDescription"`
	FullDescription  string                     `json:"fullDescription"`
	AgeLimit         int                        `json:"ageLimit"`
	SourceURL        *string                    `json:"sourceUrl"`
	Author           eventAuthorResponse        `json:"author"`
	PlaceName        string                     `json:"placeName,omitempty"`
	Place            *eventSessionPlaceResponse `json:"place,omitempty"`
	Categories       []taxonomyItemResponse     `json:"categories"`
	Tags             []taxonomyItemResponse     `json:"tags"`
	Images           []eventImageResponse       `json:"images"`
	Sessions         []eventSessionResponse     `json:"sessions"`
	CreatedAt        string                     `json:"createdAt"`
	UpdatedAt        string                     `json:"updatedAt"`
	IsFavorite       bool                       `json:"isFavorite"`
	IsOwner          bool                       `json:"isOwner"`
}

type eventIDResponse struct {
	ID string `json:"id"`
}

type categoriesResponse struct {
	Items []taxonomyItemResponse `json:"items"`
}

type tagsResponse struct {
	Items []taxonomyItemResponse `json:"items"`
}

type cityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CountryName string `json:"countryName"`
	Timezone    string `json:"timezone"`
}

type citiesResponse struct {
	Items []cityResponse `json:"items"`
}

type collectionCardResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	IsPublic    bool   `json:"isPublic"`
}

type collectionsResponse struct {
	Items []collectionCardResponse `json:"items"`
}

type collectionDetailsResponse struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	ImageURL    string              `json:"imageUrl"`
	IsPublic    bool                `json:"isPublic"`
	Events      []eventCardResponse `json:"events"`
}

type searchSuggestionResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Label       string  `json:"label"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	IsFollowing bool    `json:"isFollowing,omitempty"`
}

type searchSuggestionsResponse struct {
	Items []searchSuggestionResponse `json:"items"`
}

type homeTagResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type homeNextSessionPlaceResponse struct {
	Name        string `json:"name"`
	AddressLine string `json:"addressLine"`
}

type homeNextSessionResponse struct {
	StartAt string                       `json:"startAt"`
	Place   homeNextSessionPlaceResponse `json:"place"`
}

type homeFeaturedEventResponse struct {
	ID            string                  `json:"id"`
	Title         string                  `json:"title"`
	CoverImageURL string                  `json:"coverImageUrl"`
	Tags          []homeTagResponse       `json:"tags"`
	NextSession   homeNextSessionResponse `json:"nextSession"`
}

type homeCategoryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type homeCollectionResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
}

type homePayloadResponse struct {
	FeaturedEvents []homeFeaturedEventResponse `json:"featuredEvents"`
	Categories     []homeCategoryResponse      `json:"categories"`
	Collections    []homeCollectionResponse    `json:"collections"`
}

type mapCollectionResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	EventsCount int    `json:"eventsCount"`
	IsPublic    bool   `json:"isPublic"`
}

type mapCollectionsResponse struct {
	Items []mapCollectionResponse `json:"items"`
}

type mapOptionResponse struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type mapFiltersResponse struct {
	Tags        []taxonomyItemResponse `json:"tags"`
	DatePresets []mapOptionResponse    `json:"datePresets"`
	SortOptions []mapOptionResponse    `json:"sortOptions"`
}

type mapSpotCollectionResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type mapSpotResponse struct {
	ID         string                 `json:"id"`
	EventID    string                 `json:"eventId"`
	Title      string                 `json:"title"`
	Address    string                 `json:"address"`
	Latitude   float64                `json:"latitude"`
	Longitude  float64                `json:"longitude"`
	ImageURL   string                 `json:"imageUrl"`
	StartAt    string                 `json:"startAt"`
	Popularity int                    `json:"popularity"`
	Tags       []taxonomyItemResponse `json:"tags"`
}

type mapSpotsResponse struct {
	Collection mapSpotCollectionResponse `json:"collection"`
	Items      []mapSpotResponse         `json:"items"`
	Total      int                       `json:"total"`
	Limit      int                       `json:"limit"`
	Offset     int                       `json:"offset"`
}
