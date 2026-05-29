package http

import (
	"time"

	"cityhawk/backend/internal/platform/media"
	"cityhawk/backend/internal/platform/safety"
	socialmodel "cityhawk/backend/internal/social/model"
)

type okResponse struct {
	OK bool `json:"ok"`
}

type taxonomyItemResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type eventCardNextSessionPlaceResponse struct {
	Name        string `json:"name"`
	AddressLine string `json:"addressLine"`
}

type eventCardNextSessionResponse struct {
	StartAt string                            `json:"startAt"`
	Place   eventCardNextSessionPlaceResponse `json:"place"`
}

type eventCardResponse struct {
	ID               string                        `json:"id"`
	Title            string                        `json:"title"`
	ShortDescription string                        `json:"shortDescription"`
	CoverImageURL    string                        `json:"coverImageUrl"`
	IsFavorite       bool                          `json:"isFavorite"`
	Tags             []taxonomyItemResponse        `json:"tags"`
	NextSession      *eventCardNextSessionResponse `json:"nextSession"`
}

type eventListResponse struct {
	Items  []eventCardResponse `json:"items"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

type notificationEventInvitedByResponse struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
}

type notificationEventCardResponse struct {
	ID               string                              `json:"id"`
	Title            string                              `json:"title"`
	ShortDescription string                              `json:"shortDescription"`
	CoverImageURL    string                              `json:"coverImageUrl"`
	IsFavorite       bool                                `json:"isFavorite"`
	InvitationStatus string                              `json:"invitationStatus,omitempty"`
	Tags             []taxonomyItemResponse              `json:"tags"`
	NextSession      *eventCardNextSessionResponse       `json:"nextSession"`
	InvitedBy        *notificationEventInvitedByResponse `json:"invitedBy"`
}

type notificationEventListResponse struct {
	Items  []notificationEventCardResponse `json:"items"`
	Total  int                             `json:"total"`
	Limit  int                             `json:"limit"`
	Offset int                             `json:"offset"`
}

type cityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CountryName string `json:"countryName"`
	Timezone    string `json:"timezone"`
}

type userProfileResponse struct {
	ID          string        `json:"id"`
	Username    string        `json:"username"`
	AvatarURL   *string       `json:"avatarUrl"`
	City        *cityResponse `json:"city"`
	IsFollowing bool          `json:"isFollowing"`
}

type userListResponse struct {
	Items  []userProfileResponse `json:"items"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

type collectionCardResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	IsPublic    bool   `json:"isPublic"`
}

type collectionListResponse struct {
	Items  []collectionCardResponse `json:"items"`
	Total  int                      `json:"total"`
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
}

type inviteeResponse struct {
	ID               string        `json:"id"`
	Username         string        `json:"username"`
	AvatarURL        *string       `json:"avatarUrl"`
	City             *cityResponse `json:"city"`
	IsFollowing      bool          `json:"isFollowing,omitempty"`
	IsFriend         bool          `json:"isFriend,omitempty"`
	InvitationStatus *string       `json:"invitationStatus"`
}

type inviteeListResponse struct {
	Items []inviteeResponse `json:"items"`
}

type invitationResponseBody struct {
	ID             string  `json:"id"`
	EventID        string  `json:"eventId"`
	EventSessionID *string `json:"eventSessionId"`
	SenderID       string  `json:"senderId"`
	RecipientID    string  `json:"recipientId"`
	Status         string  `json:"status"`
	Message        string  `json:"message"`
	RespondedAt    *string `json:"respondedAt,omitempty"`
	CreatedAt      string  `json:"createdAt,omitempty"`
	UpdatedAt      string  `json:"updatedAt"`
}

type invitationListResponse struct {
	Items []invitationResponseBody `json:"items"`
}

type notificationActorResponse struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
}

type notificationEventResponse struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	CoverImageURL string `json:"coverImageUrl"`
	DateText      string `json:"dateText"`
	PlaceText     string `json:"placeText"`
}

type notificationEventSessionResponse struct {
	ID      string `json:"id"`
	StartAt string `json:"startAt"`
	EndAt   string `json:"endAt"`
}

type notificationInvitationResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type notificationCollectionResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	ImageURL string `json:"imageUrl"`
}

type notificationResponse struct {
	ID           string                            `json:"id"`
	Type         string                            `json:"type"`
	Title        string                            `json:"title"`
	Message      string                            `json:"message"`
	IsRead       bool                              `json:"isRead"`
	ReadAt       *string                           `json:"readAt"`
	CreatedAt    string                            `json:"createdAt"`
	Actor        *notificationActorResponse        `json:"actor"`
	Event        *notificationEventResponse        `json:"event"`
	EventSession *notificationEventSessionResponse `json:"eventSession"`
	Invitation   *notificationInvitationResponse   `json:"invitation"`
	Collection   *notificationCollectionResponse   `json:"collection"`
}

type notificationListResponse struct {
	Items       []notificationResponse `json:"items"`
	UnreadCount int                    `json:"unreadCount"`
	Total       int                    `json:"total"`
	Limit       int                    `json:"limit"`
	Offset      int                    `json:"offset"`
}

type readNotificationResponse struct {
	OK          bool `json:"ok"`
	UnreadCount int  `json:"unreadCount"`
}

type shareLinkResponse struct {
	ID           string  `json:"id"`
	URL          string  `json:"url"`
	Token        string  `json:"token"`
	EventID      *string `json:"eventId,omitempty"`
	CollectionID *string `json:"collectionId,omitempty"`
	CreatedAt    string  `json:"createdAt"`
}

func userProfilesResponse(items []socialmodel.UserProfile) []userProfileResponse {
	responses := make([]userProfileResponse, 0, len(items))
	for _, item := range items {
		var city *cityResponse
		if item.City != nil {
			city = &cityResponse{
				ID:          item.City.ID,
				Name:        safety.EscapeText(item.City.Name),
				CountryName: safety.EscapeText(item.City.CountryName),
				Timezone:    item.City.Timezone,
			}
		}
		responses = append(responses, userProfileResponse{
			ID:          item.ID,
			Username:    safety.EscapeText(item.Username),
			AvatarURL:   media.PublicURLPtr(item.AvatarURL),
			City:        city,
			IsFollowing: item.IsFollowing,
		})
	}
	return responses
}

func collectionCardsResponse(items []socialmodel.CollectionCard) []collectionCardResponse {
	responses := make([]collectionCardResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, collectionCardResponse{
			ID:          item.ID,
			Title:       safety.EscapeText(item.Title),
			Description: safety.EscapeText(item.Description),
			ImageURL:    media.PublicURL(item.ImageURL),
			IsPublic:    item.IsPublic,
		})
	}
	return responses
}

func inviteesResponse(items []socialmodel.InviteeCandidate) []inviteeResponse {
	responses := make([]inviteeResponse, 0, len(items))
	for _, item := range items {
		var city *cityResponse
		if item.City != nil {
			city = &cityResponse{
				ID:          item.City.ID,
				Name:        safety.EscapeText(item.City.Name),
				CountryName: safety.EscapeText(item.City.CountryName),
				Timezone:    item.City.Timezone,
			}
		}
		responses = append(responses, inviteeResponse{
			ID:               item.ID,
			Username:         safety.EscapeText(item.Username),
			AvatarURL:        media.PublicURLPtr(item.AvatarURL),
			City:             city,
			IsFollowing:      item.IsFollowing,
			IsFriend:         item.IsFriend,
			InvitationStatus: item.InvitationStatus,
		})
	}
	return responses
}

func invitationsResponse(items []socialmodel.Invitation) []invitationResponseBody {
	responses := make([]invitationResponseBody, 0, len(items))
	for _, item := range items {
		responses = append(responses, invitationResponse(item))
	}
	return responses
}

func invitationResponse(item socialmodel.Invitation) invitationResponseBody {
	var respondedAt *string
	if item.RespondedAt != nil {
		value := formatTime(*item.RespondedAt)
		respondedAt = &value
	}
	return invitationResponseBody{
		ID:             item.ID,
		EventID:        item.EventID,
		EventSessionID: item.EventSessionID,
		SenderID:       item.SenderID,
		RecipientID:    item.RecipientID,
		Status:         item.Status,
		Message:        safety.EscapeText(item.Message),
		RespondedAt:    respondedAt,
		CreatedAt:      formatTime(item.CreatedAt),
		UpdatedAt:      formatTime(item.UpdatedAt),
	}
}

func notificationsResponse(items []socialmodel.Notification) []notificationResponse {
	responses := make([]notificationResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, notificationResponseFromModel(item))
	}
	return responses
}

func notificationResponseFromModel(item socialmodel.Notification) notificationResponse {
	var readAt *string
	if item.ReadAt != nil {
		value := formatTime(*item.ReadAt)
		readAt = &value
	}
	resp := notificationResponse{
		ID:        item.ID,
		Type:      item.Type,
		Title:     notificationTitle(item),
		Message:   safety.EscapeText(item.Message),
		IsRead:    item.IsRead,
		ReadAt:    readAt,
		CreatedAt: formatTime(item.CreatedAt),
	}
	if item.Actor != nil {
		resp.Actor = &notificationActorResponse{
			ID:          item.Actor.ID,
			DisplayName: safety.EscapeText(item.Actor.DisplayName),
			AvatarURL:   media.PublicURLPtr(item.Actor.AvatarURL),
		}
	}
	if item.Event != nil {
		resp.Event = &notificationEventResponse{
			ID:            item.Event.ID,
			Title:         safety.EscapeText(item.Event.Title),
			CoverImageURL: media.PublicURL(item.Event.CoverImageURL),
			DateText:      item.Event.DateText,
			PlaceText:     safety.EscapeText(item.Event.PlaceText),
		}
	}
	if item.EventSession != nil {
		resp.EventSession = &notificationEventSessionResponse{
			ID:      item.EventSession.ID,
			StartAt: formatTime(item.EventSession.StartAt),
			EndAt:   formatTime(item.EventSession.EndAt),
		}
	}
	if item.Invitation != nil {
		resp.Invitation = &notificationInvitationResponse{ID: item.Invitation.ID, Status: item.Invitation.Status}
	}
	if item.Collection != nil {
		resp.Collection = &notificationCollectionResponse{
			ID:       item.Collection.ID,
			Title:    safety.EscapeText(item.Collection.Title),
			ImageURL: media.PublicURL(item.Collection.ImageURL),
		}
	}
	return resp
}

func notificationTitle(item socialmodel.Notification) string {
	actor := "Пользователь"
	if item.Actor != nil && item.Actor.DisplayName != "" {
		actor = item.Actor.DisplayName
	}
	event := "мероприятие"
	if item.Event != nil && item.Event.Title != "" {
		event = item.Event.Title
	}
	switch item.Type {
	case "event_invitation":
		return safety.EscapeText(actor + " пригласил вас на " + event)
	case "invitation_accepted":
		return safety.EscapeText(actor + " принял приглашение")
	case "invitation_declined":
		return safety.EscapeText(actor + " отклонил приглашение")
	case "collection_shared":
		return "С вами поделились подборкой"
	case "event_reminder":
		return "Напоминание о событии"
	default:
		return "Уведомление"
	}
}

func shareLinkResponseFromModel(item socialmodel.ShareLink) shareLinkResponse {
	return shareLinkResponse{
		ID:           item.ID,
		URL:          item.URL,
		Token:        item.Token,
		EventID:      item.EventID,
		CollectionID: item.CollectionID,
		CreatedAt:    formatTime(item.CreatedAt),
	}
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
