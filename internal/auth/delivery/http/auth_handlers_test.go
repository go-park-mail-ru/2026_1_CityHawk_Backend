package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authmodel "cityhawk/backend/internal/auth/model"
	usermodel "cityhawk/backend/internal/user/model"
)

func TestAuthHandlersFlow(t *testing.T) {
	user := usermodel.User{ID: "user-1", Email: "user@example.com", Username: "user", UserSurname: "surname", CreatedAt: time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC)}
	tokens := authmodel.TokenPair{AccessToken: "access", RefreshToken: "refresh", ExpiresIn: 900}
	authHandler := NewAuthHandler(fakeHTTPAuthFlow{user: user, tokens: tokens}, time.Minute, time.Hour)

	rec := performAuthJSON(t, authHandler.Register, registerRequest{Email: user.Email, Username: user.Username, UserSurname: user.UserSurname, Password: "password"})
	if rec.Code != http.StatusCreated || rec.Header().Get("Set-Cookie") == "" {
		t.Fatalf("Register status=%d headers=%v body=%s", rec.Code, rec.Header(), rec.Body.String())
	}

	rec = performAuthJSON(t, authHandler.Login, loginRequest{Email: user.Email, Password: "password"})
	if rec.Code != http.StatusOK || rec.Header().Get(http.CanonicalHeaderKey("X-CSRF-Token")) == "" {
		t.Fatalf("Login status=%d headers=%v body=%s", rec.Code, rec.Header(), rec.Body.String())
	}

	refreshHandler := NewRefreshHandler(fakeHTTPRefresh{tokens: tokens}, time.Minute, time.Hour)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "refresh"})
	rec = httptest.NewRecorder()
	refreshHandler.Refresh(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Refresh status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: RefreshCookieName, Value: "refresh"})
	rec = httptest.NewRecorder()
	refreshHandler.Logout(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Logout status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlersErrors(t *testing.T) {
	authHandler := NewAuthHandler(fakeHTTPAuthFlow{}, time.Minute, time.Hour)
	rec := performAuthRaw(authHandler.Login, `{"email":`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Login invalid json status=%d", rec.Code)
	}
	refreshHandler := NewRefreshHandler(fakeHTTPRefresh{}, time.Minute, time.Hour)
	rec = httptest.NewRecorder()
	refreshHandler.Refresh(rec, httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Refresh missing cookie status=%d", rec.Code)
	}
	rec = httptest.NewRecorder()
	refreshHandler.Logout(rec, httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Logout missing cookie status=%d", rec.Code)
	}
}

func performAuthJSON(t *testing.T, call func(http.ResponseWriter, *http.Request), body any) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return performAuthRaw(call, string(raw))
}

func performAuthRaw(call func(http.ResponseWriter, *http.Request), raw string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString(raw))
	rec := httptest.NewRecorder()
	call(rec, req)
	return rec
}

type fakeHTTPAuthFlow struct {
	user   usermodel.User
	tokens authmodel.TokenPair
}

func (f fakeHTTPAuthFlow) Register(context.Context, authmodel.RegisterInput) (authmodel.RegistrationResult, error) {
	return authmodel.RegistrationResult{User: f.user, Tokens: f.tokens}, nil
}

func (f fakeHTTPAuthFlow) Login(context.Context, authmodel.LoginInput) (authmodel.SessionResult, error) {
	return authmodel.SessionResult{User: f.user, Tokens: f.tokens}, nil
}

type fakeHTTPRefresh struct{ tokens authmodel.TokenPair }

func (f fakeHTTPRefresh) RotateRefresh(context.Context, string) (authmodel.TokenPair, error) {
	return f.tokens, nil
}

func (f fakeHTTPRefresh) RevokeRefresh(context.Context, string) error {
	return nil
}
