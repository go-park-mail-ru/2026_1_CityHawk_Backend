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

func (r *InMemoryRepository) HomePayload(_ context.Context) placemodel.HomePayload {
	r.mu.RLock()
	defer r.mu.RUnlock()

	featured := make([]placemodel.HomeFeaturedEvent, 0, len(r.order))
	categoriesSet := map[string]placemodel.HomeCategory{}
	for i, id := range r.order {
		event := r.byID[id]
		if i < 8 {
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
		Collections: []placemodel.HomeCollection{
			{ID: "weekend-picks", Title: "Weekend Picks", Description: "Best events for weekend", ImageURL: "https://example.com/collection.jpg"},
		},
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

	events := make([]placemodel.EventCardView, 0, minInt(2, len(r.order)))
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

	add := func(name string) {
		key := strings.ToLower(strings.TrimSpace(name))
		if key == "" {
			return
		}
		if !strings.Contains(key, query) && !strings.HasPrefix(key, query) {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		items = append(items, placemodel.SearchSuggestion{Name: name})
	}

	for _, id := range r.order {
		event := r.byID[id]
		add(event.Title)
		for _, category := range event.Categories {
			add(category.Name)
		}
		for _, tag := range event.Tags {
			add(tag.Name)
		}
	}
	add("Weekend Picks")

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
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
	r.mu.Lock()
	defer r.mu.Unlock()

	eventID := generateID("event", &r.nextEventID)
	now := time.Now().UTC()
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
		Categories: buildTaxonomyItems(derefStringSlice(input.CategoryIDs)),
		Tags:       buildTaxonomyItems(derefStringSlice(input.TagIDs)),
		Images:     buildImageItems(derefStringSlice(input.ImageURLs), &r.nextImageID),
		Sessions:   buildSessionItems(derefSessions(input.Sessions), &r.nextSessID),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	r.byID[eventID] = event
	r.order = append([]string{eventID}, r.order...)
	return eventID, nil
}

func (r *InMemoryRepository) UpdateEvent(_ context.Context, input placemodel.EventWriteInput) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	event, ok := r.byID[input.ID]
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
		event.Categories = buildTaxonomyItems(*input.CategoryIDs)
	}
	if input.TagIDs != nil {
		event.Tags = buildTaxonomyItems(*input.TagIDs)
	}
	if input.ImageURLs != nil {
		event.Images = buildImageItems(*input.ImageURLs, &r.nextImageID)
	}
	if input.Sessions != nil {
		event.Sessions = buildSessionItems(*input.Sessions, &r.nextSessID)
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func buildImageItems(urls []string, next *int) []placemodel.EventImageView {
	items := make([]placemodel.EventImageView, 0, len(urls))
	for _, imageURL := range urls {
		items = append(items, placemodel.EventImageView{ID: generateID("image", next), ImageURL: imageURL})
	}
	return items
}

func buildSessionItems(sessions []placemodel.EventSessionInput, next *int) []placemodel.EventSessionView {
	items := make([]placemodel.EventSessionView, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, placemodel.EventSessionView{
			ID:      generateID("session", next),
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
	*next = *next + 1
	return prefix + "-" + strconv.Itoa(*next)
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
