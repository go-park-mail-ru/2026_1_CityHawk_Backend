package http

import (
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

type cityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CountryName string `json:"countryName"`
	Timezone    string `json:"timezone"`
}

type userProfileResponse struct {
	ID          string        `json:"id"`
	Username    string        `json:"username"`
	UserSurname string        `json:"userSurname"`
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
			UserSurname: safety.EscapeText(item.UserSurname),
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
