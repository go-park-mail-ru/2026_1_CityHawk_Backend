package photon

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientSearch(t *testing.T) {
	client := NewClient(ClientConfig{BaseURL: "http://photon.local", RequestTimeout: time.Second})
	client.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api" {
			t.Fatalf("path = %q, want /api", r.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
			"features": [
				{
					"geometry": {"coordinates": [37.6, 55.7]},
					"properties": {
						"name": "ВДНХ",
						"street": "Проспект Мира",
						"housenumber": "119",
						"city": "Москва",
						"country": "Россия",
						"countrycode": "RU",
						"postcode": "129223",
						"district": "Останкинский",
						"osm_id": 1,
						"osm_type": "W",
						"osm_key": "tourism",
						"osm_value": "attraction"
					}
				},
				{
					"geometry": {"coordinates": [37.5]},
					"properties": {}
				}
			]
		}`)),
			Header: make(http.Header),
		}, nil
	})
	items, err := client.Search(context.Background(), "vdnh", 3)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].DisplayLabel == "" || items[0].Latitude != 55.7 || items[0].Longitude != 37.6 {
		t.Fatalf("unexpected item: %+v", items[0])
	}
}

func TestClientSearchEdgeCases(t *testing.T) {
	client := NewClient(ClientConfig{BaseURL: "http://example.com", RequestTimeout: time.Second})
	items, err := client.Search(context.Background(), "", 0)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("len(items) = %d, want 0", len(items))
	}

	client = NewClient(ClientConfig{BaseURL: "http://photon.local", RequestTimeout: time.Second})
	client.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader("bad")),
			Header:     make(http.Header),
		}, nil
	})
	if _, err := client.Search(context.Background(), "x", 1); err == nil {
		t.Fatal("expected status error")
	}
}
