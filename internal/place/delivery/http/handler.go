package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	placeusecase "cityhawk/backend/internal/place/usecase"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
)

type EventsUsecase interface {
	HomePayload(ctx context.Context) placemodel.HomePayload
	ListCategories(ctx context.Context) []placemodel.HomeCategory
	ListTags(ctx context.Context) []placemodel.HomeTag
	ListCollections(ctx context.Context) ([]placemodel.CollectionCardView, error)
	GetCollectionByID(ctx context.Context, id string) (placemodel.CollectionDetailsView, bool, error)
	SearchSuggestions(ctx context.Context, query string, limit int) ([]placemodel.SearchSuggestion, error)
	ListEvents(ctx context.Context, filter placemodel.EventListFilter) ([]placemodel.EventCardView, int, error)
	GetByID(ctx context.Context, id, userID string) (placemodel.EventDetailsView, bool, error)
	CreateEvent(ctx context.Context, input placemodel.EventWriteInput) (string, error)
	UpdateEvent(ctx context.Context, input placemodel.EventWriteInput) (bool, error)
	DeleteEvent(ctx context.Context, id, userID string) (bool, error)
}

type Handler struct {
	events EventsUsecase
}

func NewHandler(events EventsUsecase) *Handler {
	return &Handler{events: events}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		httpx.WriteJSON(w, http.StatusOK, toHomePayloadResponse(h.events.HomePayload(r.Context())))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		httpx.WriteJSON(w, http.StatusOK, toCategoriesResponse(h.events.ListCategories(r.Context())))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Tags(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		httpx.WriteJSON(w, http.StatusOK, toTagsResponse(h.events.ListTags(r.Context())))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Collections(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		items, err := h.events.ListCollections(r.Context())
		if err != nil {
			return err
		}
		httpx.WriteJSON(w, http.StatusOK, toCollectionsResponse(items))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		query, limit, err := parseSearchSuggestionsRequest(r)
		if err != nil {
			return err
		}

		items, err := h.events.SearchSuggestions(r.Context(), query, limit)
		if err != nil {
			return err
		}

		httpx.WriteJSON(w, http.StatusOK, toSearchSuggestionsResponse(items))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) CollectionByID(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		collectionID := strings.TrimPrefix(r.URL.Path, "/api/collections/")
		if collectionID == "" || strings.Contains(collectionID, "/") {
			return httpx.NewHTTPError(http.StatusNotFound, "Collection not found")
		}

		item, ok, err := h.events.GetCollectionByID(r.Context(), collectionID)
		if err != nil {
			return err
		}
		if !ok {
			return httpx.NewHTTPError(http.StatusNotFound, "Collection not found")
		}

		httpx.WriteJSON(w, http.StatusOK, toCollectionDetailsResponse(item))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return h.handleList(w, r)
		case http.MethodPost:
			return h.handleCreate(w, r)
		default:
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
	}).ServeHTTP(w, r)
}

func (h *Handler) EventByID(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		switch r.Method {
		case http.MethodGet:
			return h.handleDetails(w, r)
		case http.MethodPatch:
			return h.handlePatch(w, r)
		case http.MethodDelete:
			return h.handleDelete(w, r)
		default:
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
	}).ServeHTTP(w, r)
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) error {
	filter, err := parseEventListFilter(r)
	if err != nil {
		return err
	}

	items, total, err := h.events.ListEvents(r.Context(), filter)
	if err != nil {
		return err
	}

	httpx.WriteJSON(w, http.StatusOK, toEventListResponse(items, total, filter.Limit, filter.Offset))
	return nil
}

func (h *Handler) handleDetails(w http.ResponseWriter, r *http.Request) error {
	eventID := strings.TrimPrefix(r.URL.Path, "/api/events/")
	if eventID == "" || strings.Contains(eventID, "/") {
		return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	userID, _ := r.Context().Value(httpx.UserIDContextKey).(string)
	item, ok, err := h.events.GetByID(r.Context(), eventID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	httpx.WriteJSON(w, http.StatusOK, toEventDetailsResponse(item))
	return nil
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) error {
	userID, ok := r.Context().Value(httpx.UserIDContextKey).(string)
	if !ok || userID == "" {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	var req createEventRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse("invalid json", nil))
		return nil
	}

	input, err := validateCreateEventRequest(req, userID)
	if err != nil {
		return err
	}

	eventID, err := h.events.CreateEvent(r.Context(), input)
	if err != nil {
		return mapEventWriteError(err)
	}

	httpx.WriteJSON(w, http.StatusCreated, eventIDResponse{ID: eventID})
	return nil
}

func (h *Handler) handlePatch(w http.ResponseWriter, r *http.Request) error {
	userID, ok := r.Context().Value(httpx.UserIDContextKey).(string)
	if !ok || userID == "" {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	eventID := strings.TrimPrefix(r.URL.Path, "/api/events/")
	if eventID == "" || strings.Contains(eventID, "/") {
		return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	var req patchEventRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, httpx.NewErrorResponse("invalid json", nil))
		return nil
	}

	input, err := validatePatchEventRequest(req, eventID, userID)
	if err != nil {
		return err
	}

	ok, err = h.events.UpdateEvent(r.Context(), input)
	if err != nil {
		return mapEventWriteError(err)
	}
	if !ok {
		return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	item, found, err := h.events.GetByID(r.Context(), eventID, userID)
	if err != nil {
		return err
	}
	if !found {
		return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	httpx.WriteJSON(w, http.StatusOK, toEventDetailsResponse(item))
	return nil
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) error {
	userID, ok := r.Context().Value(httpx.UserIDContextKey).(string)
	if !ok || userID == "" {
		return httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	eventID := strings.TrimPrefix(r.URL.Path, "/api/events/")
	if eventID == "" || strings.Contains(eventID, "/") {
		return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	ok, err := h.events.DeleteEvent(r.Context(), eventID, userID)
	if err != nil {
		if errors.Is(err, platformerrors.ErrForbidden) {
			return httpx.NewHTTPError(http.StatusForbidden, "Forbidden")
		}
		return err
	}
	if !ok {
		return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"ok": true})
	return nil
}

func parseEventListFilter(r *http.Request) (placemodel.EventListFilter, error) {
	limit := placeusecase.DefaultEventsLimit
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return placemodel.EventListFilter{}, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"limit": "limit must be a positive integer"})
		}
		limit = value
	}

	offset := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return placemodel.EventListFilter{}, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"offset": "offset must be a non-negative integer"})
		}
		offset = value
	}

	dateFrom, err := parseOptionalDateQuery(r.URL.Query(), "dateFrom", false)
	if err != nil {
		return placemodel.EventListFilter{}, err
	}
	dateTo, err := parseOptionalDateQuery(r.URL.Query(), "dateTo", true)
	if err != nil {
		return placemodel.EventListFilter{}, err
	}

	sortValue := strings.TrimSpace(r.URL.Query().Get("sort"))
	if sortValue != "" && sortValue != "dateAsc" && sortValue != "dateDesc" && sortValue != "titleAsc" {
		return placemodel.EventListFilter{}, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"sort": "sort must be one of dateAsc, dateDesc, titleAsc"})
	}

	userID, _ := r.Context().Value(httpx.UserIDContextKey).(string)
	return placemodel.EventListFilter{
		Query:      strings.TrimSpace(r.URL.Query().Get("query")),
		CategoryID: strings.TrimSpace(r.URL.Query().Get("categoryId")),
		TagID:      strings.TrimSpace(r.URL.Query().Get("tagId")),
		CityID:     strings.TrimSpace(r.URL.Query().Get("cityId")),
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		AuthorID:   strings.TrimSpace(r.URL.Query().Get("authorId")),
		Sort:       sortValue,
		Limit:      limit,
		Offset:     offset,
		UserID:     userID,
	}, nil
}

func parseSearchSuggestionsRequest(r *http.Request) (string, int, error) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if len([]rune(query)) < 2 {
		return "", 0, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{
			"query": "query must be at least 2 characters",
		})
	}

	limit := 5
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 5 || value > 10 {
			return "", 0, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{
				"limit": "limit must be an integer between 5 and 10",
			})
		}
		limit = value
	}

	return query, limit, nil
}

func validateCreateEventRequest(req createEventRequest, userID string) (placemodel.EventWriteInput, error) {
	details := map[string]string{}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		details["title"] = "title is required"
	}

	shortDescription := strings.TrimSpace(req.ShortDescription)
	if shortDescription == "" {
		details["shortDescription"] = "shortDescription is required"
	}

	fullDescription := strings.TrimSpace(req.FullDescription)
	if fullDescription == "" {
		details["fullDescription"] = "fullDescription is required"
	}

	categoryIDs := normalizeStringSlice(req.CategoryIDs)
	if len(categoryIDs) == 0 {
		details["categoryIds"] = "categoryIds is required"
	}

	sessions, sessionErrs := validateSessions(req.Sessions)
	for key, value := range sessionErrs {
		details[key] = value
	}
	if len(req.Sessions) == 0 {
		details["sessions"] = "sessions is required"
	}

	var ageLimit int
	if req.AgeLimit != nil {
		ageLimit = *req.AgeLimit
		if ageLimit < 0 || ageLimit > 21 {
			details["ageLimit"] = "ageLimit must be between 0 and 21"
		}
	}

	if req.SourceURL != nil {
		value := strings.TrimSpace(*req.SourceURL)
		if value != "" && !isHTTPURL(value) {
			details["sourceUrl"] = "sourceUrl must start with http:// or https://"
		}
		if value == "" {
			req.SourceURL = nil
		} else {
			req.SourceURL = &value
		}
	}

	tagIDs := normalizeStringSlice(req.TagIDs)
	imageURLs := normalizeStringSlice(req.ImageURLs)
	for i, imageURL := range imageURLs {
		if !isHTTPURL(imageURL) {
			details["imageUrls["+strconv.Itoa(i)+"]"] = "imageUrl must start with http:// or https://"
		}
	}

	if len(details) > 0 {
		return placemodel.EventWriteInput{}, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", details)
	}

	return placemodel.EventWriteInput{
		AuthorUserID:     userID,
		Title:            &title,
		ShortDescription: &shortDescription,
		FullDescription:  &fullDescription,
		AgeLimit:         &ageLimit,
		SourceURL:        req.SourceURL,
		CategoryIDs:      &categoryIDs,
		TagIDs:           &tagIDs,
		ImageURLs:        &imageURLs,
		Sessions:         &sessions,
	}, nil
}

func validatePatchEventRequest(req patchEventRequest, eventID, userID string) (placemodel.EventWriteInput, error) {
	details := map[string]string{}
	input := placemodel.EventWriteInput{
		ID:           eventID,
		AuthorUserID: userID,
	}

	if req.Title != nil {
		value := strings.TrimSpace(*req.Title)
		if value == "" {
			details["title"] = "title must not be empty"
		} else {
			input.Title = &value
		}
	}

	if req.ShortDescription != nil {
		value := strings.TrimSpace(*req.ShortDescription)
		if value == "" {
			details["shortDescription"] = "shortDescription must not be empty"
		} else {
			input.ShortDescription = &value
		}
	}

	if req.FullDescription != nil {
		value := strings.TrimSpace(*req.FullDescription)
		if value == "" {
			details["fullDescription"] = "fullDescription must not be empty"
		} else {
			input.FullDescription = &value
		}
	}

	if req.AgeLimit != nil {
		if *req.AgeLimit < 0 || *req.AgeLimit > 21 {
			details["ageLimit"] = "ageLimit must be between 0 and 21"
		} else {
			input.AgeLimit = req.AgeLimit
		}
	}

	if req.SourceURL.Set {
		if req.SourceURL.Value == nil {
			input.ClearSourceURL = true
		} else {
			value := strings.TrimSpace(*req.SourceURL.Value)
			if value == "" {
				input.ClearSourceURL = true
			} else if !isHTTPURL(value) {
				details["sourceUrl"] = "sourceUrl must start with http:// or https://"
			} else {
				input.SourceURL = &value
			}
		}
	}

	if req.CategoryIDs != nil {
		value := normalizeStringSlice(*req.CategoryIDs)
		input.CategoryIDs = &value
	}
	if req.TagIDs != nil {
		value := normalizeStringSlice(*req.TagIDs)
		input.TagIDs = &value
	}
	if req.ImageURLs != nil {
		value := normalizeStringSlice(*req.ImageURLs)
		for i, imageURL := range value {
			if !isHTTPURL(imageURL) {
				details["imageUrls["+strconv.Itoa(i)+"]"] = "imageUrl must start with http:// or https://"
			}
		}
		input.ImageURLs = &value
	}
	if req.Sessions != nil {
		value, sessionErrs := validateSessions(*req.Sessions)
		for key, message := range sessionErrs {
			details[key] = message
		}
		input.Sessions = &value
	}

	if len(details) > 0 {
		return placemodel.EventWriteInput{}, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", details)
	}
	return input, nil
}

func validateSessions(items []eventSessionRequest) ([]placemodel.EventSessionInput, map[string]string) {
	sessions := make([]placemodel.EventSessionInput, 0, len(items))
	details := map[string]string{}

	for i, item := range items {
		placeID := strings.TrimSpace(item.PlaceID)
		placeName := strings.TrimSpace(item.PlaceName)
		if placeID == "" && placeName == "" {
			details["sessions["+strconv.Itoa(i)+"].placeId"] = "placeId or placeName is required"
			continue
		}

		startAt, err := time.Parse(time.RFC3339, strings.TrimSpace(item.StartAt))
		if err != nil {
			details["sessions["+strconv.Itoa(i)+"].startAt"] = "startAt must be RFC3339 datetime"
			continue
		}
		endAt, err := time.Parse(time.RFC3339, strings.TrimSpace(item.EndAt))
		if err != nil {
			details["sessions["+strconv.Itoa(i)+"].endAt"] = "endAt must be RFC3339 datetime"
			continue
		}
		if !startAt.Before(endAt) {
			details["sessions["+strconv.Itoa(i)+"].endAt"] = "endAt must be after startAt"
			continue
		}
		if item.Price < 0 {
			details["sessions["+strconv.Itoa(i)+"].price"] = "price must be non-negative"
			continue
		}

		sessions = append(sessions, placemodel.EventSessionInput{
			PlaceID:   placeID,
			PlaceName: placeName,
			StartAt:   startAt.UTC(),
			EndAt:     endAt.UTC(),
			Price:     item.Price,
		})
	}

	return sessions, details
}

func parseOptionalDateQuery(values url.Values, key string, endOfDay bool) (*time.Time, error) {
	raw := strings.TrimSpace(values.Get(key))
	if raw == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{
				key: key + " must be YYYY-MM-DD or RFC3339 datetime",
			})
		}
	}

	value := parsed.UTC()
	if endOfDay && len(raw) == len("2006-01-02") {
		value = value.Add(24*time.Hour - time.Nanosecond)
	}
	return &value, nil
}

func normalizeStringSlice(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func isHTTPURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func mapEventWriteError(err error) error {
	switch {
	case errors.Is(err, platformerrors.ErrForbidden):
		return httpx.NewHTTPError(http.StatusForbidden, "Forbidden")
	case errors.Is(err, platformerrors.ErrEventNotFound):
		return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
	case errors.Is(err, platformerrors.ErrInvalidReference):
		return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{
			"value": "request references unknown related entity",
		})
	default:
		return err
	}
}
