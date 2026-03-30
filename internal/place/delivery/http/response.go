package http

import placemodel "cityhawk/backend/internal/place/model"

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

func toPlaceResponse(p placemodel.Place) placeResponse {
	return placeResponse{
		ID:               p.ID,
		Title:            p.Title,
		Categories:       p.Categories,
		LikeCount:        p.LikeCount,
		ShortDescription: p.ShortDescription,
		FullDescription:  p.FullDescription,
		Address:          p.Address,
		ImageURL:         p.ImageURL,
		WorkingHours:     p.WorkingHours,
		PriceLevel:       p.PriceLevel,
	}
}

func toPlaceCardResponses(items []placemodel.PlaceCard) []placeCardResponse {
	resp := make([]placeCardResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, placeCardResponse{
			ID:               item.ID,
			Title:            item.Title,
			Categories:       item.Categories,
			LikeCount:        item.LikeCount,
			ShortDescription: item.ShortDescription,
			Address:          item.Address,
			ImageURL:         item.ImageURL,
		})
	}
	return resp
}

func toHomePayloadResponse(p placemodel.HomePayload) homePayloadResponse {
	places := make([]homePlaceCardResponse, 0, len(p.Places))
	for _, place := range p.Places {
		places = append(places, homePlaceCardResponse{
			ID:          place.ID,
			ImageURL:    place.ImageURL,
			Title:       place.Title,
			Description: place.Description,
		})
	}

	moodLeft := make([]moodCardResponse, 0, len(p.MoodLeft))
	for _, mood := range p.MoodLeft {
		moodLeft = append(moodLeft, moodCardResponse{
			ID:       mood.ID,
			ImageURL: mood.ImageURL,
			Title:    mood.Title,
			Modifier: mood.Modifier,
		})
	}

	return homePayloadResponse{
		Places:   places,
		MoodLeft: moodLeft,
		MoodTall: homeMoodTallResponse{
			ID:       p.MoodTall.ID,
			ImageURL: p.MoodTall.ImageURL,
			Title:    p.MoodTall.Title,
		},
	}
}

