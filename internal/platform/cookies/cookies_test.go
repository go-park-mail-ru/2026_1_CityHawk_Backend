package cookies

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetReadAndClear(t *testing.T) {
	rec := httptest.NewRecorder()
	Set(rec, "token", "value", Options{TTL: time.Minute})
	resp := rec.Result()
	if len(resp.Cookies()) != 1 {
		t.Fatalf("cookies = %d, want 1", len(resp.Cookies()))
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(resp.Cookies()[0])
	if got := Read(req, "token"); got != "value" {
		t.Fatalf("Read() = %q, want value", got)
	}

	rec = httptest.NewRecorder()
	Clear(rec, "token", Options{})
	resp = rec.Result()
	if resp.Cookies()[0].MaxAge != -1 {
		t.Fatalf("MaxAge = %d, want -1", resp.Cookies()[0].MaxAge)
	}
}

func TestSetDefaultsPathAndReadMissing(t *testing.T) {
	rec := httptest.NewRecorder()
	Set(rec, "token", "value", Options{})
	if got := rec.Result().Cookies()[0].Path; got != "/" {
		t.Fatalf("Path = %q, want /", got)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := Read(req, "missing"); got != "" {
		t.Fatalf("Read() = %q, want empty string", got)
	}
}
