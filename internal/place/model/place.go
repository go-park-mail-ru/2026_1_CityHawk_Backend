package model

type Place struct {
	ID               string
	Title            string
	Categories       []string
	LikeCount        int
	ShortDescription string
	FullDescription  string
	Address          string
	ImageURL         string
	WorkingHours     string
	PriceLevel       string
}

type PlaceCard struct {
	ID               string
	Title            string
	Categories       []string
	LikeCount        int
	ShortDescription string
	Address          string
	ImageURL         string
}

type HomePlaceCard struct {
	ID          string
	ImageURL    string
	Title       string
	Description string
}

type MoodCard struct {
	ID       string
	ImageURL string
	Title    string
	Modifier string
}

type HomeMoodTall struct {
	ID       string
	ImageURL string
	Title    string
}

type HomePayload struct {
	Places   []HomePlaceCard
	MoodLeft []MoodCard
	MoodTall HomeMoodTall
}
