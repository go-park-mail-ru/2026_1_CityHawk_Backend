package http

type placeResponse struct {
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	Categories          []string `json:"categories"`
	LikeCount           int      `json:"like_count"`
	LocationDescription string   `json:"location_description"`
	FullDescription     string   `json:"full_description"`
	Address             string   `json:"address"`
	ImageURL            string   `json:"image_url"`
	WorkingHours        string   `json:"working_hours"`
	PriceLevel          string   `json:"price_level"`
}

type placeCardResponse struct {
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	Categories          []string `json:"categories"`
	LikeCount           int      `json:"like_count"`
	LocationDescription string   `json:"location_description"`
	Address             string   `json:"address"`
	ImageURL            string   `json:"image_url"`
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
