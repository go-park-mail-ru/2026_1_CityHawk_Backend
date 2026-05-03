package grpc

import (
	"context"
	"errors"
	"strings"

	"cityhawk/backend/internal/grpcconv"
	placemodel "cityhawk/backend/internal/place/model"
	placeusecase "cityhawk/backend/internal/place/usecase"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/media"
	commonv1 "cityhawk/backend/pkg/pb/common/v1"
	eventsv1 "cityhawk/backend/pkg/pb/events/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventsUsecase interface {
	HomePayload(ctx context.Context, filter placemodel.HomeFilter) placemodel.HomePayload
	ListCategories(ctx context.Context) []placemodel.HomeCategory
	ListTags(ctx context.Context) []placemodel.HomeTag
	ListCities(ctx context.Context) []placemodel.City
	ListCollections(ctx context.Context) ([]placemodel.CollectionCardView, error)
	GetCollectionByID(ctx context.Context, id string) (placemodel.CollectionDetailsView, bool, error)
	SearchSuggestions(ctx context.Context, query string, limit int) ([]placemodel.SearchSuggestion, error)
	ListEvents(ctx context.Context, filter placemodel.EventListFilter) ([]placemodel.EventCardView, int, error)
	GetByID(ctx context.Context, id, userID string) (placemodel.EventDetailsView, bool, error)
	CreateEvent(ctx context.Context, input placemodel.EventWriteInput) (string, error)
	UpdateEvent(ctx context.Context, input placemodel.EventWriteInput) (bool, error)
	DeleteEvent(ctx context.Context, id, userID string) (bool, error)
}

type PlaceLookupUsecase interface {
	Suggest(ctx context.Context, query string, limit int) ([]placemodel.PlaceSuggestion, error)
	Resolve(ctx context.Context, input placemodel.PlaceResolveInput) (placemodel.PlaceResolved, error)
}

type Server struct {
	eventsv1.UnimplementedEventsServiceServer

	events EventsUsecase
	lookup PlaceLookupUsecase
}

func NewServer(events EventsUsecase, lookup PlaceLookupUsecase) *Server {
	return &Server{events: events, lookup: lookup}
}

func (s *Server) GetHome(ctx context.Context, req *eventsv1.GetHomeRequest) (*eventsv1.HomeResponse, error) {
	return homeToProto(s.events.HomePayload(ctx, placemodel.HomeFilter{City: req.GetCity()})), nil
}

func (s *Server) ListEvents(ctx context.Context, req *eventsv1.ListEventsRequest) (*eventsv1.ListEventsResponse, error) {
	filter := placemodel.EventListFilter{
		Query:      strings.TrimSpace(req.GetQuery()),
		CategoryID: strings.TrimSpace(req.GetCategoryId()),
		TagID:      strings.TrimSpace(req.GetTagId()),
		CityID:     strings.TrimSpace(req.GetCityId()),
		DateFrom:   grpcconv.TimeFromProto(req.GetDateFrom()),
		DateTo:     grpcconv.TimeFromProto(req.GetDateTo()),
		AuthorID:   strings.TrimSpace(req.GetAuthorUserId()),
		Sort:       strings.TrimSpace(req.GetSort()),
		Limit:      int(req.GetPage().GetLimit()),
		Offset:     int(req.GetPage().GetOffset()),
		UserID:     grpcconv.UserIDFromContext(req.GetViewerContext()),
	}
	items, total, err := s.events.ListEvents(ctx, filter)
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	return eventListToProto(items, total, filter.Limit, filter.Offset), nil
}

func (s *Server) GetEvent(ctx context.Context, req *eventsv1.GetEventRequest) (*eventsv1.EventDetails, error) {
	item, ok, err := s.events.GetByID(ctx, req.GetEventId(), grpcconv.UserIDFromContext(req.GetViewerContext()))
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	if !ok {
		return nil, grpcconv.NotFound("event not found")
	}
	return eventDetailsToProto(item), nil
}

func (s *Server) CreateEvent(ctx context.Context, req *eventsv1.CreateEventRequest) (*eventsv1.EventIDResponse, error) {
	userID := grpcconv.UserIDFromContext(req.GetActor())
	if userID == "" {
		return nil, grpcconv.Error(platformerrors.ErrUnauthorized)
	}

	title := strings.TrimSpace(req.GetTitle())
	shortDescription := strings.TrimSpace(req.GetShortDescription())
	fullDescription := strings.TrimSpace(req.GetFullDescription())
	ageLimit := int(req.GetAgeLimit())
	categoryIDs := normalizeStrings(req.GetCategoryIds())
	tagIDs := normalizeStrings(req.GetTagIds())
	imageURLs := normalizeStrings(req.GetImageUrls())
	sessions := sessionsFromProto(req.GetSessions())

	id, err := s.events.CreateEvent(ctx, placemodel.EventWriteInput{
		AuthorUserID:     userID,
		Title:            &title,
		ShortDescription: &shortDescription,
		FullDescription:  &fullDescription,
		AgeLimit:         &ageLimit,
		SourceURL:        req.SourceUrl,
		CategoryIDs:      &categoryIDs,
		TagIDs:           &tagIDs,
		ImageURLs:        &imageURLs,
		Sessions:         &sessions,
	})
	if err != nil {
		return nil, eventWriteError(err)
	}
	return &eventsv1.EventIDResponse{Id: id}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *eventsv1.UpdateEventRequest) (*eventsv1.EventIDResponse, error) {
	userID := grpcconv.UserIDFromContext(req.GetActor())
	if userID == "" {
		return nil, grpcconv.Error(platformerrors.ErrUnauthorized)
	}

	input := placemodel.EventWriteInput{
		ID:             req.GetEventId(),
		AuthorUserID:   userID,
		SourceURL:      req.SourceUrl,
		ClearSourceURL: req.GetClearSourceUrl(),
	}
	if req.Title != nil {
		value := strings.TrimSpace(req.GetTitle())
		input.Title = &value
	}
	if req.ShortDescription != nil {
		value := strings.TrimSpace(req.GetShortDescription())
		input.ShortDescription = &value
	}
	if req.FullDescription != nil {
		value := strings.TrimSpace(req.GetFullDescription())
		input.FullDescription = &value
	}
	if req.AgeLimit != nil {
		value := int(req.GetAgeLimit())
		input.AgeLimit = &value
	}
	if req.GetReplaceCategoryIds() {
		values := normalizeStrings(req.GetCategoryIds())
		input.CategoryIDs = &values
	}
	if req.GetReplaceTagIds() {
		values := normalizeStrings(req.GetTagIds())
		input.TagIDs = &values
	}
	if req.GetReplaceImageUrls() {
		values := normalizeStrings(req.GetImageUrls())
		input.ImageURLs = &values
	}
	if req.GetReplaceSessions() {
		values := sessionsFromProto(req.GetSessions())
		input.Sessions = &values
	}

	ok, err := s.events.UpdateEvent(ctx, input)
	if err != nil {
		return nil, eventWriteError(err)
	}
	if !ok {
		return nil, grpcconv.NotFound("event not found")
	}
	return &eventsv1.EventIDResponse{Id: req.GetEventId()}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *eventsv1.DeleteEventRequest) (*commonv1.BoolResponse, error) {
	userID := grpcconv.UserIDFromContext(req.GetActor())
	if userID == "" {
		return nil, grpcconv.Error(platformerrors.ErrUnauthorized)
	}
	ok, err := s.events.DeleteEvent(ctx, req.GetEventId(), userID)
	if err != nil {
		return nil, eventWriteError(err)
	}
	if !ok {
		return nil, grpcconv.NotFound("event not found")
	}
	return &commonv1.BoolResponse{Ok: true}, nil
}

func (s *Server) ListCategories(ctx context.Context, _ *eventsv1.ListTaxonomyRequest) (*eventsv1.ListTaxonomyResponse, error) {
	items := s.events.ListCategories(ctx)
	response := &eventsv1.ListTaxonomyResponse{Items: make([]*eventsv1.TaxonomyItem, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, taxonomyToProto(item.ID, item.Name, item.Slug))
	}
	return response, nil
}

func (s *Server) ListTags(ctx context.Context, _ *eventsv1.ListTaxonomyRequest) (*eventsv1.ListTaxonomyResponse, error) {
	items := s.events.ListTags(ctx)
	response := &eventsv1.ListTaxonomyResponse{Items: make([]*eventsv1.TaxonomyItem, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, taxonomyToProto(item.ID, item.Name, item.Slug))
	}
	return response, nil
}

func (s *Server) ListCities(ctx context.Context, _ *eventsv1.ListCitiesRequest) (*eventsv1.ListCitiesResponse, error) {
	items := s.events.ListCities(ctx)
	response := &eventsv1.ListCitiesResponse{Items: make([]*eventsv1.City, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, cityToProto(item.ID, item.Name, item.CountryName, item.Timezone))
	}
	return response, nil
}

func (s *Server) ListCollections(ctx context.Context, req *eventsv1.ListCollectionsRequest) (*eventsv1.ListCollectionsResponse, error) {
	items, err := s.events.ListCollections(ctx)
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	response := &eventsv1.ListCollectionsResponse{
		Items: make([]*eventsv1.CollectionCard, 0, len(items)),
		Page: &commonv1.PageResponse{
			Total:  int32(len(items)),
			Limit:  req.GetPage().GetLimit(),
			Offset: req.GetPage().GetOffset(),
		},
	}
	for _, item := range items {
		response.Items = append(response.Items, collectionCardToProto(item))
	}
	return response, nil
}

func (s *Server) GetCollection(ctx context.Context, req *eventsv1.GetCollectionRequest) (*eventsv1.CollectionDetails, error) {
	item, ok, err := s.events.GetCollectionByID(ctx, req.GetCollectionId())
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	if !ok {
		return nil, grpcconv.NotFound("collection not found")
	}
	return collectionDetailsToProto(item), nil
}

func (s *Server) SearchSuggestions(ctx context.Context, req *eventsv1.SearchSuggestionsRequest) (*eventsv1.SearchSuggestionsResponse, error) {
	items, err := s.events.SearchSuggestions(ctx, strings.TrimSpace(req.GetQuery()), int(req.GetLimit()))
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	response := &eventsv1.SearchSuggestionsResponse{Items: make([]string, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, item.Label)
	}
	return response, nil
}

func (s *Server) SuggestPlaces(ctx context.Context, req *eventsv1.SuggestPlacesRequest) (*eventsv1.SuggestPlacesResponse, error) {
	if s.lookup == nil {
		return nil, status.Error(codes.Unimplemented, "place suggestions are disabled")
	}
	items, err := s.lookup.Suggest(ctx, req.GetQuery(), int(req.GetLimit()))
	if err != nil {
		return nil, grpcconv.Error(err)
	}
	response := &eventsv1.SuggestPlacesResponse{Items: make([]*eventsv1.PlaceSuggestion, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, placeSuggestionToProto(item))
	}
	return response, nil
}

func (s *Server) ResolvePlace(ctx context.Context, req *eventsv1.ResolvePlaceRequest) (*eventsv1.Place, error) {
	if s.lookup == nil {
		return nil, status.Error(codes.Unimplemented, "place resolve is disabled")
	}
	place, err := s.lookup.Resolve(ctx, placemodel.PlaceResolveInput{Token: req.GetToken()})
	if err != nil {
		if errors.Is(err, placeusecase.ErrInvalidPlaceSuggestion) {
			return nil, grpcconv.InvalidArgument("invalid place suggestion")
		}
		return nil, grpcconv.Error(err)
	}
	return resolvedPlaceToProto(place), nil
}

func eventListToProto(items []placemodel.EventCardView, total, limit, offset int) *eventsv1.ListEventsResponse {
	response := &eventsv1.ListEventsResponse{
		Items: make([]*eventsv1.EventCard, 0, len(items)),
		Page: &commonv1.PageResponse{
			Total:  int32(total),
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	}
	for _, item := range items {
		response.Items = append(response.Items, eventCardToProto(item))
	}
	return response
}

func eventCardToProto(item placemodel.EventCardView) *eventsv1.EventCard {
	var nextSession *eventsv1.EventCardNextSession
	if item.NextSession != nil {
		nextSession = &eventsv1.EventCardNextSession{
			StartAt: grpcconv.TimeToProto(item.NextSession.StartAt),
			Place: &eventsv1.EventCardNextSessionPlace{
				Name:        item.NextSession.Place.Name,
				AddressLine: item.NextSession.Place.AddressLine,
			},
		}
	}
	return &eventsv1.EventCard{
		Id:               item.ID,
		Title:            item.Title,
		ShortDescription: item.ShortDescription,
		CoverImageUrl:    media.PublicURL(item.CoverImageURL),
		Tags:             eventTaxonomyToProto(item.Tags),
		NextSession:      nextSession,
	}
}

func eventDetailsToProto(item placemodel.EventDetailsView) *eventsv1.EventDetails {
	images := make([]*eventsv1.EventImage, 0, len(item.Images))
	for _, image := range item.Images {
		images = append(images, &eventsv1.EventImage{Id: image.ID, ImageUrl: media.PublicURL(image.ImageURL)})
	}

	sessions := make([]*eventsv1.EventSession, 0, len(item.Sessions))
	for _, session := range item.Sessions {
		sessions = append(sessions, &eventsv1.EventSession{
			Id:      session.ID,
			StartAt: grpcconv.TimeToProto(session.StartAt),
			EndAt:   grpcconv.TimeToProto(session.EndAt),
			Price:   int32(session.Price),
			Place: &eventsv1.Place{
				Id:          session.Place.ID,
				Name:        session.Place.Name,
				AddressLine: session.Place.AddressLine,
				Coordinates: &eventsv1.Coordinates{Latitude: session.Place.Latitude, Longitude: session.Place.Longitude},
				City:        cityToProto(session.Place.City.ID, session.Place.City.Name, session.Place.City.CountryName, session.Place.City.Timezone),
			},
		})
	}

	return &eventsv1.EventDetails{
		Id:               item.ID,
		Title:            item.Title,
		ShortDescription: item.ShortDescription,
		FullDescription:  item.FullDescription,
		AgeLimit:         int32(item.AgeLimit),
		SourceUrl:        item.SourceURL,
		Author: &eventsv1.EventAuthor{
			Id:        item.Author.ID,
			Username:  item.Author.Username,
			AvatarUrl: item.Author.AvatarURL,
		},
		Categories: eventTaxonomyToProto(item.Categories),
		Tags:       eventTaxonomyToProto(item.Tags),
		Images:     images,
		Sessions:   sessions,
		CreatedAt:  grpcconv.TimeToProto(item.CreatedAt),
		UpdatedAt:  grpcconv.TimeToProto(item.UpdatedAt),
		IsFavorite: item.IsFavorite,
		IsOwner:    item.IsOwner,
	}
}

func homeToProto(payload placemodel.HomePayload) *eventsv1.HomeResponse {
	response := &eventsv1.HomeResponse{
		FeaturedEvents: make([]*eventsv1.HomeFeaturedEvent, 0, len(payload.FeaturedEvents)),
		Categories:     make([]*eventsv1.TaxonomyItem, 0, len(payload.Categories)),
		Collections:    make([]*eventsv1.HomeCollection, 0, len(payload.Collections)),
	}
	for _, item := range payload.FeaturedEvents {
		response.FeaturedEvents = append(response.FeaturedEvents, &eventsv1.HomeFeaturedEvent{
			Id:            item.ID,
			Title:         item.Title,
			CoverImageUrl: media.PublicURL(item.CoverImageURL),
			Tags:          homeTagsToProto(item.Tags),
			NextSession: &eventsv1.EventCardNextSession{
				StartAt: grpcconv.TimeToProto(item.NextSession.StartAt),
				Place: &eventsv1.EventCardNextSessionPlace{
					Name:        item.NextSession.Place.Name,
					AddressLine: item.NextSession.Place.AddressLine,
				},
			},
		})
	}
	for _, item := range payload.Categories {
		response.Categories = append(response.Categories, taxonomyToProto(item.ID, item.Name, item.Slug))
	}
	for _, item := range payload.Collections {
		response.Collections = append(response.Collections, &eventsv1.HomeCollection{
			Id:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			ImageUrl:    media.PublicURL(item.ImageURL),
		})
	}
	return response
}

func collectionCardToProto(item placemodel.CollectionCardView) *eventsv1.CollectionCard {
	return &eventsv1.CollectionCard{
		Id:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		ImageUrl:    media.PublicURL(item.ImageURL),
		IsPublic:    item.IsPublic,
	}
}

func collectionDetailsToProto(item placemodel.CollectionDetailsView) *eventsv1.CollectionDetails {
	events := make([]*eventsv1.EventCard, 0, len(item.Events))
	for _, event := range item.Events {
		events = append(events, eventCardToProto(event))
	}
	return &eventsv1.CollectionDetails{
		Id:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		ImageUrl:    media.PublicURL(item.ImageURL),
		IsPublic:    item.IsPublic,
		Events:      events,
	}
}

func placeSuggestionToProto(item placemodel.PlaceSuggestion) *eventsv1.PlaceSuggestion {
	return &eventsv1.PlaceSuggestion{
		Token:       item.Token,
		Label:       item.Label,
		Name:        item.Name,
		AddressLine: item.AddressLine,
		CityName:    item.CityName,
		CountryName: item.CountryName,
		Timezone:    item.Timezone,
		Coordinates: &eventsv1.Coordinates{Latitude: item.Latitude, Longitude: item.Longitude},
		Postcode:    item.Postcode,
		District:    item.District,
		Source: &eventsv1.PlaceSuggestionSource{
			Provider: "photon",
			OsmId:    item.PhotonSource.OSMID,
			OsmType:  item.PhotonSource.OSMType,
			OsmKey:   item.PhotonSource.OSMKey,
			OsmValue: item.PhotonSource.OSMValue,
		},
	}
}

func resolvedPlaceToProto(item placemodel.PlaceResolved) *eventsv1.Place {
	return &eventsv1.Place{
		Id:          item.ID,
		CityId:      item.CityID,
		Name:        item.Name,
		AddressLine: item.AddressLine,
		Coordinates: &eventsv1.Coordinates{Latitude: item.Latitude, Longitude: item.Longitude},
		City:        cityToProto(item.CityID, item.CityName, item.CountryName, item.Timezone),
	}
}

func eventTaxonomyToProto(items []placemodel.EventTaxonomyItem) []*eventsv1.TaxonomyItem {
	response := make([]*eventsv1.TaxonomyItem, 0, len(items))
	for _, item := range items {
		response = append(response, taxonomyToProto(item.ID, item.Name, item.Slug))
	}
	return response
}

func homeTagsToProto(items []placemodel.HomeTag) []*eventsv1.TaxonomyItem {
	response := make([]*eventsv1.TaxonomyItem, 0, len(items))
	for _, item := range items {
		response = append(response, taxonomyToProto(item.ID, item.Name, item.Slug))
	}
	return response
}

func taxonomyToProto(id, name, slug string) *eventsv1.TaxonomyItem {
	return &eventsv1.TaxonomyItem{Id: id, Name: name, Slug: slug}
}

func cityToProto(id, name, countryName, timezone string) *eventsv1.City {
	return &eventsv1.City{Id: id, Name: name, CountryName: countryName, Timezone: timezone}
}

func sessionsFromProto(items []*eventsv1.EventSessionInput) []placemodel.EventSessionInput {
	sessions := make([]placemodel.EventSessionInput, 0, len(items))
	for _, item := range items {
		if item == nil || item.GetStartAt() == nil || item.GetEndAt() == nil {
			continue
		}
		startAt := item.GetStartAt().AsTime().UTC()
		endAt := item.GetEndAt().AsTime().UTC()
		sessions = append(sessions, placemodel.EventSessionInput{
			PlaceID: strings.TrimSpace(item.GetPlaceId()),
			StartAt: startAt,
			EndAt:   endAt,
			Price:   int(item.GetPrice()),
		})
	}
	return sessions
}

func normalizeStrings(items []string) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func eventWriteError(err error) error {
	switch {
	case errors.Is(err, platformerrors.ErrForbidden):
		return status.Error(codes.PermissionDenied, "forbidden")
	case errors.Is(err, platformerrors.ErrEventNotFound):
		return status.Error(codes.NotFound, "event not found")
	case errors.Is(err, platformerrors.ErrInvalidReference):
		if _, message, ok := platformerrors.InvalidReferenceDetails(err); ok && message != "" {
			return status.Error(codes.InvalidArgument, message)
		}
		return status.Error(codes.InvalidArgument, "request references unknown related entity")
	default:
		return grpcconv.Error(err)
	}
}
