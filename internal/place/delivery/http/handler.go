package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	placeusecase "cityhawk/backend/internal/place/usecase"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
	"cityhawk/backend/internal/platform/media"
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
	images media.Storage
}

func NewHandler(events EventsUsecase, images media.Storage) *Handler {
	return &Handler{events: events, images: images}
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

func (h *Handler) EventsByTag(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}

		tagID := strings.TrimSpace(r.PathValue("id"))
		if tagID == "" {
			return httpx.NewHTTPError(http.StatusNotFound, "Tag not found")
		}

		filter, err := parseEventListFilter(r)
		if err != nil {
			return err
		}
		filter.TagID = tagID

		items, total, err := h.events.ListEvents(r.Context(), filter)
		if err != nil {
			return err
		}

		httpx.WriteJSON(w, http.StatusOK, toEventListResponse(items, total, filter.Limit, filter.Offset))
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

	req, imageHeaders, err := decodeCreateEventRequest(r)
	if err != nil {
		return err
	}
	uploadedURLs, err := h.saveEventImages(r, userID, imageHeaders)
	if err != nil {
		return err
	}
	req.ImageURLs = append(req.ImageURLs, uploadedURLs...)

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

	req, imageHeaders, err := decodePatchEventRequest(r)
	if err != nil {
		return err
	}
	if imageHeaders != nil {
		uploadedURLs, err := h.saveEventImages(r, userID, imageHeaders)
		if err != nil {
			return err
		}
		current := []string{}
		if req.ImageURLs != nil {
			current = append(current, *req.ImageURLs...)
		}
		current = append(current, uploadedURLs...)
		req.ImageURLs = &current
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

func (h *Handler) saveEventImages(r *http.Request, userID string, headers []*multipart.FileHeader) ([]string, error) {
	if len(headers) == 0 {
		return nil, nil
	}
	if h.images == nil {
		return nil, fmt.Errorf("event image storage is not configured")
	}

	urls := make([]string, 0, len(headers))
	for _, header := range headers {
		publicPath, err := h.images.Save(userID, header)
		if err != nil {
			switch {
			case errors.Is(err, media.ErrImageEmpty),
				errors.Is(err, media.ErrImageTooLarge),
				errors.Is(err, media.ErrUnsupportedImageType):
				return nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{
					"images": err.Error(),
				})
			default:
				return nil, err
			}
		}
		urls = append(urls, publicPath)
	}

	return urls, nil
}

func decodeCreateEventRequest(r *http.Request) (createEventRequest, []*multipart.FileHeader, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return decodeMultipartCreateEventRequest(r)
	}

	var req createEventRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return createEventRequest{}, nil, httpx.NewHTTPError(http.StatusBadRequest, "invalid json")
	}

	return req, nil, nil
}

func decodePatchEventRequest(r *http.Request) (patchEventRequest, []*multipart.FileHeader, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return decodeMultipartPatchEventRequest(r)
	}

	var req patchEventRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return patchEventRequest{}, nil, httpx.NewHTTPError(http.StatusBadRequest, "invalid json")
	}

	return req, nil, nil
}

func decodeMultipartCreateEventRequest(r *http.Request) (createEventRequest, []*multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(media.MaxImageSize * 4); err != nil {
		return createEventRequest{}, nil, httpx.NewHTTPError(http.StatusBadRequest, "invalid multipart form")
	}

	req := createEventRequest{
		Title:            multipartStringValue(r.MultipartForm, "title"),
		ShortDescription: multipartStringValue(r.MultipartForm, "shortDescription"),
		FullDescription:  multipartStringValue(r.MultipartForm, "fullDescription"),
	}

	if value, ok, err := multipartIntValue(r.MultipartForm, "ageLimit"); err != nil {
		return createEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"ageLimit": "ageLimit must be an integer"})
	} else if ok {
		req.AgeLimit = &value
	}

	if value, ok := multipartOptionalStringValue(r.MultipartForm, "sourceUrl"); ok {
		req.SourceURL = value
	}

	categoryIDs, err := multipartStringSliceValue(r.MultipartForm, "categoryIds")
	if err != nil {
		return createEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"categoryIds": "categoryIds must be a JSON array of strings"})
	}
	req.CategoryIDs = categoryIDs

	tagIDs, err := multipartStringSliceValue(r.MultipartForm, "tagIds")
	if err != nil {
		return createEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"tagIds": "tagIds must be a JSON array of strings"})
	}
	req.TagIDs = tagIDs

	imageURLs, err := multipartStringSliceValue(r.MultipartForm, "imageUrls")
	if err != nil {
		return createEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"imageUrls": "imageUrls must be a JSON array of strings"})
	}
	req.ImageURLs = imageURLs

	sessions, err := multipartSessionsValue(r.MultipartForm, "sessions")
	if err != nil {
		return createEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"sessions": "sessions must be a JSON array"})
	}
	req.Sessions = sessions

	return req, multipartFiles(r.MultipartForm, "images"), nil
}

func decodeMultipartPatchEventRequest(r *http.Request) (patchEventRequest, []*multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(media.MaxImageSize * 4); err != nil {
		return patchEventRequest{}, nil, httpx.NewHTTPError(http.StatusBadRequest, "invalid multipart form")
	}

	req := patchEventRequest{}
	if value, ok := multipartOptionalStringValue(r.MultipartForm, "title"); ok && value != nil {
		req.Title = value
	}
	if value, ok := multipartOptionalStringValue(r.MultipartForm, "shortDescription"); ok && value != nil {
		req.ShortDescription = value
	}
	if value, ok := multipartOptionalStringValue(r.MultipartForm, "fullDescription"); ok && value != nil {
		req.FullDescription = value
	}
	if value, ok, err := multipartIntValue(r.MultipartForm, "ageLimit"); err != nil {
		return patchEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"ageLimit": "ageLimit must be an integer"})
	} else if ok {
		req.AgeLimit = &value
	}
	if value, ok := multipartOptionalStringValue(r.MultipartForm, "sourceUrl"); ok {
		req.SourceURL.Set = true
		req.SourceURL.Value = value
	}
	if values, ok, err := multipartOptionalStringSliceValue(r.MultipartForm, "categoryIds"); err != nil {
		return patchEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"categoryIds": "categoryIds must be a JSON array of strings"})
	} else if ok {
		req.CategoryIDs = &values
	}
	if values, ok, err := multipartOptionalStringSliceValue(r.MultipartForm, "tagIds"); err != nil {
		return patchEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"tagIds": "tagIds must be a JSON array of strings"})
	} else if ok {
		req.TagIDs = &values
	}
	if values, ok, err := multipartOptionalStringSliceValue(r.MultipartForm, "imageUrls"); err != nil {
		return patchEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"imageUrls": "imageUrls must be a JSON array of strings"})
	} else if ok {
		req.ImageURLs = &values
	}
	if values, ok, err := multipartOptionalSessionsValue(r.MultipartForm, "sessions"); err != nil {
		return patchEventRequest{}, nil, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"sessions": "sessions must be a JSON array"})
	} else if ok {
		req.Sessions = &values
	}

	return req, multipartFiles(r.MultipartForm, "images"), nil
}

func multipartStringValue(form *multipart.Form, key string) string {
	if form == nil || form.Value == nil {
		return ""
	}
	values := form.Value[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func multipartOptionalStringValue(form *multipart.Form, key string) (*string, bool) {
	if form == nil || form.Value == nil {
		return nil, false
	}
	values, ok := form.Value[key]
	if !ok || len(values) == 0 {
		return nil, false
	}
	value := values[0]
	return &value, true
}

func multipartIntValue(form *multipart.Form, key string) (int, bool, error) {
	raw, ok := multipartOptionalStringValue(form, key)
	if !ok || raw == nil {
		return 0, false, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(*raw))
	if err != nil {
		return 0, false, err
	}
	return value, true, nil
}

func multipartStringSliceValue(form *multipart.Form, key string) ([]string, error) {
	raw := strings.TrimSpace(multipartStringValue(form, key))
	if raw == "" {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func multipartOptionalStringSliceValue(form *multipart.Form, key string) ([]string, bool, error) {
	raw, ok := multipartOptionalStringValue(form, key)
	if !ok || raw == nil {
		return nil, false, nil
	}
	if strings.TrimSpace(*raw) == "" {
		return []string{}, true, nil
	}
	var values []string
	if err := json.Unmarshal([]byte(*raw), &values); err != nil {
		return nil, false, err
	}
	return values, true, nil
}

func multipartSessionsValue(form *multipart.Form, key string) ([]eventSessionRequest, error) {
	raw := strings.TrimSpace(multipartStringValue(form, key))
	if raw == "" {
		return nil, nil
	}
	var values []eventSessionRequest
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func multipartOptionalSessionsValue(form *multipart.Form, key string) ([]eventSessionRequest, bool, error) {
	raw, ok := multipartOptionalStringValue(form, key)
	if !ok || raw == nil {
		return nil, false, nil
	}
	if strings.TrimSpace(*raw) == "" {
		return []eventSessionRequest{}, true, nil
	}
	var values []eventSessionRequest
	if err := json.Unmarshal([]byte(*raw), &values); err != nil {
		return nil, false, err
	}
	return values, true, nil
}

func multipartFiles(form *multipart.Form, key string) []*multipart.FileHeader {
	if form == nil || form.File == nil {
		return nil
	}
	return form.File[key]
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
		if !media.IsFileReference(imageURL) {
			details["imageUrls["+strconv.Itoa(i)+"]"] = "imageUrl must be an http(s) URL or /uploads path"
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
			if !media.IsFileReference(imageURL) {
				details["imageUrls["+strconv.Itoa(i)+"]"] = "imageUrl must be an http(s) URL or /uploads path"
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
		if placeID == "" {
			details["sessions["+strconv.Itoa(i)+"].placeId"] = "placeId is required"
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
			PlaceID: placeID,
			StartAt: startAt.UTC(),
			EndAt:   endAt.UTC(),
			Price:   item.Price,
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
	return media.IsHTTPURL(value)
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
