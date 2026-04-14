package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStateHelpers(t *testing.T) {
	state, err := NewState()
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}
	if state == "" {
		t.Fatal("state is empty")
	}

	rec := httptest.NewRecorder()
	SetStateCookie(rec, "oauth_state", state, time.Minute)
	resp := rec.Result()
	cookies := resp.Cookies()
	if len(cookies) != 1 || cookies[0].Name != "oauth_state" {
		t.Fatalf("unexpected cookies: %+v", cookies)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookies[0])
	if got := ReadStateCookie(req, "oauth_state"); got != state {
		t.Fatalf("ReadStateCookie() = %q, want %q", got, state)
	}

	rec = httptest.NewRecorder()
	ClearStateCookie(rec, "oauth_state")
	resp = rec.Result()
	if got := resp.Cookies()[0].MaxAge; got != -1 {
		t.Fatalf("cleared cookie MaxAge = %d, want -1", got)
	}
}

func TestStateCookieOptions(t *testing.T) {
	opts := stateCookieOptions(time.Minute)
	if opts.Path != "/" || !opts.HttpOnly || opts.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected options: %+v", opts)
	}
}
