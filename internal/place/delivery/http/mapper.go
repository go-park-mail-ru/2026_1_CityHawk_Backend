package http

import placemodel "cityhawk/backend/internal/place/model"

func toEventListResponse(items []placemodel.EventCardView, total, limit, offset int) eventListResponse {
	respItems := make([]eventCardResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, toEventCardResponse(item))
	}

	return eventListResponse{
		Items:  respItems,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}

func toEventDetailsResponse(item placemodel.EventDetailsView) eventDetailsResponse {
	categories := make([]taxonomyItemResponse, 0, len(item.Categories))
	for _, category := range item.Categories {
		categories = append(categories, taxonomyItemResponse{
			ID:   category.ID,
			Name: category.Name,
			Slug: category.Slug,
		})
	}

	tags := make([]taxonomyItemResponse, 0, len(item.Tags))
	for _, tag := range item.Tags {
		tags = append(tags, taxonomyItemResponse{
			ID:   tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}

	images := make([]eventImageResponse, 0, len(item.Images))
	for _, image := range item.Images {
		images = append(images, eventImageResponse{
			ID:       image.ID,
			ImageURL: image.ImageURL,
		})
	}

	sessions := make([]eventSessionResponse, 0, len(item.Sessions))
	for _, session := range item.Sessions {
		sessions = append(sessions, eventSessionResponse{
			ID:      session.ID,
			StartAt: session.StartAt.UTC().Format("2006-01-02T15:04:05Z"),
			EndAt:   session.EndAt.UTC().Format("2006-01-02T15:04:05Z"),
			Price:   session.Price,
			Place: eventSessionPlaceResponse{
				ID:          session.Place.ID,
				Name:        session.Place.Name,
				AddressLine: session.Place.AddressLine,
				Latitude:    session.Place.Latitude,
				Longitude:   session.Place.Longitude,
				City: eventSessionPlaceCityResponse{
					ID:          session.Place.City.ID,
					Name:        session.Place.City.Name,
					CountryName: session.Place.City.CountryName,
					Timezone:    session.Place.City.Timezone,
				},
			},
		})
	}

	return eventDetailsResponse{
		ID:               item.ID,
		Title:            item.Title,
		ShortDescription: item.ShortDescription,
		FullDescription:  item.FullDescription,
		AgeLimit:         item.AgeLimit,
		SourceURL:        item.SourceURL,
		Author: eventAuthorResponse{
			ID:        item.Author.ID,
			Username:  item.Author.Username,
			AvatarURL: item.Author.AvatarURL,
		},
		Categories: categories,
		Tags:       tags,
		Images:     images,
		Sessions:   sessions,
		CreatedAt:  item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		IsFavorite: item.IsFavorite,
		IsOwner:    item.IsOwner,
	}
}

func toCategoriesResponse(items []placemodel.HomeCategory) categoriesResponse {
	respItems := make([]taxonomyItemResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, taxonomyItemResponse{
			ID:   item.ID,
			Name: item.Name,
			Slug: item.Slug,
		})
	}
	return categoriesResponse{Items: respItems}
}

func toTagsResponse(items []placemodel.HomeTag) tagsResponse {
	respItems := make([]taxonomyItemResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, taxonomyItemResponse{
			ID:   item.ID,
			Name: item.Name,
			Slug: item.Slug,
		})
	}
	return tagsResponse{Items: respItems}
}

func toCollectionsResponse(items []placemodel.CollectionCardView) collectionsResponse {
	respItems := make([]collectionCardResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, collectionCardResponse{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			ImageURL:    item.ImageURL,
			IsPublic:    item.IsPublic,
		})
	}
	return collectionsResponse{Items: respItems}
}

func toCollectionDetailsResponse(item placemodel.CollectionDetailsView) collectionDetailsResponse {
	events := make([]eventCardResponse, 0, len(item.Events))
	for _, event := range item.Events {
		events = append(events, toEventCardResponse(event))
	}

	return collectionDetailsResponse{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		ImageURL:    item.ImageURL,
		IsPublic:    item.IsPublic,
		Events:      events,
	}
}

func toSearchSuggestionsResponse(items []placemodel.SearchSuggestion) searchSuggestionsResponse {
	respItems := make([]string, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, item.Name)
	}
	return searchSuggestionsResponse{Items: respItems}
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

func toEventCardResponse(item placemodel.EventCardView) eventCardResponse {
	tags := make([]taxonomyItemResponse, 0, len(item.Tags))
	for _, tag := range item.Tags {
		tags = append(tags, taxonomyItemResponse{
			ID:   tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}

	var nextSession *eventCardNextSessionResponse
	if item.NextSession != nil {
		nextSession = &eventCardNextSessionResponse{
			StartAt: item.NextSession.StartAt.UTC().Format("2006-01-02T15:04:05Z"),
			Place: eventCardNextSessionPlaceResponse{
				Name:        item.NextSession.Place.Name,
				AddressLine: item.NextSession.Place.AddressLine,
			},
		}
	}

	return eventCardResponse{
		ID:               item.ID,
		Title:            item.Title,
		ShortDescription: item.ShortDescription,
		CoverImageURL:    item.CoverImageURL,
		Tags:             tags,
		NextSession:      nextSession,
	}
}
