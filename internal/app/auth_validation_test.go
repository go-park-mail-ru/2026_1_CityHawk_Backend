package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authdelivery "cityhawk/backend/internal/auth/delivery/http"
	authrepo "cityhawk/backend/internal/auth/repository"
	authusecase "cityhawk/backend/internal/auth/usecase"
	authvalidation "cityhawk/backend/internal/auth/validation"
	"cityhawk/backend/internal/platform/httpx"
	platformid "cityhawk/backend/internal/platform/id"
	platformmiddleware "cityhawk/backend/internal/platform/middleware"
	platformsecurity "cityhawk/backend/internal/platform/security"
	userdelivery "cityhawk/backend/internal/user/delivery/http"
	userrepo "cityhawk/backend/internal/user/repository"
)

type testAuthDeps struct {
	store        *userrepo.InMemoryUserRepository
	accessTTL    time.Duration
	refreshTTL   time.Duration
	authUsecase  *authusecase.Service
	authFlowUC   *authusecase.AuthFlowService
	tokenService *platformsecurity.JWTTokenService
}

func newTestAuthDeps(t *testing.T) *testAuthDeps {
	t.Helper()
	store := userrepo.NewInMemoryUserRepository()
	accessTTL := time.Minute
	refreshTTL := 2 * time.Minute

	refreshRepo := authrepo.NewInMemoryRefreshRepository()
	tokenService := platformsecurity.NewJWTTokenService([]byte("test-secret"))
	authUC := authusecase.NewService(accessTTL, refreshTTL, store, refreshRepo, tokenService)
	authFlowUC := authusecase.NewAuthFlowService(
		store,
		authUC,
		platformsecurity.NewBcryptPasswordService(),
		platformid.NewUUIDUserIDProvider(),
	)

	return &testAuthDeps{
		store:        store,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
		authUsecase:  authUC,
		authFlowUC:   authFlowUC,
		tokenService: tokenService,
	}
}

func mustJSONBody(t *testing.T, payload any) *bytes.Reader {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return bytes.NewReader(raw)
}

func findCookieByName(cookies []*http.Cookie, name string) *http.Cookie {
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestRegisterLoginRefreshLogoutFlow(t *testing.T) {
	deps := newTestAuthDeps(t)
	authFlowHandler := authdelivery.NewAuthHandler(deps.authFlowUC, deps.accessTTL, deps.refreshTTL)
	authRefreshHandler := authdelivery.NewRefreshHandler(deps.authUsecase, deps.accessTTL, deps.refreshTTL)
	meHandler := userdelivery.NewMeHandler(deps.store)

	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", mustJSONBody(t, map[string]any{
		"email":    "Tester@example.com ",
		"password": "verysecret",
		"username": "тест_user-1",
	}))
	registerRec := httptest.NewRecorder()
	http.HandlerFunc(authFlowHandler.Register).ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d, body=%s", registerRec.Code, http.StatusCreated, registerRec.Body.String())
	}

	registerPayload := decodeJSONMap(t, registerRec.Body)
	if registerPayload["message"] != "registration successful" {
		t.Fatalf("unexpected register response: %+v", registerPayload)
	}

	if _, ok := deps.store.GetByEmail(context.Background(), "tester@example.com"); !ok {
		t.Fatal("registered user not found in store")
	}

	registerCookies := registerRec.Result().Cookies()
	accessCookie := findCookieByName(registerCookies, authdelivery.AccessCookieName)
	refreshCookie := findCookieByName(registerCookies, authdelivery.RefreshCookieName)
	if accessCookie == nil || accessCookie.Value == "" {
		t.Fatalf("access cookie not set: %+v", registerCookies)
	}
	if refreshCookie == nil || refreshCookie.Value == "" {
		t.Fatalf("refresh cookie not set: %+v", registerCookies)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	meReq.AddCookie(accessCookie)
	meRec := httptest.NewRecorder()
	platformmiddleware.AuthMiddleware(
		http.HandlerFunc(meHandler.Me),
		authdelivery.ReadAccessCookie,
		func(token string) (string, string, error) {
			claims, err := deps.tokenService.Parse(token)
			if err != nil {
				return "", "", err
			}
			return claims.Subject, claims.Type, nil
		},
		httpx.UserIDContextKey,
	).ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d, body=%s", meRec.Code, http.StatusOK, meRec.Body.String())
	}

	mePayload := decodeJSONMap(t, meRec.Body)
	username, ok := mePayload["username"].(string)
	if !ok || username != "тест_user-1" {
		t.Fatalf("unexpected username: %+v", mePayload)
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	refreshReq.AddCookie(refreshCookie)
	refreshRec := httptest.NewRecorder()
	http.HandlerFunc(authRefreshHandler.Refresh).ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d, body=%s", refreshRec.Code, http.StatusOK, refreshRec.Body.String())
	}

	logoutCookie := refreshRec.Result().Cookies()[0]
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(logoutCookie)
	logoutRec := httptest.NewRecorder()
	http.HandlerFunc(authRefreshHandler.Logout).ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want %d, body=%s", logoutRec.Code, http.StatusOK, logoutRec.Body.String())
	}

	refreshAfterLogoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	refreshAfterLogoutReq.AddCookie(logoutCookie)
	refreshAfterLogoutRec := httptest.NewRecorder()
	http.HandlerFunc(authRefreshHandler.Refresh).ServeHTTP(refreshAfterLogoutRec, refreshAfterLogoutReq)
	if refreshAfterLogoutRec.Code != http.StatusUnauthorized {
		t.Fatalf("refresh-after-logout status = %d, want %d", refreshAfterLogoutRec.Code, http.StatusUnauthorized)
	}
}

func TestRegisterAndLoginValidation(t *testing.T) {
	deps := newTestAuthDeps(t)
	authFlowHandler := authdelivery.NewAuthHandler(deps.authFlowUC, deps.accessTTL, deps.refreshTTL)
	register := http.HandlerFunc(authFlowHandler.Register)
	login := http.HandlerFunc(authFlowHandler.Login)

	tests := []struct {
		name   string
		h      http.HandlerFunc
		body   string
		status int
		errMsg string
	}{
		{
			name:   "register unknown field",
			h:      register,
			body:   `{"email":"a@b.com","password":"verysecret","username":"user","role":"admin"}`,
			status: http.StatusBadRequest,
			errMsg: "invalid json",
		},
		{
			name:   "register bad email",
			h:      register,
			body:   `{"email":"wrong","password":"verysecret","username":"user"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
		{
			name:   "register short password",
			h:      register,
			body:   `{"email":"a@b.com","password":"short","username":"user"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
		{
			name:   "register bad username",
			h:      register,
			body:   `{"email":"a@b.com","password":"verysecret","username":"имя пробел"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
		{
			name:   "login invalid credentials",
			h:      login,
			body:   `{"email":"ghost@example.com","password":"verysecret"}`,
			status: http.StatusUnauthorized,
			errMsg: "invalid credentials",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			tc.h.ServeHTTP(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, tc.status, rec.Body.String())
			}

			payload := decodeJSONMap(t, rec.Body)
			got, _ := payload["error"].(string)
			if got != tc.errMsg {
				t.Fatalf("error = %q, want %q", got, tc.errMsg)
			}

			details, ok := payload["details"].(map[string]any)
			if !ok {
				t.Fatalf("error response missing details object: %+v", payload)
			}

			if tc.errMsg == "Validation failed" && len(details) == 0 {
				t.Fatalf("validation error missing details: %+v", payload)
			}
		})
	}
}

func TestValidationHelpersAndSecurity(t *testing.T) {
	email, pass, username, err := authvalidation.ValidateRegister(" USER@Example.com ", "12345678", "Юзер_1")
	if err != nil {
		t.Fatalf("ValidateRegister: %v", err)
	}
	if email != "user@example.com" || pass != "12345678" || username != "Юзер_1" {
		t.Fatalf("unexpected normalized values: %q %q %q", email, pass, username)
	}

	passwordService := platformsecurity.NewBcryptPasswordService()
	hashed, err := passwordService.Hash(pass)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if !passwordService.Verify(pass, hashed) {
		t.Fatal("verify password returned false for valid password")
	}
}
