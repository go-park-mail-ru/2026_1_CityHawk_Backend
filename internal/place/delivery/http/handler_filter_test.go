package http

import (
	"net/http/httptest"
	"testing"

	placemodel "cityhawk/backend/internal/place/model"
)

func TestParseEventListFilterTagFallback(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/api/events?tag=rock", nil)
	filter, err := parseEventListFilter(req)
	if err != nil {
		t.Fatalf("parseEventListFilter() error = %v", err)
	}
	if filter.TagID != "rock" {
		t.Fatalf("TagID = %q, want %q", filter.TagID, "rock")
	}

	req = httptest.NewRequest("GET", "/api/events?tagSlug=live-show", nil)
	filter, err = parseEventListFilter(req)
	if err != nil {
		t.Fatalf("parseEventListFilter() with tagSlug error = %v", err)
	}
	if filter.TagID != "live-show" {
		t.Fatalf("TagID = %q, want %q", filter.TagID, "live-show")
	}
}

func TestResolveTagFilter(t *testing.T) {
	t.Parallel()

	tags := []placemodel.HomeTag{
		{ID: "tag-1", Name: "Rock", Slug: "rock"},
		{ID: "tag-2", Name: "Рок концерт", Slug: "рокконцерт"},
	}

	if got := resolveTagFilter("tag-1", tags); got != "tag-1" {
		t.Fatalf("resolveTagFilter(id) = %q, want tag-1", got)
	}
	if got := resolveTagFilter("rock", tags); got != "tag-1" {
		t.Fatalf("resolveTagFilter(slug) = %q, want tag-1", got)
	}
	if got := resolveTagFilter("Rock", tags); got != "tag-1" {
		t.Fatalf("resolveTagFilter(name) = %q, want tag-1", got)
	}
	if got := resolveTagFilter("рок-концерт", tags); got != "tag-2" {
		t.Fatalf("resolveTagFilter(normalized) = %q, want tag-2", got)
	}
}
