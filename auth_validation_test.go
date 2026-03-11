package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func newTestAuthDeps(t *testing.T) (*userStore, *authService) {
	t.Helper()
	return &userStore{byEmail: make(map[string]user)}, &authService{
		secret:     []byte("test-secret"),
		accessTTL:  time.Minute,
		refreshTTL: 2 * time.Minute,
		refreshSessions: &refreshStore{
			session: make(map[string]refreshSession),
		},
	}
}

func decodeJSONMap(t *testing.T, body *bytes.Buffer) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}

func mustJSONBody(t *testing.T, payload any) *bytes.Reader {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return bytes.NewReader(raw)
}

func TestRegisterLoginRefreshLogoutFlow(t *testing.T) {
	store, auth := newTestAuthDeps(t)

	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", mustJSONBody(t, map[string]any{
		"email":    "Tester@example.com ",
		"password": "verysecret",
		"username": "тест_user-1",
	}))
	registerRec := httptest.NewRecorder()
	registerHandler(store, auth).ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d, body=%s", registerRec.Code, http.StatusCreated, registerRec.Body.String())
	}

	registerPayload := decodeJSONMap(t, registerRec.Body)
	accessToken, _ := registerPayload["access_token"].(string)
	refreshToken, _ := registerPayload["refresh_token"].(string)
	if accessToken == "" || refreshToken == "" {
		t.Fatalf("missing tokens in register response: %+v", registerPayload)
	}

	if _, ok := store.getByEmail("tester@example.com"); !ok {
		t.Fatal("registered user not found in store")
	}

	meReq := httptest.NewRequest(http.MethodGet, "/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+accessToken)
	meRec := httptest.NewRecorder()
	authMiddleware(auth, meHandler(store)).ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d, body=%s", meRec.Code, http.StatusOK, meRec.Body.String())
	}

	mePayload := decodeJSONMap(t, meRec.Body)
	username, ok := mePayload["username"].(string)
	if !ok || username != "тест_user-1" {
		t.Fatalf("unexpected username: %+v", mePayload)
	}

	refreshCookie := registerRec.Result().Cookies()[0]
	refreshReq := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	refreshReq.AddCookie(refreshCookie)
	refreshRec := httptest.NewRecorder()
	refreshHandler(store, auth).ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d, body=%s", refreshRec.Code, http.StatusOK, refreshRec.Body.String())
	}

	logoutCookie := refreshRec.Result().Cookies()[0]
	logoutReq := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	logoutReq.AddCookie(logoutCookie)
	logoutRec := httptest.NewRecorder()
	logoutHandler(auth).ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want %d, body=%s", logoutRec.Code, http.StatusOK, logoutRec.Body.String())
	}

	refreshAfterLogoutReq := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	refreshAfterLogoutReq.AddCookie(logoutCookie)
	refreshAfterLogoutRec := httptest.NewRecorder()
	refreshHandler(store, auth).ServeHTTP(refreshAfterLogoutRec, refreshAfterLogoutReq)
	if refreshAfterLogoutRec.Code != http.StatusUnauthorized {
		t.Fatalf("refresh-after-logout status = %d, want %d", refreshAfterLogoutRec.Code, http.StatusUnauthorized)
	}
}

func TestRegisterAndLoginValidation(t *testing.T) {
	store, auth := newTestAuthDeps(t)
	register := registerHandler(store, auth)
	login := loginHandler(store, auth)

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
			errMsg: "invalid email format",
		},
		{
			name:   "register short password",
			h:      register,
			body:   `{"email":"a@b.com","password":"short","username":"user"}`,
			status: http.StatusBadRequest,
			errMsg: "password must be at least 8 characters",
		},
		{
			name:   "register bad username",
			h:      register,
			body:   `{"email":"a@b.com","password":"verysecret","username":"имя пробел"}`,
			status: http.StatusBadRequest,
			errMsg: "username contains invalid characters",
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
		})
	}
}

func TestValidationHelpersAndSecurity(t *testing.T) {
	email, pass, username, err := validateRegisterInput(registerRequest{
		Email:    " USER@Example.com ",
		Password: "12345678",
		Username: "Юзер_1",
	})
	if err != nil {
		t.Fatalf("validateRegisterInput: %v", err)
	}
	if !reflect.DeepEqual([]string{email, pass, username}, []string{"user@example.com", "12345678", "Юзер_1"}) {
		t.Fatalf("unexpected normalized values: %q %q %q", email, pass, username)
	}

	hashed, err := hashPassword(pass)
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if !verifyPassword(pass, hashed) {
		t.Fatal("verifyPassword returned false for valid password")
	}
	if verifyPassword("wrong-pass", hashed) {
		t.Fatal("verifyPassword returned true for invalid password")
	}
}
