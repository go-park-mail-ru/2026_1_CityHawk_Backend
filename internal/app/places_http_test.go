package app

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	placedelivery "cityhawk/backend/internal/place/delivery/http"
	placerepo "cityhawk/backend/internal/place/repository"
	placeusecase "cityhawk/backend/internal/place/usecase"
	"cityhawk/backend/internal/platform/httpx"
	"cityhawk/backend/internal/platform/media"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
)

func TestEventsHandlers(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	imageStore, imageDir := newTestEventImageStorage(t)
	handler := placedelivery.NewHandler(placeusecase.NewService(repo), imageStore)

	t.Run("list ok and method not allowed", func(t *testing.T) {
		h := http.HandlerFunc(handler.Events)

		req := httptest.NewRequest(http.MethodGet, "/api/events?limit=12&offset=0", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("list status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		items, ok := payload["items"].([]any)
		if !ok || len(items) == 0 {
			t.Fatalf("events response has no items: %+v", payload)
		}
		if payload["total"] == nil || payload["limit"] != float64(12) || payload["offset"] != float64(0) {
			t.Fatalf("unexpected list pagination: %+v", payload)
		}

		reqBad := httptest.NewRequest(http.MethodPut, "/api/events", nil)
		recBad := httptest.NewRecorder()
		h.ServeHTTP(recBad, reqBad)
		if recBad.Code != http.StatusMethodNotAllowed {
			t.Fatalf("bad method status = %d, want %d", recBad.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("list by tag id", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.Handle("GET /api/tags/{id}/events", http.HandlerFunc(handler.EventsByTag))

		req := httptest.NewRequest(http.MethodGet, "/api/tags/item/events?limit=12&offset=0", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("list by tag status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		items, ok := payload["items"].([]any)
		if !ok || len(items) == 0 {
			t.Fatalf("events by tag response has no items: %+v", payload)
		}
		for _, raw := range items {
			item, ok := raw.(map[string]any)
			if !ok {
				t.Fatalf("unexpected item type: %#v", raw)
			}
			tags, ok := item["tags"].([]any)
			if !ok {
				t.Fatalf("missing tags in item: %+v", item)
			}
			found := false
			for _, rawTag := range tags {
				tag, ok := rawTag.(map[string]any)
				if ok && tag["id"] == "item" {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("item does not contain requested tag: %+v", item)
			}
		}
	})

	t.Run("details get", func(t *testing.T) {
		h := http.HandlerFunc(handler.EventByID)

		req := httptest.NewRequest(http.MethodGet, "/api/events/futurione", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("details status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		if payload["id"] != "futurione" {
			t.Fatalf("unexpected details response: %+v", payload)
		}
		if _, ok := payload["sessions"].([]any); !ok {
			t.Fatalf("missing sessions in details: %+v", payload)
		}

		missingReq := httptest.NewRequest(http.MethodGet, "/api/events/unknown", nil)
		missingRec := httptest.NewRecorder()
		h.ServeHTTP(missingRec, missingReq)
		if missingRec.Code != http.StatusNotFound {
			t.Fatalf("details missing status = %d, want %d", missingRec.Code, http.StatusNotFound)
		}
	})

	t.Run("create patch delete", func(t *testing.T) {
		events := http.HandlerFunc(handler.Events)
		eventByID := http.HandlerFunc(handler.EventByID)

		createBody, createContentType := mustMultipartEventBody(t, map[string]string{
			"title":            "New Event",
			"shortDescription": "Short text",
			"fullDescription":  "Long event description",
			"categoryIds":      `["music"]`,
			"tagIds":           `["rock"]`,
			"imageUrls":        `["https://example.com/new.jpg"]`,
			"sessions":         `[{"placeId":"place-1","startAt":"2026-04-20T19:00:00Z","endAt":"2026-04-20T21:00:00Z","price":1200}]`,
		}, "images", []namedFile{{name: "cover.png", content: tinyPNG()}})
		createReq := httptest.NewRequest(http.MethodPost, "/api/events", createBody)
		createReq.Header.Set("Content-Type", createContentType)
		createReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: "test-csrf"})
		createReq.Header.Set(httpx.CSRFHeader, "test-csrf")
		createReq = createReq.WithContext(context.WithValue(createReq.Context(), httpx.UserIDContextKey, "user-1"))
		createRec := httptest.NewRecorder()
		platformmiddleware.CSRFMiddleware(events, func(r *http.Request) string {
			c, err := r.Cookie("csrf_token")
			if err != nil {
				return ""
			}
			return c.Value
		}).ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("create status = %d, want %d body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
		}

		createPayload := decodeJSONMap(t, createRec.Body)
		eventID, ok := createPayload["id"].(string)
		if !ok || eventID == "" {
			t.Fatalf("unexpected create payload: %+v", createPayload)
		}

		detailsReq := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID, nil)
		detailsRec := httptest.NewRecorder()
		eventByID.ServeHTTP(detailsRec, detailsReq)
		if detailsRec.Code != http.StatusOK {
			t.Fatalf("details status = %d, want %d body=%s", detailsRec.Code, http.StatusOK, detailsRec.Body.String())
		}
		detailsPayload := decodeJSONMap(t, detailsRec.Body)
		images, ok := detailsPayload["images"].([]any)
		if !ok || len(images) != 2 {
			t.Fatalf("unexpected created images: %+v", detailsPayload)
		}

		patchBody, patchContentType := mustMultipartEventBody(t, map[string]string{
			"title":       "Updated Event",
			"sourceUrl":   "",
			"categoryIds": `[]`,
		}, "images", []namedFile{{name: "updated.png", content: tinyPNG()}})
		patchReq := httptest.NewRequest(http.MethodPatch, "/api/events/"+eventID, patchBody)
		patchReq.Header.Set("Content-Type", patchContentType)
		patchReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: "test-csrf"})
		patchReq.Header.Set(httpx.CSRFHeader, "test-csrf")
		patchReq = patchReq.WithContext(context.WithValue(patchReq.Context(), httpx.UserIDContextKey, "user-1"))
		patchRec := httptest.NewRecorder()
		platformmiddleware.CSRFMiddleware(eventByID, func(r *http.Request) string {
			c, err := r.Cookie("csrf_token")
			if err != nil {
				return ""
			}
			return c.Value
		}).ServeHTTP(patchRec, patchReq)
		if patchRec.Code != http.StatusOK {
			t.Fatalf("patch status = %d, want %d body=%s", patchRec.Code, http.StatusOK, patchRec.Body.String())
		}
		patchPayload := decodeJSONMap(t, patchRec.Body)
		updatedImages, ok := patchPayload["images"].([]any)
		if !ok || len(updatedImages) != 1 {
			t.Fatalf("unexpected patched images: %+v", patchPayload)
		}
		entries, err := os.ReadDir(imageDir)
		if err != nil {
			t.Fatalf("read image dir: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("stored image files = %d, want 2", len(entries))
		}
		if _, err := os.ReadFile(filepath.Join(imageDir, entries[0].Name())); err != nil {
			t.Fatalf("read stored file: %v", err)
		}

		deleteReq := httptest.NewRequest(http.MethodDelete, "/api/events/"+eventID, nil)
		deleteReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: "test-csrf"})
		deleteReq.Header.Set(httpx.CSRFHeader, "test-csrf")
		deleteReq = deleteReq.WithContext(context.WithValue(deleteReq.Context(), httpx.UserIDContextKey, "user-1"))
		deleteRec := httptest.NewRecorder()
		platformmiddleware.CSRFMiddleware(eventByID, func(r *http.Request) string {
			c, err := r.Cookie("csrf_token")
			if err != nil {
				return ""
			}
			return c.Value
		}).ServeHTTP(deleteRec, deleteReq)
		if deleteRec.Code != http.StatusOK {
			t.Fatalf("delete status = %d, want %d body=%s", deleteRec.Code, http.StatusOK, deleteRec.Body.String())
		}
	})

	t.Run("details escape html", func(t *testing.T) {
		events := http.HandlerFunc(handler.Events)
		eventByID := http.HandlerFunc(handler.EventByID)

		createReq := httptest.NewRequest(http.MethodPost, "/api/events", mustJSONBody(t, map[string]any{
			"title":            `<script>alert(1)</script>`,
			"shortDescription": `<b>bold</b>`,
			"fullDescription":  `<img src=x onerror=alert(1)>`,
			"categoryIds":      []string{"music"},
			"sessions": []map[string]any{
				{
					"placeId": "place-1",
					"startAt": "2026-04-20T19:00:00Z",
					"endAt":   "2026-04-20T21:00:00Z",
					"price":   1200,
				},
			},
		}))
		createReq.AddCookie(&http.Cookie{Name: "csrf_token", Value: "test-csrf"})
		createReq.Header.Set(httpx.CSRFHeader, "test-csrf")
		createReq = createReq.WithContext(context.WithValue(createReq.Context(), httpx.UserIDContextKey, "user-1"))
		createRec := httptest.NewRecorder()
		platformmiddleware.CSRFMiddleware(events, func(r *http.Request) string {
			c, err := r.Cookie("csrf_token")
			if err != nil {
				return ""
			}
			return c.Value
		}).ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("create status = %d, want %d body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
		}

		eventID := decodeJSONMap(t, createRec.Body)["id"].(string)
		detailsReq := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID, nil)
		detailsRec := httptest.NewRecorder()
		eventByID.ServeHTTP(detailsRec, detailsReq)
		if detailsRec.Code != http.StatusOK {
			t.Fatalf("details status = %d, want %d body=%s", detailsRec.Code, http.StatusOK, detailsRec.Body.String())
		}
		payload := decodeJSONMap(t, detailsRec.Body)
		if payload["title"] != "&lt;script&gt;alert(1)&lt;/script&gt;" {
			t.Fatalf("title not escaped: %+v", payload)
		}
		if payload["shortDescription"] != "&lt;b&gt;bold&lt;/b&gt;" {
			t.Fatalf("shortDescription not escaped: %+v", payload)
		}
	})
}

func TestHomeHandlerReturnsHomePayload(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(handler.Home).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("home status = %d, want %d", rec.Code, http.StatusOK)
	}

	payload := decodeJSONMap(t, rec.Body)
	if _, ok := payload["featuredEvents"].([]any); !ok {
		t.Fatalf("home payload missing featuredEvents: %+v", payload)
	}
	if _, ok := payload["categories"].([]any); !ok {
		t.Fatalf("home payload missing categories: %+v", payload)
	}
	if _, ok := payload["collections"].([]any); !ok {
		t.Fatalf("home payload missing collections: %+v", payload)
	}
}

func TestCategoriesHandlerReturnsCategories(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(handler.Categories).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("categories status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	payload := decodeJSONMap(t, rec.Body)
	items, ok := payload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("categories payload missing items: %+v", payload)
	}

	badReq := httptest.NewRequest(http.MethodPost, "/api/categories", nil)
	badRec := httptest.NewRecorder()
	http.HandlerFunc(handler.Categories).ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("bad method status = %d, want %d", badRec.Code, http.StatusMethodNotAllowed)
	}
}

func TestTagsHandlerReturnsTags(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(handler.Tags).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("tags status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	payload := decodeJSONMap(t, rec.Body)
	items, ok := payload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("tags payload missing items: %+v", payload)
	}

	badReq := httptest.NewRequest(http.MethodPost, "/api/tags", nil)
	badRec := httptest.NewRecorder()
	http.HandlerFunc(handler.Tags).ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("bad method status = %d, want %d", badRec.Code, http.StatusMethodNotAllowed)
	}
}

func TestCollectionsHandlers(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo), nil)

	t.Run("list collections", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/collections", nil)
		rec := httptest.NewRecorder()
		http.HandlerFunc(handler.Collections).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("collections status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		items, ok := payload["items"].([]any)
		if !ok || len(items) == 0 {
			t.Fatalf("collections payload missing items: %+v", payload)
		}

		badReq := httptest.NewRequest(http.MethodPost, "/api/collections", nil)
		badRec := httptest.NewRecorder()
		http.HandlerFunc(handler.Collections).ServeHTTP(badRec, badReq)
		if badRec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("bad method status = %d, want %d", badRec.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("collection details", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/collections/weekend-picks", nil)
		rec := httptest.NewRecorder()
		http.HandlerFunc(handler.CollectionByID).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("collection details status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}

		payload := decodeJSONMap(t, rec.Body)
		if payload["id"] != "weekend-picks" {
			t.Fatalf("unexpected collection details response: %+v", payload)
		}
		if _, ok := payload["events"].([]any); !ok {
			t.Fatalf("missing events in collection details: %+v", payload)
		}

		missingReq := httptest.NewRequest(http.MethodGet, "/api/collections/unknown", nil)
		missingRec := httptest.NewRecorder()
		http.HandlerFunc(handler.CollectionByID).ServeHTTP(missingRec, missingReq)
		if missingRec.Code != http.StatusNotFound {
			t.Fatalf("collection missing status = %d, want %d", missingRec.Code, http.StatusNotFound)
		}
	})
}

func TestSearchSuggestionsHandler(t *testing.T) {
	repo := placerepo.NewInMemoryRepository(placerepo.SeedPlaces())
	handler := placedelivery.NewHandler(placeusecase.NewService(repo), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/search?query=рок&limit=5", nil)
	rec := httptest.NewRecorder()
	http.HandlerFunc(handler.Search).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("search status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	payload := decodeJSONMap(t, rec.Body)
	items, ok := payload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("search payload missing items: %+v", payload)
	}

	badReq := httptest.NewRequest(http.MethodGet, "/api/search?query=r", nil)
	badRec := httptest.NewRecorder()
	http.HandlerFunc(handler.Search).ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("bad query status = %d, want %d body=%s", badRec.Code, http.StatusBadRequest, badRec.Body.String())
	}

	badLimitReq := httptest.NewRequest(http.MethodGet, "/api/search?query=рок&limit=11", nil)
	badLimitRec := httptest.NewRecorder()
	http.HandlerFunc(handler.Search).ServeHTTP(badLimitRec, badLimitReq)
	if badLimitRec.Code != http.StatusBadRequest {
		t.Fatalf("bad limit status = %d, want %d body=%s", badLimitRec.Code, http.StatusBadRequest, badLimitRec.Body.String())
	}
}

type namedFile struct {
	name    string
	content []byte
}

func newTestEventImageStorage(t *testing.T) (*media.LocalStorage, string) {
	t.Helper()
	dir := t.TempDir()
	storage, err := media.NewLocalStorage(dir, "/uploads/events")
	if err != nil {
		t.Fatalf("create event image storage: %v", err)
	}
	return storage, dir
}

func mustMultipartEventBody(t *testing.T, fields map[string]string, fieldName string, files []namedFile) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field %s: %v", key, err)
		}
	}

	for _, file := range files {
		part, err := writer.CreateFormFile(fieldName, file.name)
		if err != nil {
			t.Fatalf("create file %s: %v", file.name, err)
		}
		if _, err := part.Write(file.content); err != nil {
			t.Fatalf("write file %s: %v", file.name, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}

func tinyPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
		0x54, 0x08, 0x99, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x03, 0x01, 0x01, 0x00, 0xc9, 0xfe, 0x92,
		0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
