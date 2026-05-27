package http

import (
	placemodel "cityhawk/backend/internal/place/model"
	"cityhawk/backend/internal/platform/media"
	"cityhawk/backend/internal/platform/safety"
)

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
			ImageURL: media.PublicURL(image.ImageURL),
		})
	}

	sessions := make([]eventSessionResponse, 0, len(item.Sessions))
	for _, session := range item.Sessions {
		sessions = append(sessions, eventSessionResponse{
			ID:        session.ID,
			StartAt:   session.StartAt.UTC().Format("2006-01-02T15:04:05Z"),
			EndAt:     session.EndAt.UTC().Format("2006-01-02T15:04:05Z"),
			Price:     session.Price,
			PlaceName: safety.EscapeText(session.Place.Name),
			Place: eventSessionPlaceResponse{
				ID:          session.Place.ID,
				Name:        safety.EscapeText(session.Place.Name),
				AddressLine: safety.EscapeText(session.Place.AddressLine),
				Latitude:    session.Place.Latitude,
				Longitude:   session.Place.Longitude,
				City: eventSessionPlaceCityResponse{
					ID:          session.Place.City.ID,
					Name:        safety.EscapeText(session.Place.City.Name),
					CountryName: safety.EscapeText(session.Place.City.CountryName),
					Timezone:    session.Place.City.Timezone,
				},
			},
		})
	}

	var place *eventSessionPlaceResponse
	placeName := ""
	if eventPlace := primaryEventPlace(item); eventPlace != nil {
		place = toEventSessionPlaceResponse(*eventPlace)
		placeName = place.Name
	}

	return eventDetailsResponse{
		ID:               item.ID,
		Title:            safety.EscapeText(item.Title),
		ShortDescription: safety.EscapeText(item.ShortDescription),
		FullDescription:  safety.EscapeText(item.FullDescription),
		AgeLimit:         item.AgeLimit,
		SourceURL:        item.SourceURL,
		Author: eventAuthorResponse{
			ID:        item.Author.ID,
			Username:  safety.EscapeText(item.Author.Username),
			AvatarURL: media.PublicURLPtr(item.Author.AvatarURL),
		},
		PlaceName:  placeName,
		Place:      place,
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

func primaryEventPlace(item placemodel.EventDetailsView) *placemodel.EventSessionPlaceView {
	if item.Place != nil {
		return item.Place
	}
	if len(item.Sessions) == 0 || item.Sessions[0].Place.ID == "" {
		return nil
	}
	return &item.Sessions[0].Place
}

func toEventSessionPlaceResponse(item placemodel.EventSessionPlaceView) *eventSessionPlaceResponse {
	return &eventSessionPlaceResponse{
		ID:          item.ID,
		Name:        safety.EscapeText(item.Name),
		AddressLine: safety.EscapeText(item.AddressLine),
		Latitude:    item.Latitude,
		Longitude:   item.Longitude,
		City: eventSessionPlaceCityResponse{
			ID:          item.City.ID,
			Name:        safety.EscapeText(item.City.Name),
			CountryName: safety.EscapeText(item.City.CountryName),
			Timezone:    item.City.Timezone,
		},
	}
}

func toCategoriesResponse(items []placemodel.HomeCategory) categoriesResponse {
	respItems := make([]taxonomyItemResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, taxonomyItemResponse{
			ID:   item.ID,
			Name: safety.EscapeText(item.Name),
			Slug: item.Slug,
		})
	}
	return categoriesResponse{Items: respItems}
}

func toTagsResponse(items []placemodel.HomeTag) tagsResponse {
	respItems := make([]taxonomyItemResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, taxonomyItemResponse{
			ID:    item.ID,
			Name:  safety.EscapeText(item.Name),
			Slug:  item.Slug,
			Group: item.Group,
		})
	}
	return tagsResponse{Items: respItems}
}

func toCitiesResponse(items []placemodel.City) citiesResponse {
	respItems := make([]cityResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, cityResponse{
			ID:          item.ID,
			Name:        safety.EscapeText(item.Name),
			CountryName: safety.EscapeText(item.CountryName),
			Timezone:    item.Timezone,
		})
	}
	return citiesResponse{Items: respItems}
}

func toCollectionsResponse(items []placemodel.CollectionCardView) collectionsResponse {
	respItems := make([]collectionCardResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, collectionCardResponse{
			ID:          item.ID,
			Title:       safety.EscapeText(item.Title),
			Description: safety.EscapeText(item.Description),
			ImageURL:    media.PublicURL(item.ImageURL),
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
		Title:       safety.EscapeText(item.Title),
		Description: safety.EscapeText(item.Description),
		ImageURL:    media.PublicURL(item.ImageURL),
		IsPublic:    item.IsPublic,
		Events:      events,
	}
}

func toSearchSuggestionsResponse(items []placemodel.SearchSuggestion) searchSuggestionsResponse {
	respItems := make([]searchSuggestionResponse, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, searchSuggestionResponse{
			ID:    item.ID,
			Title: safety.EscapeText(item.Title),
			Label: safety.EscapeText(item.Label),
		})
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
				Name: safety.EscapeText(tag.Name),
				Slug: tag.Slug,
			})
		}

		startAt := ""
		if !item.NextSession.StartAt.IsZero() {
			startAt = item.NextSession.StartAt.UTC().Format("2006-01-02T15:04:05Z")
		}

		featuredEvents = append(featuredEvents, homeFeaturedEventResponse{
			ID:            item.ID,
			Title:         safety.EscapeText(item.Title),
			CoverImageURL: media.PublicURL(item.CoverImageURL),
			Tags:          tags,
			NextSession: homeNextSessionResponse{
				StartAt: startAt,
				Place: homeNextSessionPlaceResponse{
					Name:        safety.EscapeText(item.NextSession.Place.Name),
					AddressLine: safety.EscapeText(item.NextSession.Place.AddressLine),
				},
			},
		})
	}

	categories := make([]homeCategoryResponse, 0, len(p.Categories))
	for _, item := range p.Categories {
		categories = append(categories, homeCategoryResponse{
			ID:   item.ID,
			Name: safety.EscapeText(item.Name),
			Slug: item.Slug,
		})
	}

	collections := make([]homeCollectionResponse, 0, len(p.Collections))
	for _, item := range p.Collections {
		collections = append(collections, homeCollectionResponse{
			ID:          item.ID,
			Title:       safety.EscapeText(item.Title),
			Description: safety.EscapeText(item.Description),
			ImageURL:    media.PublicURL(item.ImageURL),
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
			ID:    tag.ID,
			Name:  safety.EscapeText(tag.Name),
			Slug:  tag.Slug,
			Group: tag.Group,
		})
	}

	var place *eventCardNextSessionPlaceResponse
	if item.Place != nil {
		place = toEventCardPlaceResponse(*item.Place)
	}

	var nextSession *eventCardNextSessionResponse
	if item.NextSession != nil {
		nextSession = &eventCardNextSessionResponse{
			StartAt: item.NextSession.StartAt.UTC().Format("2006-01-02T15:04:05Z"),
			Place:   *toEventCardPlaceResponse(item.NextSession.Place),
		}
		if place == nil {
			place = toEventCardPlaceResponse(item.NextSession.Place)
		}
	}

	placeName := ""
	addressLine := ""
	if place != nil {
		placeName = place.Name
		addressLine = place.AddressLine
	}

	return eventCardResponse{
		ID:               item.ID,
		Title:            safety.EscapeText(item.Title),
		ShortDescription: safety.EscapeText(item.ShortDescription),
		CoverImageURL:    media.PublicURL(item.CoverImageURL),
		IsFavorite:       item.IsFavorite,
		Tags:             tags,
		Place:            place,
		PlaceName:        placeName,
		AddressLine:      addressLine,
		NextSession:      nextSession,
	}
}

func toEventCardPlaceResponse(place placemodel.EventCardNextSessionPlace) *eventCardNextSessionPlaceResponse {
	return &eventCardNextSessionPlaceResponse{
		Name:        safety.EscapeText(place.Name),
		AddressLine: safety.EscapeText(place.AddressLine),
		ID:          place.ID,
		Latitude:    place.Latitude,
		Longitude:   place.Longitude,
		City: eventSessionPlaceCityResponse{
			ID:          place.City.ID,
			Name:        safety.EscapeText(place.City.Name),
			CountryName: safety.EscapeText(place.City.CountryName),
			Timezone:    place.City.Timezone,
		},
	}
}
