package middleware

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"cityhawk/backend/internal/platform/httpx"
)

func TestCorsMiddlewareAndHelpers(t *testing.T) {
	t.Setenv("FRONTEND_ORIGIN", "http://frontend.local")

	called := false
	handler := CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://frontend.local")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://frontend.local" {
		t.Fatalf("unexpected allow origin header: %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}

	req = httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://frontend.local")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	if !isUnsafeMethod(http.MethodPost) || isUnsafeMethod(http.MethodGet) {
		t.Fatal("unexpected unsafe method detection")
	}
	if !isAllowedOrigin("http://frontend.local", parseAllowedOrigins("http://frontend.local, http://other")) {
		t.Fatal("expected allowed origin")
	}
	if isAllowedOrigin("", []string{"x"}) {
		t.Fatal("empty origin should not be allowed")
	}
}

func TestCSRFMiddlewareTokenHeaderRequestIDAndAuth(t *testing.T) {
	t.Setenv("FRONTEND_ORIGIN", "http://frontend.local")

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, _ := r.Context().Value(httpx.UserIDContextKey).(string); got != "user-1" {
			t.Fatalf("user id in context = %q, want user-1", got)
		}
		w.WriteHeader(http.StatusOK)
	})

	authHandler := AuthMiddleware(
		base,
		func(*http.Request) string { return "token" },
		func(token string) (string, string, error) { return "user-1", "access", nil },
		httpx.UserIDContextKey,
	)
	csrfHandler := CSRFMiddleware(authHandler, func(*http.Request) string { return "csrf" })
	reqIDHandler := RequestIDMiddleware(CSRFTokenHeaderMiddleware(csrfHandler, func(*http.Request) string { return "csrf" }))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "http://frontend.local")
	req.Header.Set(httpx.CSRFHeader, "csrf")
	rec := httptest.NewRecorder()
	reqIDHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if rec.Header().Get(httpx.RequestIDHeader) == "" || rec.Header().Get(httpx.CSRFHeader) != "csrf" {
		t.Fatalf("missing propagated headers: %+v", rec.Header())
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "http://bad.local")
	req.Header.Set(httpx.CSRFHeader, "csrf")
	rec = httptest.NewRecorder()
	CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), func(*http.Request) string { return "csrf" }).ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestOptionalAuthRecoveryErrorAndLoggingHelpers(t *testing.T) {
	next := OptionalAuthMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got, _ := r.Context().Value(httpx.UserIDContextKey).(string); got != "user-1" {
				t.Fatalf("user id in context = %q, want user-1", got)
			}
			w.WriteHeader(http.StatusAccepted)
		}),
		func(*http.Request) string { return "token" },
		func(string) (string, string, error) { return "user-1", "access", nil },
		httpx.UserIDContextKey,
	)
	rec := httptest.NewRecorder()
	next.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	errHandler := ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		return httpx.NewHTTPError(http.StatusBadRequest, "bad request")
	})
	rec = httptest.NewRecorder()
	errHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	recovery := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	rec = httptest.NewRecorder()
	recovery.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	lrw := &loggingResponseWriter{ResponseWriter: httptest.NewRecorder(), status: http.StatusOK}
	lrw.WriteHeader(http.StatusCreated)
	if _, err := lrw.Write([]byte("ok")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if lrw.status != http.StatusCreated || lrw.bytes != 2 {
		t.Fatalf("unexpected loggingResponseWriter: %+v", lrw)
	}

	if got := requestRemoteAddr(nil); got != "" {
		t.Fatalf("requestRemoteAddr(nil) = %q", got)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	if got := requestRemoteAddr(req); got == "" {
		t.Fatal("expected remote addr")
	}

	if got := newRequestID(); got == "" {
		t.Fatal("expected request id")
	}

	auth := AuthMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		func(*http.Request) string { return "" },
		func(string) (string, string, error) { return "", "", errors.New("bad") },
		httpx.UserIDContextKey,
	)
	rec = httptest.NewRecorder()
	auth.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAccessLogMiddlewareAndRemoteAddrVariants(t *testing.T) {
	originalOutput := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(originalOutput) })

	handler := AccessLogMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/events", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10, 10.0.0.2")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated || rec.Body.String() != "created" {
		t.Fatalf("unexpected response: status=%d body=%q", rec.Code, rec.Body.String())
	}
	if got := requestRemoteAddr(req); got != "203.0.113.10" {
		t.Fatalf("requestRemoteAddr forwarded = %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.7")
	if got := requestRemoteAddr(req); got != "198.51.100.7" {
		t.Fatalf("requestRemoteAddr real ip = %q", got)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "malformed-addr"
	if got := requestRemoteAddr(req); got != "malformed-addr" {
		t.Fatalf("requestRemoteAddr fallback = %q", got)
	}
}
