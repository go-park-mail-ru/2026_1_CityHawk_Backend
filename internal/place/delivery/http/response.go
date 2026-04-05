package http

type placeResponse struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Categories       []string `json:"categories"`
	LikeCount        int      `json:"like_count"`
	ShortDescription string   `json:"short_description"`
	FullDescription  string   `json:"full_description"`
	Address          string   `json:"address"`
	ImageURL         string   `json:"image_url"`
	WorkingHours     string   `json:"working_hours"`
	PriceLevel       string   `json:"price_level"`
}

type placeCardResponse struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Categories       []string `json:"categories"`
	LikeCount        int      `json:"like_count"`
	ShortDescription string   `json:"short_description"`
	Address          string   `json:"address"`
	ImageURL         string   `json:"image_url"`
}

type homePlaceCardResponse struct {
	ID          string `json:"id"`
	ImageURL    string `json:"imageUrl"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type moodCardResponse struct {
	ID       string `json:"id"`
	ImageURL string `json:"imageUrl"`
	Title    string `json:"title"`
	Modifier string `json:"modifier,omitempty"`
}

type homeMoodTallResponse struct {
	ID       string `json:"id"`
	ImageURL string `json:"imageUrl"`
	Title    string `json:"title"`
}

type homePayloadResponse struct {
	Places   []homePlaceCardResponse `json:"places"`
	MoodLeft []moodCardResponse      `json:"moodLeft"`
	MoodTall homeMoodTallResponse    `json:"moodTall"`
}
