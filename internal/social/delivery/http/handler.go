package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	placemodel "cityhawk/backend/internal/place/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"cityhawk/backend/internal/platform/httpx"
	"cityhawk/backend/internal/platform/media"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	"cityhawk/backend/internal/platform/safety"
	platformsecurity "cityhawk/backend/internal/platform/security"
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
	SearchUsers(ctx context.Context, viewerID, query string, limit, offset int) ([]socialmodel.UserProfile, int, error)
	ListFollowerProfiles(ctx context.Context, userID, viewerID string, limit, offset int) ([]socialmodel.UserProfile, int, error)
	ListFollowingProfiles(ctx context.Context, userID, viewerID string, limit, offset int) ([]socialmodel.UserProfile, int, error)
	UserCollections(ctx context.Context, userID string, limit, offset int) ([]socialmodel.CollectionCard, int, error)
	ListInvitees(ctx context.Context, eventID string) ([]socialmodel.InviteeCandidate, error)
	SearchInvitees(ctx context.Context, eventID, viewerID, query string, limit int) ([]socialmodel.InviteeCandidate, error)
	CreateInvitations(ctx context.Context, senderID, eventID string, recipientIDs []string, message string, eventSessionID *string) ([]socialmodel.Invitation, error)
	UpdateInvitationStatus(ctx context.Context, userID, invitationID, status string) (socialmodel.Invitation, error)
	ListNotifications(ctx context.Context, userID, filterType string, unreadOnly bool, limit, offset int) ([]socialmodel.Notification, int, int, error)
	ListNotificationEvents(ctx context.Context, userID, status string, limit, offset int) ([]socialmodel.NotificationEventRef, int, error)
	MarkNotificationRead(ctx context.Context, userID, notificationID string) (int, error)
	MarkAllNotificationsRead(ctx context.Context, userID string) (int, error)
	CreateEventShareLink(ctx context.Context, creatorUserID, eventID, token string) (socialmodel.ShareLink, error)
	CreateCollectionShareLink(ctx context.Context, creatorUserID, collectionID, token string) (socialmodel.ShareLink, error)
	ResolveShareLink(ctx context.Context, token string) (socialmodel.ShareLink, bool, error)
}

type EventsReader interface {
	GetByID(ctx context.Context, id, userID string) (placemodel.EventDetailsView, bool, error)
}

func (h *Handler) InviteesSearch(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		eventID, ok := eventSubpath(r.URL.Path, "invitees/search")
		if !ok {
			return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
		}
		query := strings.TrimSpace(r.URL.Query().Get("query"))
		if len([]rune(query)) < 2 {
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"query": "query must contain at least 2 characters"})
		}
		limit, err := parseBoundedLimit(r, 10, 5, 10)
		if err != nil {
			return err
		}
		items, err := h.repo.SearchInvitees(r.Context(), eventID, userID, query, limit)
		if err != nil {
			if errors.Is(err, platformerrors.ErrInvalidReference) {
				return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
			}
			return err
		}
		httpx.WriteJSON(w, http.StatusOK, inviteeListResponse{Items: inviteesResponse(items)})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Invitees(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		eventID, ok := eventSubpath(r.URL.Path, "invitees")
		if !ok {
			return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
		}
		items, err := h.repo.ListInvitees(r.Context(), eventID)
		if err != nil {
			if errors.Is(err, platformerrors.ErrInvalidReference) {
				return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
			}
			return err
		}
		httpx.WriteJSON(w, http.StatusOK, inviteeListResponse{Items: inviteesResponse(items)})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) EventInvitations(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		eventID, ok := eventSubpath(r.URL.Path, "invitations")
		if !ok {
			return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
		}
		var req createInvitationsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return httpx.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
		}
		recipients := compactStrings(req.RecipientIDs)
		if len(recipients) == 0 || len(recipients) > 20 {
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"recipientIds": "recipientIds must contain from 1 to 20 ids"})
		}
		message := strings.TrimSpace(req.Message)
		if len([]rune(message)) > 500 {
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"message": "message must be at most 500 characters"})
		}
		var sessionID *string
		if strings.TrimSpace(req.EventSessionID) != "" {
			value := strings.TrimSpace(req.EventSessionID)
			sessionID = &value
		}
		items, err := h.repo.CreateInvitations(r.Context(), userID, eventID, recipients, message, sessionID)
		if err != nil {
			if errors.Is(err, platformerrors.ErrOnlyFriendsInvite) {
				httpx.WriteJSON(w, http.StatusForbidden, map[string]string{
					"error":   "only_friends_can_be_invited",
					"message": "Можно приглашать только друзей",
				})
				return nil
			}
			return mapInvitationWriteError(err)
		}
		httpx.WriteJSON(w, http.StatusCreated, invitationListResponse{Items: invitationsResponse(items)})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) InvitationByID(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		invitationID := strings.TrimPrefix(r.URL.Path, "/api/invitations/")
		if invitationID == "" || strings.Contains(invitationID, "/") {
			return httpx.NewHTTPError(http.StatusNotFound, "Invitation not found")
		}
		var req updateInvitationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return httpx.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
		}
		status := strings.TrimSpace(req.Status)
		if status != "accepted" && status != "declined" {
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"status": "status must be accepted or declined"})
		}
		item, err := h.repo.UpdateInvitationStatus(r.Context(), userID, invitationID, status)
		if err != nil {
			switch {
			case errors.Is(err, platformerrors.ErrNotFound):
				return httpx.NewHTTPError(http.StatusNotFound, "Invitation not found")
			case errors.Is(err, platformerrors.ErrForbidden):
				return httpx.NewHTTPError(http.StatusForbidden, "Forbidden")
			case errors.Is(err, platformerrors.ErrAlreadyExists):
				return httpx.NewHTTPError(http.StatusConflict, "Invitation is not pending")
			default:
				return err
			}
		}
		httpx.WriteJSON(w, http.StatusOK, invitationResponse(item))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) Notifications(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		filterType := strings.TrimSpace(r.URL.Query().Get("type"))
		if filterType == "" {
			filterType = "all"
		}
		if filterType != "all" && filterType != "invitations" && filterType != "system" {
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"type": "unsupported notification type"})
		}
		unreadOnly := strings.EqualFold(r.URL.Query().Get("unreadOnly"), "true")
		limit, offset, err := parsePage(r, 20, maxLimit)
		if err != nil {
			return err
		}
		items, total, unread, err := h.repo.ListNotifications(r.Context(), userID, filterType, unreadOnly, limit, offset)
		if err != nil {
			return err
		}
		httpx.WriteJSON(w, http.StatusOK, notificationListResponse{
			Items:       notificationsResponse(items),
			UnreadCount: unread,
			Total:       total,
			Limit:       limit,
			Offset:      offset,
		})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) NotificationEvents(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		limit, offset, err := parsePage(r, 12, maxLimit)
		if err != nil {
			return err
		}
		status := strings.TrimSpace(r.URL.Query().Get("status"))
		if status != "" && status != "pending" && status != "accepted" && status != "declined" && status != "cancelled" {
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"status": "status must be one of pending, accepted, declined, cancelled"})
		}
		refs, total, err := h.repo.ListNotificationEvents(r.Context(), userID, status, limit, offset)
		if err != nil {
			return err
		}
		items := make([]notificationEventCardResponse, 0, len(refs))
		for _, ref := range refs {
			event, ok, err := h.events.GetByID(r.Context(), ref.EventID, userID)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			items = append(items, notificationEventCard(event, ref))
		}
		httpx.WriteJSON(w, http.StatusOK, notificationEventListResponse{Items: items, Total: total, Limit: limit, Offset: offset})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) NotificationByID(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		notificationID := strings.TrimPrefix(r.URL.Path, "/api/me/notifications/")
		notificationID = strings.TrimSuffix(notificationID, "/read")
		if notificationID == "" || strings.Contains(notificationID, "/") {
			return httpx.NewHTTPError(http.StatusNotFound, "Notification not found")
		}
		unread, err := h.repo.MarkNotificationRead(r.Context(), userID, notificationID)
		if err != nil {
			if errors.Is(err, platformerrors.ErrNotFound) {
				return httpx.NewHTTPError(http.StatusNotFound, "Notification not found")
			}
			return err
		}
		httpx.WriteJSON(w, http.StatusOK, readNotificationResponse{OK: true, UnreadCount: unread})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) NotificationsReadAll(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		unread, err := h.repo.MarkAllNotificationsRead(r.Context(), userID)
		if err != nil {
			return err
		}
		httpx.WriteJSON(w, http.StatusOK, readNotificationResponse{OK: true, UnreadCount: unread})
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) EventShareLinks(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		eventID, ok := eventSubpath(r.URL.Path, "share-links")
		if !ok {
			return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
		}
		userID, _ := optionalUserID(r)
		token, err := platformsecurity.RandomHex(6)
		if err != nil {
			return err
		}
		item, err := h.repo.CreateEventShareLink(r.Context(), userID, eventID, token)
		if err != nil {
			return mapShareLinkError(err, "Event not found")
		}
		item.URL = shareURL(r, item.Token)
		httpx.WriteJSON(w, http.StatusCreated, shareLinkResponseFromModel(item))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) CollectionShareLinks(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		collectionID := strings.TrimPrefix(r.URL.Path, "/api/collections/")
		collectionID = strings.TrimSuffix(collectionID, "/share-links")
		if collectionID == "" || strings.Contains(collectionID, "/") {
			return httpx.NewHTTPError(http.StatusNotFound, "Collection not found")
		}
		userID, _ := optionalUserID(r)
		token, err := platformsecurity.RandomHex(6)
		if err != nil {
			return err
		}
		item, err := h.repo.CreateCollectionShareLink(r.Context(), userID, collectionID, token)
		if err != nil {
			return mapShareLinkError(err, "Collection not found")
		}
		item.URL = shareURL(r, item.Token)
		httpx.WriteJSON(w, http.StatusCreated, shareLinkResponseFromModel(item))
		return nil
	}).ServeHTTP(w, r)
}

func (h *Handler) ShareRedirect(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		token := strings.TrimPrefix(r.URL.Path, "/s/")
		if token == "" || strings.Contains(token, "/") {
			return httpx.NewHTTPError(http.StatusNotFound, "Share link not found")
		}
		item, ok, err := h.repo.ResolveShareLink(r.Context(), token)
		if err != nil {
			return err
		}
		if !ok {
			return httpx.NewHTTPError(http.StatusNotFound, "Share link not found")
		}
		if item.EventID != nil {
			http.Redirect(w, r, "/events/"+*item.EventID, http.StatusFound)
			return nil
		}
		if item.CollectionID != nil {
			http.Redirect(w, r, "/collections/"+*item.CollectionID, http.StatusFound)
			return nil
		}
		return httpx.NewHTTPError(http.StatusNotFound, "Share link not found")
	}).ServeHTTP(w, r)
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

func (h *Handler) UsersSearch(w http.ResponseWriter, r *http.Request) {
	platformmiddleware.ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return httpx.NewHTTPError(http.StatusMethodNotAllowed, "method not allowed")
		}
		userID, err := requireUserID(r)
		if err != nil {
			return err
		}
		query := strings.TrimSpace(r.URL.Query().Get("query"))
		if len([]rune(query)) < 2 {
			return httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"query": "query must contain at least 2 characters"})
		}
		limit, offset, err := parsePage(r, defaultLimit, maxLimit)
		if err != nil {
			return err
		}
		items, total, err := h.repo.SearchUsers(r.Context(), userID, query, limit, offset)
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

func optionalUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(httpx.UserIDContextKey).(string)
	userID = strings.TrimSpace(userID)
	return userID, ok && userID != ""
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

func parseBoundedLimit(r *http.Request, fallback, minValue, maxValue int) (int, error) {
	limit := fallback
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			return 0, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"limit": "limit must be an integer"})
		}
		limit = value
	}
	if limit < minValue || limit > maxValue {
		return 0, httpx.NewHTTPErrorWithDetails(http.StatusBadRequest, "Validation failed", map[string]string{"limit": "limit is out of range"})
	}
	return limit, nil
}

func eventSubpath(path, suffix string) (string, bool) {
	rest := strings.TrimPrefix(path, "/api/events/")
	want := "/" + suffix
	if !strings.HasSuffix(rest, want) {
		return "", false
	}
	eventID := strings.TrimSuffix(rest, want)
	return eventID, eventID != "" && !strings.Contains(eventID, "/")
}

func compactStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func mapInvitationWriteError(err error) error {
	switch {
	case errors.Is(err, platformerrors.ErrInvalidReference):
		return httpx.NewHTTPError(http.StatusNotFound, "Event not found")
	case errors.Is(err, platformerrors.ErrAlreadyExists):
		return httpx.NewHTTPError(http.StatusConflict, "Invitation already exists")
	default:
		return err
	}
}

func mapShareLinkError(err error, notFound string) error {
	switch {
	case errors.Is(err, platformerrors.ErrInvalidReference):
		return httpx.NewHTTPError(http.StatusNotFound, notFound)
	case errors.Is(err, platformerrors.ErrAlreadyExists):
		return httpx.NewHTTPError(http.StatusConflict, "Share token already exists")
	default:
		return err
	}
}

func shareURL(r *http.Request, token string) string {
	if baseURL := configuredShareBaseURL(); baseURL != "" {
		return baseURL + "/s/" + token
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		scheme = strings.Split(proto, ",")[0]
	}
	return scheme + "://" + r.Host + "/s/" + token
}

func configuredShareBaseURL() string {
	for _, name := range []string{"SHARE_BASE_URL", "PUBLIC_BASE_URL"} {
		if value := strings.TrimRight(strings.TrimSpace(os.Getenv(name)), "/"); value != "" {
			return value
		}
	}
	return ""
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

func notificationEventCard(item placemodel.EventDetailsView, ref socialmodel.NotificationEventRef) notificationEventCardResponse {
	card := eventDetailsToCard(item)
	resp := notificationEventCardResponse{
		ID:               card.ID,
		Title:            card.Title,
		ShortDescription: card.ShortDescription,
		CoverImageURL:    card.CoverImageURL,
		IsFavorite:       card.IsFavorite,
		Tags:             card.Tags,
		NextSession:      card.NextSession,
	}
	if ref.InvitedBy != nil {
		resp.InvitedBy = &notificationEventInvitedByResponse{
			ID:          ref.InvitedBy.ID,
			Username:    safety.EscapeText(ref.InvitedBy.Username),
			DisplayName: safety.EscapeText(ref.InvitedBy.Username),
			AvatarURL:   media.PublicURLPtr(ref.InvitedBy.AvatarURL),
		}
	}
	if ref.Invitation != nil {
		resp.InvitationStatus = ref.Invitation.Status
	}
	return resp
}
