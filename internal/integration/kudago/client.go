package kudago

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ClientConfig struct {
	BaseURL        string
	Location       string
	PageSize       int
	RequestTimeout time.Duration
}

type Session struct {
	StartAt time.Time
	EndAt   time.Time
}

type Event struct {
	ExternalID       string
	Title            string
	ShortDescription string
	FullDescription  string
	AgeLimit         int
	SourceURL        string
	Images           []string
	PlaceName        string
	PlaceAddress     string
	Latitude         float64
	Longitude        float64
	Sessions         []Session
	Categories       []string
	Tags             []string
}

type Client struct {
	baseURL  string
	location string
	pageSize int
	http     *http.Client
}

func NewClient(cfg ClientConfig) *Client {
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	pageSize := cfg.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	return &Client{
		baseURL:  strings.TrimRight(cfg.BaseURL, "/"),
		location: strings.TrimSpace(cfg.Location),
		pageSize: pageSize,
		http:     &http.Client{Timeout: timeout},
	}
}

func (c *Client) FetchEvents(ctx context.Context) ([]Event, error) {
	list, err := c.fetchList(ctx)
	if err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(list))
	for _, raw := range list {
		base, ok := parseEvent(raw)
		if !ok {
			continue
		}

		detailRaw, err := c.fetchDetails(ctx, base.ExternalID)
		if err == nil {
			if detail, ok := parseEvent(detailRaw); ok {
				base = mergeEvent(base, detail)
			}
		}

		events = append(events, base)
	}

	return events, nil
}

func (c *Client) fetchList(ctx context.Context) ([]map[string]any, error) {
	u, err := url.Parse(c.baseURL + "/events/")
	if err != nil {
		return nil, fmt.Errorf("parse list url: %w", err)
	}

	query := u.Query()
	if c.location != "" {
		query.Set("location", c.location)
	}
	query.Set("page_size", strconv.Itoa(c.pageSize))
	query.Set("order_by", "-publication_date")
	query.Set("fields", strings.Join([]string{
		"id", "title", "description", "body_text", "site_url", "age_restriction", "images",
		"place", "dates", "categories", "tags",
	}, ","))
	query.Set("expand", "place,dates,categories,tags")
	query.Set("text_format", "text")
	u.RawQuery = query.Encode()

	var payload struct {
		Results []map[string]any `json:"results"`
	}
	if err := c.doJSON(ctx, u.String(), &payload); err != nil {
		return nil, err
	}
	return payload.Results, nil
}

func (c *Client) fetchDetails(ctx context.Context, externalID string) (map[string]any, error) {
	u, err := url.Parse(c.baseURL + "/events/" + url.PathEscape(externalID) + "/")
	if err != nil {
		return nil, fmt.Errorf("parse details url: %w", err)
	}

	query := u.Query()
	query.Set("fields", strings.Join([]string{
		"id", "title", "description", "body_text", "site_url", "age_restriction", "images",
		"place", "dates", "categories", "tags",
	}, ","))
	query.Set("expand", "place,dates,categories,tags")
	query.Set("text_format", "text")
	u.RawQuery = query.Encode()

	var payload map[string]any
	if err := c.doJSON(ctx, u.String(), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) doJSON(ctx context.Context, fullURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request kudago: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from kudago", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func parseEvent(raw map[string]any) (Event, bool) {
	externalID := trim(toString(raw["id"]))
	title := trim(toString(raw["title"]))
	if externalID == "" || title == "" {
		return Event{}, false
	}

	shortDescription := trim(toString(raw["description"]))
	fullDescription := trim(toString(raw["body_text"]))
	sourceURL := trim(toString(raw["site_url"]))
	ageLimit := parseAgeLimit(raw["age_restriction"])
	images := parseImageURLs(raw["images"])
	placeName, placeAddress, lat, lon := parsePlace(raw["place"])
	sessions := parseSessions(raw["dates"])
	categories := parseNameList(raw["categories"])
	tags := parseNameList(raw["tags"])
	if len(tags) == 0 {
		tags = categories
	}

	return Event{
		ExternalID:       externalID,
		Title:            title,
		ShortDescription: shortDescription,
		FullDescription:  fullDescription,
		AgeLimit:         ageLimit,
		SourceURL:        sourceURL,
		Images:           images,
		PlaceName:        placeName,
		PlaceAddress:     placeAddress,
		Latitude:         lat,
		Longitude:        lon,
		Sessions:         sessions,
		Categories:       categories,
		Tags:             tags,
	}, true
}

func mergeEvent(base, detail Event) Event {
	if detail.Title != "" {
		base.Title = detail.Title
	}
	if detail.ShortDescription != "" {
		base.ShortDescription = detail.ShortDescription
	}
	if detail.FullDescription != "" {
		base.FullDescription = detail.FullDescription
	}
	if detail.AgeLimit > 0 {
		base.AgeLimit = detail.AgeLimit
	}
	if detail.SourceURL != "" {
		base.SourceURL = detail.SourceURL
	}
	if len(detail.Images) > 0 {
		base.Images = detail.Images
	}
	if detail.PlaceName != "" && detail.PlaceName != "Неизвестная площадка" {
		base.PlaceName = detail.PlaceName
	}
	if detail.PlaceAddress != "" && detail.PlaceAddress != "Москва" {
		base.PlaceAddress = detail.PlaceAddress
	}
	if detail.Latitude != 0 {
		base.Latitude = detail.Latitude
	}
	if detail.Longitude != 0 {
		base.Longitude = detail.Longitude
	}
	if len(detail.Sessions) > 0 {
		base.Sessions = detail.Sessions
	}
	if len(detail.Categories) > 0 {
		base.Categories = detail.Categories
	}
	if len(detail.Tags) > 0 {
		base.Tags = detail.Tags
	}
	return base
}

func parseSessions(raw any) []Session {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}

	out := make([]Session, 0, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		startUnix := int64FromAny(obj["start"])
		if startUnix <= 0 {
			continue
		}
		endUnix := int64FromAny(obj["end"])
		if endUnix <= startUnix {
			endUnix = startUnix + 2*3600
		}

		out = append(out, Session{
			StartAt: time.Unix(startUnix, 0).UTC(),
			EndAt:   time.Unix(endUnix, 0).UTC(),
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].StartAt.Before(out[j].StartAt) })
	return dedupSessions(out)
}

func dedupSessions(items []Session) []Session {
	seen := make(map[string]struct{}, len(items))
	out := make([]Session, 0, len(items))
	for _, item := range items {
		key := item.StartAt.Format(time.RFC3339) + "|" + item.EndAt.Format(time.RFC3339)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func parsePlace(raw any) (string, string, float64, float64) {
	place, ok := raw.(map[string]any)
	if !ok {
		return "Неизвестная площадка", "Москва", 55.751244, 37.618423
	}

	name := trim(firstNonEmpty(
		toString(place["title"]),
		toString(place["name"]),
	))
	address := trim(firstNonEmpty(
		toString(place["address"]),
		toString(place["location"]),
	))
	if name == "" {
		name = "Неизвестная площадка"
	}
	if address == "" {
		address = "Москва"
	}

	lat := 55.751244
	lon := 37.618423
	if coords, ok := place["coords"].(map[string]any); ok {
		if parsed := float64FromAny(firstNonNil(coords["lat"], coords["latitude"])); parsed != 0 {
			lat = parsed
		}
		if parsed := float64FromAny(firstNonNil(coords["lon"], coords["longitude"])); parsed != 0 {
			lon = parsed
		}
	}

	return name, address, lat, lon
}

func parseImageURLs(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}

	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		image := trim(firstNonEmpty(
			toString(obj["image"]),
			toString(obj["source"]),
		))
		if image == "" {
			continue
		}
		if _, exists := seen[image]; exists {
			continue
		}
		seen[image] = struct{}{}
		out = append(out, image)
	}
	return out
}

func parseNameList(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}

	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		switch value := item.(type) {
		case string:
			name := trim(value)
			if name == "" {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		case map[string]any:
			name := trim(firstNonEmpty(
				toString(value["slug"]),
				toString(value["name"]),
			))
			if name == "" {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}

	sort.Strings(out)
	return out
}

func parseAgeLimit(raw any) int {
	switch v := raw.(type) {
	case float64:
		return clampAge(int(v))
	case int:
		return clampAge(v)
	case string:
		value := strings.TrimSpace(strings.TrimSuffix(v, "+"))
		if value == "" {
			return 0
		}
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0
		}
		return clampAge(parsed)
	default:
		return 0
	}
}

func clampAge(v int) int {
	if v < 0 {
		return 0
	}
	if v > 21 {
		return 21
	}
	return v
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func toString(raw any) string {
	if raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func int64FromAny(raw any) int64 {
	switch v := raw.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func float64FromAny(raw any) float64 {
	switch v := raw.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func trim(value string) string {
	return strings.TrimSpace(value)
}
