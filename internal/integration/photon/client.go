package photon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ClientConfig struct {
	BaseURL        string
	RequestTimeout time.Duration
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type SearchResult struct {
	Name         string
	Street       string
	HouseNumber  string
	City         string
	State        string
	Country      string
	CountryCode  string
	Postcode     string
	District     string
	Latitude     float64
	Longitude    float64
	OSMID        int64
	OSMType      string
	OSMKey       string
	OSMValue     string
	DisplayLabel string
}

type searchResponse struct {
	Features []struct {
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
		Properties photonProperties `json:"properties"`
	} `json:"features"`
}

type photonProperties struct {
	Name        string `json:"name"`
	Street      string `json:"street"`
	HouseNumber string `json:"housenumber"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	CountryCode string `json:"countrycode"`
	Postcode    string `json:"postcode"`
	District    string `json:"district"`
	OSMID       int64  `json:"osm_id"`
	OSMType     string `json:"osm_type"`
	OSMKey      string `json:"osm_key"`
	OSMValue    string `json:"osm_value"`
}

func NewClient(cfg ClientConfig) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
	}
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return []SearchResult{}, nil
	}
	if limit <= 0 {
		limit = 5
	}

	u, err := url.Parse(c.baseURL + "/api")
	if err != nil {
		return nil, err
	}

	params := u.Query()
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("photon search status: %d", resp.StatusCode)
	}

	var payload searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	items := make([]SearchResult, 0, len(payload.Features))
	for _, feature := range payload.Features {
		if len(feature.Geometry.Coordinates) < 2 {
			continue
		}

		item := SearchResult{
			Name:         strings.TrimSpace(feature.Properties.Name),
			Street:       strings.TrimSpace(feature.Properties.Street),
			HouseNumber:  strings.TrimSpace(feature.Properties.HouseNumber),
			City:         strings.TrimSpace(feature.Properties.City),
			State:        strings.TrimSpace(feature.Properties.State),
			Country:      strings.TrimSpace(feature.Properties.Country),
			CountryCode:  strings.TrimSpace(feature.Properties.CountryCode),
			Postcode:     strings.TrimSpace(feature.Properties.Postcode),
			District:     strings.TrimSpace(feature.Properties.District),
			Latitude:     feature.Geometry.Coordinates[1],
			Longitude:    feature.Geometry.Coordinates[0],
			OSMID:        feature.Properties.OSMID,
			OSMType:      strings.TrimSpace(feature.Properties.OSMType),
			OSMKey:       strings.TrimSpace(feature.Properties.OSMKey),
			OSMValue:     strings.TrimSpace(feature.Properties.OSMValue),
			DisplayLabel: buildDisplayLabel(feature.Properties),
		}
		if item.Name == "" {
			item.Name = fallbackName(item)
		}
		if item.Name == "" {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

func buildDisplayLabel(p photonProperties) string {
	parts := make([]string, 0, 5)
	name := strings.TrimSpace(p.Name)
	streetLine := strings.TrimSpace(strings.TrimSpace(p.Street + " " + p.HouseNumber))
	city := strings.TrimSpace(p.City)
	state := strings.TrimSpace(p.State)
	country := strings.TrimSpace(p.Country)

	if name != "" {
		parts = append(parts, name)
	}
	if streetLine != "" && streetLine != name {
		parts = append(parts, streetLine)
	}
	if city != "" {
		parts = append(parts, city)
	}
	if state != "" && state != city {
		parts = append(parts, state)
	}
	if country != "" {
		parts = append(parts, country)
	}

	return strings.Join(parts, ", ")
}

func fallbackName(item SearchResult) string {
	switch {
	case item.Street != "" && item.HouseNumber != "":
		return item.Street + " " + item.HouseNumber
	case item.Street != "":
		return item.Street
	case item.City != "":
		return item.City
	default:
		return ""
	}
}
