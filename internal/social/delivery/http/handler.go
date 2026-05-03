package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	placemodel "cityhawk/backend/internal/place/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
	"cityhawk/backend/internal/platform/media"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	"cityhawk/backend/internal/platform/safety"
	socialmodel "cityhawk/backend/internal/social/model"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Repository interface {
	AddFavorite(ctx context.Context, userID, eventID string) error
	RemoveFavorite(ctx context.Context, userID, eventID string) error
	ListFavoriteEvents(ctx context.Context, userID string, limit, offset int) ([]socialmodel.FavoriteEvent, error)
	CountFavoriteEvents(ctx context.Context, userID string) (int, error)
	FollowUser(ctx context.Context, followerUserID, followedUserID string) error
	UnfollowUser(ctx context.Context, followerUserID, followedUserID string) error
	ListFollowerProfiles(ctx context.Context, userID, viewerID string, limit, offset int) ([]socialmodel.UserProfile, int, error)
	ListFollowingProfiles(ctx context.Context, userID, viewerID string, limit, offset int) ([]socialmodel.UserProfile, int, error)
	UserCollections(ctx context.Context, userID string, limit, offset int) ([]socialmodel.CollectionCard, int, error)
}

type EventsReader interface {
	GetByID(ctx context.Context, id, userID string) (placemodel.EventDetailsView, bool, error)
}

type Handler struct {
	repo   Repository
	events EventsReader
}

func NewHandler(repo Repository, events EventsReader) *Handler {
	return &Handler{repo: repo, events: events}
}

func (h *Handler) FavoriteByID(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		eventID := strings.TrimPrefix(r.URL.Path, "/api/me/favorites/")
		if eventID == "" || strings.Contains(eventID, "/") {
			return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
		}

		switch r.Method {
		case http.MethodPost:
			if err := h.repo.AddFavorite(r.Context(), userID, eventID); err != nil {
				return mapSocialWriteError(err, "Event not found", "Already in favorites")
			}
			httpx.WriteJSON(w, http.StatusOK, okResponse{OK: true})
			return nil
		case http.MethodDelete:
			if err := h.repo.RemoveFavorite(r.Context(), userID, eventID); err != nil {
				return mapSocialWriteError(err, "Event not found", "")
			}
			httpx.WriteJSON(w, http.StatusOK, okResponse{OK: true})
			return nil
		default:
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
	}).ServeHTTP(w, r)
}

func (h *Handler) Favorites(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		limit, offset, err := parsePage(r, 12, 0)
		if err != nil {
			return err
		}

		favorites, err := h.repo.ListFavoriteEvents(r.Context(), userID, limit, offset)
		if err != nil {
			return err
		}
		total, err := h.repo.CountFavoriteEvents(r.Context(), userID)
		if err != nil {
			return err
		}
		items := make([]eventCardResponse, 0, len(favorites))
		for _, favorite := range favorites {
			event, ok, err := h.events.GetByID(r.Context(), favorite.EventID, userID)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			card := eventDetailsToCard(event)
			card.IsFavorite = true
			items = append(items, card)
		}
		httpx.WriteJSON(w, http.StatusOK, eventListResponse{Items: items, Total: total, Limit: limit, Offset: offset})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Followers(w http.ResponseWriter, r *http.Request) {
	h.userList(w, r, true)
}

func (h *Handler) Following(w http.ResponseWriter, r *http.Request) {
	h.userList(w, r, false)
}

func (h *Handler) userList(w http.ResponseWriter, r *http.Request, followers bool) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		limit, offset, err := parsePage(r, defaultLimit, maxLimit)
		if err != nil {
			return err
		}
		var (
			items []socialmodel.UserProfile
			total int
		)
		if followers {
			items, total, err = h.repo.ListFollowerProfiles(r.Context(), userID, userID, limit, offset)
		} else {
			items, total, err = h.repo.ListFollowingProfiles(r.Context(), userID, userID, limit, offset)
		}
		if err != nil {
			return err
		}
		httpx.WriteJSON(w, http.StatusOK, userListResponse{
			Items:  userProfilesResponse(items),
			Total:  total,
			Limit:  limit,
			Offset: offset,
		})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) FollowByID(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		rest := strings.TrimPrefix(r.URL.Path, "/api/users/")
		if !strings.HasSuffix(rest, "/follow") {
			return httpx.NewHTTPError(http.StatusNotFound, "User not found")
		}
		targetID := strings.TrimSuffix(rest, "/follow")
		if targetID == "" || strings.Contains(targetID, "/") {
			return httpx.NewHTTPError(http.StatusNotFound, "User not found")
		}

		switch r.Method {
		case http.MethodPost:
			if err := h.repo.FollowUser(r.Context(), userID, targetID); err != nil {
				return mapSocialWriteError(err, "User not found", "Already following")
			}
			httpx.WriteJSON(w, http.StatusOK, okResponse{OK: true})
			return nil
		case http.MethodDelete:
			if err := h.repo.UnfollowUser(r.Context(), userID, targetID); err != nil {
				return mapSocialWriteError(err, "User not found", "")
			}
			httpx.WriteJSON(w, http.StatusOK, okResponse{OK: true})
			return nil
		default:
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
	}).ServeHTTP(w, r)
}

func (h *Handler) Collections(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		limit, offset, err := parsePage(r, 12, 0)
		if err != nil {
			return err
		}
		items, total, err := h.repo.UserCollections(r.Context(), userID, limit, offset)
		if err != nil {
			return err
		}
		httpx.WriteJSON(w, http.StatusOK, collectionListResponse{
			Items:  collectionCardsResponse(items),
			Total:  total,
			Limit:  limit,
			Offset: offset,
		})
		return nil
	}).ServeHTTP(w, r)
}

func requireUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(httpx.UserIDContextKey).(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", httpx.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	return strings.TrimSpace(userID), nil
}

func parsePage(r *http.Request, fallbackLimit int, maxLimit int) (int, int, error) {
	limit := fallbackLimit
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return 0, 0, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"limit": "limit must be a positive integer"})
		}
		limit = value
	}
	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}
	offset := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return 0, 0, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"offset": "offset must be a non-negative integer"})
		}
		offset = value
	}
	return limit, offset, nil
}

func mapSocialWriteError(err error, notFound string, conflict string) error {
	switch {
	case errors.Is(err, platformerrors.ErrInvalidReference):
		return httpx.NewHTTPError(http.StatusNotFound, notFound)
	case errors.Is(err, platformerrors.ErrAlreadyExists):
		if conflict == "" {
			conflict = "Already exists"
		}
		return httpx.NewHTTPError(http.StatusConflict, conflict)
	default:
		return err
	}
}

func eventDetailsToCard(item placemodel.EventDetailsView) eventCardResponse {
	tags := make([]taxonomyItemResponse, 0, len(item.Tags))
	for _, tag := range item.Tags {
		tags = append(tags, taxonomyItemResponse{ID: tag.ID, Name: safety.EscapeText(tag.Name), Slug: tag.Slug})
	}
	cover := ""
	if len(item.Images) > 0 {
		cover = media.PublicURL(item.Images[0].ImageURL)
	}
	var nextSession *eventCardNextSessionResponse
	if len(item.Sessions) > 0 {
		session := item.Sessions[0]
		nextSession = &eventCardNextSessionResponse{
			StartAt: session.StartAt.UTC().Format("2006-01-02T15:04:05Z"),
			Place: eventCardNextSessionPlaceResponse{
				Name:        safety.EscapeText(session.Place.Name),
				AddressLine: safety.EscapeText(session.Place.AddressLine),
			},
		}
	}
	return eventCardResponse{
		ID:               item.ID,
		Title:            safety.EscapeText(item.Title),
		ShortDescription: safety.EscapeText(item.ShortDescription),
		CoverImageURL:    cover,
		Tags:             tags,
		NextSession:      nextSession,
		IsFavorite:       item.IsFavorite,
	}
}
