package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewErrorResponseNormalizesDetails(t *testing.T) {
	t.Run("nil details", func(t *testing.T) {
		resp := NewErrorResponse("boom", nil)
		if resp.Error != "boom" {
			t.Fatalf("Error = %q, want %q", resp.Error, "boom")
		}
		if len(resp.Details) != 0 {
			t.Fatalf("Details = %#v, want empty map", resp.Details)
		}
	})

	t.Run("string map", func(t *testing.T) {
		resp := NewErrorResponse("boom", map[string]string{"field": "required"})
		if got := resp.Details["field"]; got != "required" {
			t.Fatalf("Details[field] = %v, want %q", got, "required")
		}
	})

	t.Run("arbitrary value", func(t *testing.T) {
		resp := NewErrorResponse("boom", 42)
		if got := resp.Details["value"]; got != 42 {
			t.Fatalf("Details[value] = %v, want %d", got, 42)
		}
	})
}

func TestWriteMappedError(t *testing.T) {
	t.Run("http error", func(t *testing.T) {
		rec := httptest.NewRecorder()

		WriteMappedError(rec, NewHTTPErrorWithDetails(http.StatusBadRequest, "invalid input", map[string]string{"field": "email"}))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
			t.Fatalf("Content-Type = %q, want application/json; charset=utf-8", got)
		}
		body := rec.Body.String()
		if !strings.Contains(body, `"error":"invalid input"`) {
			t.Fatalf("body = %s, want error message", body)
		}
		if !strings.Contains(body, `"field":"email"`) {
			t.Fatalf("body = %s, want details", body)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		rec := httptest.NewRecorder()

		WriteMappedError(rec, errors.New("boom"))

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
		}
		if !strings.Contains(rec.Body.String(), `"error":"internal error"`) {
			t.Fatalf("body = %s, want internal error response", rec.Body.String())
		}
	})
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteJSON(rec, http.StatusCreated, map[string]string{"id": "event-1"})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want application/json; charset=utf-8", got)
	}
	if !strings.Contains(rec.Body.String(), `"id":"event-1"`) {
		t.Fatalf("body = %s, want payload", rec.Body.String())
	}
}

func TestOpenAPIYAMLHandler(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)

		OpenAPIYAMLHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/yaml; charset=utf-8" {
			t.Fatalf("Content-Type = %q, want application/yaml; charset=utf-8", got)
		}
		if !strings.Contains(rec.Body.String(), "openapi: 3.0.3") {
			t.Fatalf("body = %s, want OpenAPI spec", rec.Body.String())
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/openapi.yaml", nil)

		OpenAPIYAMLHandler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})
}

func TestSwaggerUIHandler(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/swagger", nil)

		SwaggerUIHandler("/openapi.yaml")(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
			t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", got)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "SwaggerUIBundle") || !strings.Contains(body, "/openapi.yaml") {
			t.Fatalf("body = %s, want swagger html with spec path", body)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/swagger", nil)

		SwaggerUIHandler("/openapi.yaml")(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})
}
