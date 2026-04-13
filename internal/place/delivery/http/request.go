package http

import "encoding/json"

type eventSessionRequest struct {
	PlaceID   string `json:"placeId"`
	PlaceName string `json:"placeName"`
	StartAt   string `json:"startAt"`
	EndAt     string `json:"endAt"`
	Price     int    `json:"price"`
}

type createEventRequest struct {
	Title            string                `json:"title"`
	ShortDescription string                `json:"shortDescription"`
	FullDescription  string                `json:"fullDescription"`
	AgeLimit         *int                  `json:"ageLimit"`
	SourceURL        *string               `json:"sourceUrl"`
	CategoryIDs      []string              `json:"categoryIds"`
	TagIDs           []string              `json:"tagIds"`
	ImageURLs        []string              `json:"imageUrls"`
	Sessions         []eventSessionRequest `json:"sessions"`
}

type optionalString struct {
	Set   bool
	Value *string
}

func (o *optionalString) UnmarshalJSON(data []byte) error {
	o.Set = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	o.Value = &value
	return nil
}

type patchEventRequest struct {
	Title            *string                `json:"title"`
	ShortDescription *string                `json:"shortDescription"`
	FullDescription  *string                `json:"fullDescription"`
	AgeLimit         *int                   `json:"ageLimit"`
	SourceURL        optionalString         `json:"sourceUrl"`
	CategoryIDs      *[]string              `json:"categoryIds"`
	TagIDs           *[]string              `json:"tagIds"`
	ImageURLs        *[]string              `json:"imageUrls"`
	Sessions         *[]eventSessionRequest `json:"sessions"`
}
