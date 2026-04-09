package http

import placemodel "cityhawk/backend/internal/place/model"

func toPlaceResponse(p placemodel.EventDetailsView) placeResponse {
	return placeResponse{
		ID:                  p.ID,
		Title:               p.Title,
		Categories:          p.Categories,
		LikeCount:           p.LikeCount,
		LocationDescription: p.LocationDescription,
		FullDescription:     p.FullDescription,
		Address:             p.AddressLine,
		ImageURL:            p.ImageURL,
		WorkingHours:        p.SessionLabel,
		PriceLevel:          p.PriceLevel,
	}
}

func toPlaceCardResponses(items []placemodel.EventCardView) []placeCardResponse {
	resp := make([]placeCardResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, placeCardResponse{
			ID:                  item.ID,
			Title:               item.Title,
			Categories:          item.Categories,
			LikeCount:           item.LikeCount,
			LocationDescription: item.LocationDescription,
			Address:             item.AddressLine,
			ImageURL:            item.ImageURL,
		})
	}
	return resp
}

func toHomePayloadResponse(p placemodel.HomePayload) homePayloadResponse {
	featuredEvents := make([]homeFeaturedEventResponse, 0, len(p.FeaturedEvents))
	for _, item := range p.FeaturedEvents {
		tags := make([]homeTagResponse, 0, len(item.Tags))
		for _, tag := range item.Tags {
			tags = append(tags, homeTagResponse{
				ID:   tag.ID,
				Name: tag.Name,
				Slug: tag.Slug,
			})
		}

		featuredEvents = append(featuredEvents, homeFeaturedEventResponse{
			ID:            item.ID,
			Title:         item.Title,
			CoverImageURL: item.CoverImageURL,
			Tags:          tags,
			NextSession: homeNextSessionResponse{
				StartAt: item.NextSession.StartAt.UTC().Format("2006-01-02T15:04:05Z"),
				Place: homeNextSessionPlaceResponse{
					Name:        item.NextSession.Place.Name,
					AddressLine: item.NextSession.Place.AddressLine,
				},
			},
		})
	}

	categories := make([]homeCategoryResponse, 0, len(p.Categories))
	for _, item := range p.Categories {
		categories = append(categories, homeCategoryResponse{
			ID:   item.ID,
			Name: item.Name,
			Slug: item.Slug,
		})
	}

	collections := make([]homeCollectionResponse, 0, len(p.Collections))
	for _, item := range p.Collections {
		collections = append(collections, homeCollectionResponse{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			ImageURL:    item.ImageURL,
		})
	}

	return homePayloadResponse{
		FeaturedEvents: featuredEvents,
		Categories:     categories,
		Collections:    collections,
	}
}
