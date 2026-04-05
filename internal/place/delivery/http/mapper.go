package http

import placemodel "cityhawk/backend/internal/place/model"

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
