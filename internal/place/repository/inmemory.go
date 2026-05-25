package repository

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	placemodel "cityhawk/backend/internal/place/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
)

type InMemoryRepository struct {
	mu          sync.RWMutex
	byID        map[string]placemodel.EventDetailsView
	order       []string
	nextEventID int
	nextImageID int
	nextSessID  int
}

func NewInMemoryRepository(events []placemodel.EventDetailsView) *InMemoryRepository {
	r := &InMemoryRepository{
		byID:        make(map[string]placemodel.EventDetailsView),
		nextEventID: 100,
		nextImageID: 100,
		nextSessID:  100,
	}
	for _, event := range events {
		r.byID[event.ID] = event
		r.order = append(r.order, event.ID)
	}
	return r
}

func (r *InMemoryRepository) HomePayload(_ context.Context, filter placemodel.HomeFilter) placemodel.HomePayload {
	r.mu.RLock()
	defer r.mu.RUnlock()

	featured := make([]placemodel.HomeFeaturedEvent, 0, len(r.order))
	categoriesSet := map[string]placemodel.HomeCategory{}
	filteredEvents := make([]placemodel.EventDetailsView, 0, len(r.order))
	for _, id := range r.order {
		event := r.byID[id]
		if !matchesHomeCityFilter(event, filter.City) {
			continue
		}
		filteredEvents = append(filteredEvents, event)
		if len(featured) < 8 {
			nextSession := placemodel.HomeNextSession{}
			if len(event.Sessions) > 0 {
				nextSession = placemodel.HomeNextSession{
					StartAt: event.Sessions[0].StartAt,
					Place: placemodel.HomeNextSessionPlace{
						Name:        event.Sessions[0].Place.Name,
						AddressLine: event.Sessions[0].Place.AddressLine,
					},
				}
			}
			featured = append(featured, placemodel.HomeFeaturedEvent{
				ID:            event.ID,
				Title:         event.Title,
				CoverImageURL: firstImageURL(event),
				Tags:          toHomeTags(event.Tags),
				NextSession:   nextSession,
			})
		}
		for _, category := range event.Categories {
			categoriesSet[category.ID] = placemodel.HomeCategory{ID: category.ID, Name: category.Name, Slug: category.Slug}
		}
	}

	categories := make([]placemodel.HomeCategory, 0, len(categoriesSet))
	for _, category := range categoriesSet {
		categories = append(categories, category)
	}

	return placemodel.HomePayload{
		FeaturedEvents: featured,
		Categories:     categories,
		Collections:    homeCollectionsForEvents(filteredEvents),
	}
}

func (r *InMemoryRepository) ListCategories(_ context.Context) []placemodel.HomeCategory {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categoriesSet := map[string]placemodel.HomeCategory{}
	for _, id := range r.order {
		event := r.byID[id]
		for _, category := range event.Categories {
			categoriesSet[category.ID] = placemodel.HomeCategory{
				ID:   category.ID,
				Name: category.Name,
				Slug: category.Slug,
			}
		}
	}

	items := make([]placemodel.HomeCategory, 0, len(categoriesSet))
	for _, category := range categoriesSet {
		items = append(items, category)
	}
	return items
}

func (r *InMemoryRepository) ListTags(_ context.Context) []placemodel.HomeTag {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tagsSet := map[string]placemodel.HomeTag{}
	for _, id := range r.order {
		event := r.byID[id]
		for _, tag := range event.Tags {
			tagsSet[tag.ID] = placemodel.HomeTag{
				ID:   tag.ID,
				Name: tag.Name,
				Slug: tag.Slug,
			}
		}
	}

	items := make([]placemodel.HomeTag, 0, len(tagsSet))
	for _, tag := range tagsSet {
		items = append(items, tag)
	}
	return items
}

func (r *InMemoryRepository) ListCities(_ context.Context) []placemodel.City {
	r.mu.RLock()
	defer r.mu.RUnlock()

	citiesSet := map[string]placemodel.City{}
	for _, id := range r.order {
		event := r.byID[id]
		for _, session := range event.Sessions {
			city := session.Place.City
			if city.ID == "" {
				continue
			}
			citiesSet[city.ID] = placemodel.City{
				ID:          city.ID,
				Name:        city.Name,
				CountryName: city.CountryName,
				Timezone:    city.Timezone,
			}
		}
	}

	items := make([]placemodel.City, 0, len(citiesSet))
	for _, city := range citiesSet {
		items = append(items, city)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Name == items[j].Name {
			return items[i].ID < items[j].ID
		}
		return items[i].Name < items[j].Name
	})
	return items
}

func (r *InMemoryRepository) ListCollections(_ context.Context) ([]placemodel.CollectionCardView, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return []placemodel.CollectionCardView{
		{
			ID:          "weekend-picks",
			Title:       "Weekend Picks",
			Description: "Best events for weekend",
			ImageURL:    "https://example.com/collection.jpg",
			IsPublic:    true,
		},
	}, nil
}

func (r *InMemoryRepository) GetCollectionByID(_ context.Context, id string) (placemodel.CollectionDetailsView, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if id != "weekend-picks" {
		return placemodel.CollectionDetailsView{}, false, nil
	}

	events := make([]placemodel.EventCardView, 0, min(2, len(r.order)))
	for i, eventID := range r.order {
		if i >= 2 {
			break
		}
		events = append(events, toCard(r.byID[eventID]))
	}

	return placemodel.CollectionDetailsView{
		ID:          "weekend-picks",
		Title:       "Weekend Picks",
		Description: "Best events for weekend",
		ImageURL:    "https://example.com/collection.jpg",
		IsPublic:    true,
		Events:      events,
	}, true, nil
}

func (r *InMemoryRepository) SearchSuggestions(_ context.Context, query string, limit int) ([]placemodel.SearchSuggestion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(strings.TrimSpace(query))
	if limit <= 0 {
		limit = 5
	}

	seen := map[string]struct{}{}
	items := make([]placemodel.SearchSuggestion, 0)

	add := func(id, kind, title string) {
		key := kind + ":" + strings.ToLower(strings.TrimSpace(id))
		labelKey := strings.ToLower(strings.TrimSpace(title))
		if key == "" {
			return
		}
		if !strings.Contains(labelKey, query) && !strings.HasPrefix(labelKey, query) {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		items = append(items, placemodel.SearchSuggestion{ID: id, Type: kind, Title: title, Label: title})
	}

	for _, id := range r.order {
		event := r.byID[id]
		add(event.ID, "event", event.Title)
		for _, category := range event.Categories {
			add(category.ID, "category", category.Name)
		}
		for _, tag := range event.Tags {
			add(tag.ID, "tag", tag.Name)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Title) < strings.ToLower(items[j].Title)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (r *InMemoryRepository) ListEvents(_ context.Context, filter placemodel.EventListFilter) ([]placemodel.EventCardView, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]placemodel.EventCardView, 0)
	for _, id := range r.order {
		event := r.byID[id]
		if !matchesFilter(event, filter) {
			continue
		}
		items = append(items, toCard(event))
	}

	total := len(items)
	start := filter.Offset
	if start > len(items) {
		start = len(items)
	}
	end := start + filter.Limit
	if filter.Limit <= 0 || end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func (r *InMemoryRepository) GetByID(_ context.Context, id, userID string) (placemodel.EventDetailsView, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	event, ok := r.byID[id]
	if !ok {
		return placemodel.EventDetailsView{}, false, nil
	}
	event.IsOwner = userID != "" && event.Author.ID == userID
	event.IsFavorite = false
	return event, true, nil
}

func (r *InMemoryRepository) CreateEvent(_ context.Context, input placemodel.EventWriteInput) (string, error) {
	now := time.Now().UTC()
	categoryItems := buildTaxonomyItems(derefStringSlice(input.CategoryIDs))
	tagItems := buildTaxonomyItems(derefStringSlice(input.TagIDs))
	imageURLs := derefStringSlice(input.ImageURLs)
	sessionInputs := derefSessions(input.Sessions)

	r.mu.Lock()
	eventID := generateID("event", &r.nextEventID)
	imageIDs := reserveIDs("image", &r.nextImageID, len(imageURLs))
	sessionIDs := reserveIDs("session", &r.nextSessID, len(sessionInputs))
	r.mu.Unlock()

	event := placemodel.EventDetailsView{
		ID:               eventID,
		Title:            derefStringLocal(input.Title),
		ShortDescription: derefStringLocal(input.ShortDescription),
		FullDescription:  derefStringLocal(input.FullDescription),
		AgeLimit:         derefIntLocal(input.AgeLimit),
		SourceURL:        input.SourceURL,
		Author: placemodel.EventAuthorView{
			ID:       input.AuthorUserID,
			Username: "author",
		},
		Categories: categoryItems,
		Tags:       tagItems,
		Images:     buildImageItemsWithIDs(imageURLs, imageIDs),
		Place:      buildPlaceItem(input.PlaceID),
		Sessions:   buildSessionItemsWithIDs(sessionInputs, sessionIDs),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	r.mu.Lock()
	r.byID[eventID] = event
	r.order = append([]string{eventID}, r.order...)
	r.mu.Unlock()
	return eventID, nil
}

func (r *InMemoryRepository) UpdateEvent(_ context.Context, input placemodel.EventWriteInput) (bool, error) {
	r.mu.RLock()
	event, ok := r.byID[input.ID]
	if !ok {
		r.mu.RUnlock()
		return false, nil
	}
	if event.Author.ID != input.AuthorUserID {
		r.mu.RUnlock()
		return false, platformerrors.ErrForbidden
	}
	r.mu.RUnlock()

	var categoryItems []placemodel.EventTaxonomyItem
	if input.CategoryIDs != nil {
		categoryItems = buildTaxonomyItems(*input.CategoryIDs)
	}

	var tagItems []placemodel.EventTaxonomyItem
	if input.TagIDs != nil {
		tagItems = buildTaxonomyItems(*input.TagIDs)
	}

	var imageIDs []string
	if input.ImageURLs != nil {
		r.mu.Lock()
		imageIDs = reserveIDs("image", &r.nextImageID, len(*input.ImageURLs))
		r.mu.Unlock()
	}

	var sessionIDs []string
	if input.Sessions != nil {
		r.mu.Lock()
		sessionIDs = reserveIDs("session", &r.nextSessID, len(*input.Sessions))
		r.mu.Unlock()
	}

	var imageItems []placemodel.EventImageView
	if input.ImageURLs != nil {
		imageItems = buildImageItemsWithIDs(*input.ImageURLs, imageIDs)
	}

	var sessionItems []placemodel.EventSessionView
	if input.Sessions != nil {
		sessionItems = buildSessionItemsWithIDs(*input.Sessions, sessionIDs)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	event, ok = r.byID[input.ID]
	if !ok {
		return false, nil
	}
	if event.Author.ID != input.AuthorUserID {
		return false, platformerrors.ErrForbidden
	}

	if input.Title != nil {
		event.Title = *input.Title
	}
	if input.ShortDescription != nil {
		event.ShortDescription = *input.ShortDescription
	}
	if input.FullDescription != nil {
		event.FullDescription = *input.FullDescription
	}
	if input.AgeLimit != nil {
		event.AgeLimit = *input.AgeLimit
	}
	if input.SourceURL != nil {
		event.SourceURL = input.SourceURL
	}
	if input.ClearSourceURL {
		event.SourceURL = nil
	}
	if input.CategoryIDs != nil {
		event.Categories = categoryItems
	}
	if input.TagIDs != nil {
		event.Tags = tagItems
	}
	if input.ImageURLs != nil {
		event.Images = imageItems
	}
	if input.PlaceID != nil {
		event.Place = buildPlaceItem(input.PlaceID)
	}
	if input.ClearPlace {
		event.Place = nil
	}
	if input.Sessions != nil {
		event.Sessions = sessionItems
	}
	event.UpdatedAt = time.Now().UTC()
	r.byID[input.ID] = event
	return true, nil
}

func (r *InMemoryRepository) DeleteEvent(_ context.Context, id, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	event, ok := r.byID[id]
	if !ok {
		return false, nil
	}
	if event.Author.ID != userID {
		return false, platformerrors.ErrForbidden
	}
	delete(r.byID, id)
	nextOrder := make([]string, 0, len(r.order))
	for _, itemID := range r.order {
		if itemID != id {
			nextOrder = append(nextOrder, itemID)
		}
	}
	r.order = nextOrder
	return true, nil
}

func firstImageURL(event placemodel.EventDetailsView) string {
	if len(event.Images) == 0 {
		return ""
	}
	return event.Images[0].ImageURL
}

func toHomeTags(tags []placemodel.EventTaxonomyItem) []placemodel.HomeTag {
	items := make([]placemodel.HomeTag, 0, len(tags))
	for _, tag := range tags {
		items = append(items, placemodel.HomeTag{ID: tag.ID, Name: tag.Name, Slug: tag.Slug})
	}
	return items
}

func homeCollectionsForEvents(events []placemodel.EventDetailsView) []placemodel.HomeCollection {
	if len(events) == 0 {
		return nil
	}
	return []placemodel.HomeCollection{
		{ID: "weekend-picks", Title: "Weekend Picks", Description: "Best events for weekend", ImageURL: "https://example.com/collection.jpg"},
	}
}

func toCard(event placemodel.EventDetailsView) placemodel.EventCardView {
	var nextSession *placemodel.EventCardNextSession
	if len(event.Sessions) > 0 {
		first := event.Sessions[0]
		nextSession = &placemodel.EventCardNextSession{
			StartAt: first.StartAt,
			Place: placemodel.EventCardNextSessionPlace{
				Name:        first.Place.Name,
				AddressLine: first.Place.AddressLine,
			},
		}
	}
	return placemodel.EventCardView{
		ID:               event.ID,
		Title:            event.Title,
		ShortDescription: event.ShortDescription,
		CoverImageURL:    firstImageURL(event),
		Tags:             event.Tags,
		NextSession:      nextSession,
		IsFavorite:       event.IsFavorite,
	}
}

func matchesFilter(event placemodel.EventDetailsView, filter placemodel.EventListFilter) bool {
	if filter.Query != "" {
		query := strings.ToLower(filter.Query)
		if !strings.Contains(strings.ToLower(event.Title), query) &&
			!strings.Contains(strings.ToLower(event.ShortDescription), query) &&
			!strings.Contains(strings.ToLower(event.FullDescription), query) {
			return false
		}
	}
	if filter.CategoryID != "" && !hasTaxonomyID(event.Categories, filter.CategoryID) {
		return false
	}
	if filter.TagID != "" && !hasTaxonomyID(event.Tags, filter.TagID) {
		return false
	}
	if filter.AuthorID != "" && event.Author.ID != filter.AuthorID {
		return false
	}
	return true
}

func matchesHomeCityFilter(event placemodel.EventDetailsView, city string) bool {
	city = strings.ToLower(strings.TrimSpace(city))
	if city == "" {
		return true
	}
	for _, session := range event.Sessions {
		sessionCity := session.Place.City
		if strings.ToLower(sessionCity.ID) == city || strings.ToLower(sessionCity.Name) == city {
			return true
		}
	}
	return false
}

func hasTaxonomyID(items []placemodel.EventTaxonomyItem, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func buildTaxonomyItems(ids []string) []placemodel.EventTaxonomyItem {
	items := make([]placemodel.EventTaxonomyItem, 0, len(ids))
	for _, id := range ids {
		items = append(items, placemodel.EventTaxonomyItem{ID: id, Name: id, Slug: id})
	}
	return items
}

func buildImageItems(urls []string, next *int) []placemodel.EventImageView {
	return buildImageItemsWithIDs(urls, reserveIDs("image", next, len(urls)))
}

func buildImageItemsWithIDs(urls, ids []string) []placemodel.EventImageView {
	items := make([]placemodel.EventImageView, 0, len(urls))
	for i, imageURL := range urls {
		items = append(items, placemodel.EventImageView{ID: ids[i], ImageURL: imageURL})
	}
	return items
}

func buildSessionItems(sessions []placemodel.EventSessionInput, next *int) []placemodel.EventSessionView {
	return buildSessionItemsWithIDs(sessions, reserveIDs("session", next, len(sessions)))
}

func buildPlaceItem(placeID *string) *placemodel.EventSessionPlaceView {
	if placeID == nil || strings.TrimSpace(*placeID) == "" {
		return nil
	}
	return &placemodel.EventSessionPlaceView{ID: strings.TrimSpace(*placeID)}
}

func buildSessionItemsWithIDs(sessions []placemodel.EventSessionInput, ids []string) []placemodel.EventSessionView {
	items := make([]placemodel.EventSessionView, 0, len(sessions))
	for i, session := range sessions {
		items = append(items, placemodel.EventSessionView{
			ID:      ids[i],
			StartAt: session.StartAt,
			EndAt:   session.EndAt,
			Price:   session.Price,
			Place: placemodel.EventSessionPlaceView{
				ID:          session.PlaceID,
				Name:        session.PlaceID,
				AddressLine: "Address for " + session.PlaceID,
				City: placemodel.EventSessionPlaceCityView{
					ID:          "city-1",
					Name:        "Moscow",
					CountryName: "Russia",
					Timezone:    "Europe/Moscow",
				},
			},
		})
	}
	return items
}

func derefStringSlice(value *[]string) []string {
	if value == nil {
		return nil
	}
	return *value
}

func derefSessions(value *[]placemodel.EventSessionInput) []placemodel.EventSessionInput {
	if value == nil {
		return nil
	}
	return *value
}

func generateID(prefix string, next *int) string {
	if next == nil {
		panic("generateID: next must not be nil")
	}
	*next = *next + 1
	return prefix + "-" + strconv.Itoa(*next)
}

func reserveIDs(prefix string, next *int, count int) []string {
	ids := make([]string, 0, count)
	for range count {
		ids = append(ids, generateID(prefix, next))
	}
	return ids
}

func derefStringLocal(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefIntLocal(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
