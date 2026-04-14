package app

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	authdelivery "cityhawk/backend/internal/auth/delivery/http"
	authrepo "cityhawk/backend/internal/auth/repository"
	authusecase "cityhawk/backend/internal/auth/usecase"
	authvalidation "cityhawk/backend/internal/auth/validation"
	"cityhawk/backend/internal/platform/httpx"
	"cityhawk/backend/internal/platform/media"
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

func newTestAvatarStorage(t *testing.T) (*media.LocalStorage, string) {
	t.Helper()
	dir := t.TempDir()
	storage, err := media.NewLocalStorage(dir, "/uploads/avatars")
	if err != nil {
		t.Fatalf("create avatar storage: %v", err)
	}
	return storage, dir
}

func mustMultipartBody(t *testing.T, fields map[string]string, fieldName, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write multipart field %s: %v", key, err)
		}
	}

	if fieldName != "" {
		part, err := writer.CreateFormFile(fieldName, filename)
		if err != nil {
			t.Fatalf("create multipart file: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("write multipart file: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}

func TestRegisterLoginRefreshLogoutFlow(t *testing.T) {
	deps := newTestAuthDeps(t)
	authFlowHandler := authdelivery.NewAuthHandler(deps.authFlowUC, deps.accessTTL, deps.refreshTTL)
	authRefreshHandler := authdelivery.NewRefreshHandler(deps.authUsecase, deps.accessTTL, deps.refreshTTL)
	avatarStore, avatarDir := newTestAvatarStorage(t)
	meHandler := userdelivery.NewMeHandler(deps.store, avatarStore)

	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", mustJSONBody(t, map[string]any{
		"email":       "Tester@example.com ",
		"password":    "verysecret",
		"username":    "тест_user-1",
		"userSurname": "Иванова",
		"birthday":    "2004-01-12",
		"cityId":      "11111111-1111-1111-1111-111111111111",
	}))
	registerRec := httptest.NewRecorder()
	http.HandlerFunc(authFlowHandler.Register).ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d, body=%s", registerRec.Code, http.StatusCreated, registerRec.Body.String())
	}

	registerPayload := decodeJSONMap(t, registerRec.Body)
	if registerPayload["email"] != "tester@example.com" || registerPayload["userSurname"] != "Иванова" {
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
	if mePayload["userSurname"] != "Иванова" {
		t.Fatalf("unexpected me surname: %+v", mePayload)
	}
	if mePayload["birthday"] != "2004-01-12" {
		t.Fatalf("unexpected me birthday: %+v", mePayload)
	}
	city, ok := mePayload["city"].(map[string]any)
	if !ok || city["id"] != "11111111-1111-1111-1111-111111111111" || city["name"] != "Moscow" {
		t.Fatalf("unexpected me city: %+v", mePayload)
	}

	patchBody, contentType := mustMultipartBody(t, map[string]string{
		"username":    "patched_user",
		"userSurname": "Петрова",
		"birthday":    "2005-02-13",
	}, "avatar", "avatar.png", []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41,
		0x54, 0x08, 0x99, 0x63, 0xf8, 0xcf, 0xc0, 0x00,
		0x00, 0x03, 0x01, 0x01, 0x00, 0xc9, 0xfe, 0x92,
		0xef, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e,
		0x44, 0xae, 0x42, 0x60, 0x82,
	})
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/me", patchBody)
	patchReq.Header.Set("Content-Type", contentType)
	patchReq.AddCookie(accessCookie)
	patchRec := httptest.NewRecorder()
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
	).ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("patch me status = %d, want %d, body=%s", patchRec.Code, http.StatusOK, patchRec.Body.String())
	}

	patchPayload := decodeJSONMap(t, patchRec.Body)
	if patchPayload["username"] != "patched_user" || patchPayload["userSurname"] != "Петрова" {
		t.Fatalf("unexpected patch response: %+v", patchPayload)
	}
	avatarURL, ok := patchPayload["avatarUrl"].(string)
	if !ok || !strings.HasPrefix(avatarURL, "http://example.com/uploads/avatars/") {
		t.Fatalf("unexpected patch fields: %+v", patchPayload)
	}
	if patchPayload["birthday"] != "2005-02-13" {
		t.Fatalf("unexpected patch birthday: %+v", patchPayload)
	}
	entries, err := os.ReadDir(avatarDir)
	if err != nil {
		t.Fatalf("read avatar dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("avatar files = %d, want 1", len(entries))
	}
	storedAvatar, err := os.ReadFile(filepath.Join(avatarDir, entries[0].Name()))
	if err != nil {
		t.Fatalf("read stored avatar: %v", err)
	}
	if len(storedAvatar) == 0 {
		t.Fatal("stored avatar is empty")
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	refreshReq.AddCookie(refreshCookie)
	refreshRec := httptest.NewRecorder()
	http.HandlerFunc(authRefreshHandler.Refresh).ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d, body=%s", refreshRec.Code, http.StatusOK, refreshRec.Body.String())
	}
	refreshPayload := decodeJSONMap(t, refreshRec.Body)
	if refreshPayload["ok"] != true {
		t.Fatalf("unexpected refresh response: %+v", refreshPayload)
	}

	logoutCookie := refreshRec.Result().Cookies()[0]
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(logoutCookie)
	logoutRec := httptest.NewRecorder()
	http.HandlerFunc(authRefreshHandler.Logout).ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want %d, body=%s", logoutRec.Code, http.StatusOK, logoutRec.Body.String())
	}
	logoutPayload := decodeJSONMap(t, logoutRec.Body)
	if logoutPayload["ok"] != true {
		t.Fatalf("unexpected logout response: %+v", logoutPayload)
	}

	refreshAfterLogoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	refreshAfterLogoutReq.AddCookie(logoutCookie)
	refreshAfterLogoutRec := httptest.NewRecorder()
	http.HandlerFunc(authRefreshHandler.Refresh).ServeHTTP(refreshAfterLogoutRec, refreshAfterLogoutReq)
	if refreshAfterLogoutRec.Code != http.StatusUnauthorized {
		t.Fatalf("refresh-after-logout status = %d, want %d", refreshAfterLogoutRec.Code, http.StatusUnauthorized)
	}
}

func TestRegisterWithOnlyRequiredFields(t *testing.T) {
	deps := newTestAuthDeps(t)
	authFlowHandler := authdelivery.NewAuthHandler(deps.authFlowUC, deps.accessTTL, deps.refreshTTL)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", mustJSONBody(t, map[string]any{
		"email":       "minimal@example.com",
		"password":    "verysecret",
		"username":    "minimal_user",
		"userSurname": "Смирнова",
	}))
	rec := httptest.NewRecorder()
	http.HandlerFunc(authFlowHandler.Register).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	user, ok := deps.store.GetByEmail(context.Background(), "minimal@example.com")
	if !ok {
		t.Fatal("registered user not found")
	}
	if user.Birthday != nil {
		t.Fatalf("birthday = %v, want nil", user.Birthday)
	}
	if user.CityID != nil {
		t.Fatalf("cityID = %v, want nil", user.CityID)
	}
}

func TestPatchMeValidation(t *testing.T) {
	deps := newTestAuthDeps(t)
	authFlowHandler := authdelivery.NewAuthHandler(deps.authFlowUC, deps.accessTTL, deps.refreshTTL)
	avatarStore, _ := newTestAvatarStorage(t)
	meHandler := userdelivery.NewMeHandler(deps.store, avatarStore)

	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", mustJSONBody(t, map[string]any{
		"email":       "patchme@example.com",
		"password":    "verysecret",
		"username":    "valid_user",
		"userSurname": "Иванова",
		"birthday":    "2004-01-12",
		"cityId":      "11111111-1111-1111-1111-111111111111",
	}))
	registerRec := httptest.NewRecorder()
	http.HandlerFunc(authFlowHandler.Register).ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d, body=%s", registerRec.Code, http.StatusCreated, registerRec.Body.String())
	}

	accessCookie := findCookieByName(registerRec.Result().Cookies(), authdelivery.AccessCookieName)
	if accessCookie == nil {
		t.Fatal("missing access cookie")
	}

	tests := []struct {
		name   string
		body   string
		status int
		errMsg string
	}{
		{
			name:   "unknown field",
			body:   `{"username":"next_user","extra":"value"}`,
			status: http.StatusBadRequest,
			errMsg: "invalid json",
		},
		{
			name:   "bad avatar url",
			body:   `{"avatarUrl":"ftp://example.com/avatar.jpg"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
		{
			name:   "bad birthday",
			body:   `{"birthday":"13-02-2005"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/api/me", strings.NewReader(tc.body))
			req.AddCookie(accessCookie)
			rec := httptest.NewRecorder()

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
			).ServeHTTP(rec, req)

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

func TestPatchMeRejectsUnsupportedAvatar(t *testing.T) {
	deps := newTestAuthDeps(t)
	authFlowHandler := authdelivery.NewAuthHandler(deps.authFlowUC, deps.accessTTL, deps.refreshTTL)
	avatarStore, _ := newTestAvatarStorage(t)
	meHandler := userdelivery.NewMeHandler(deps.store, avatarStore)

	registerReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", mustJSONBody(t, map[string]any{
		"email":       "avatar-invalid@example.com",
		"password":    "verysecret",
		"username":    "valid_user",
		"userSurname": "Иванова",
	}))
	registerRec := httptest.NewRecorder()
	http.HandlerFunc(authFlowHandler.Register).ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d, body=%s", registerRec.Code, http.StatusCreated, registerRec.Body.String())
	}

	accessCookie := findCookieByName(registerRec.Result().Cookies(), authdelivery.AccessCookieName)
	if accessCookie == nil {
		t.Fatal("missing access cookie")
	}

	body, contentType := mustMultipartBody(t, nil, "avatar", "avatar.txt", []byte("plain text"))
	req := httptest.NewRequest(http.MethodPatch, "/api/me", body)
	req.Header.Set("Content-Type", contentType)
	req.AddCookie(accessCookie)
	rec := httptest.NewRecorder()

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
	).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	payload := decodeJSONMap(t, rec.Body)
	if payload["error"] != "Validation failed" {
		t.Fatalf("unexpected error response: %+v", payload)
	}
	details, ok := payload["details"].(map[string]any)
	if !ok || details["avatar"] == nil {
		t.Fatalf("unexpected details: %+v", payload)
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
			body:   `{"email":"a@b.com","password":"verysecret","username":"user","userSurname":"Ivanova","birthday":"2004-01-12","cityId":"11111111-1111-1111-1111-111111111111","role":"admin"}`,
			status: http.StatusBadRequest,
			errMsg: "invalid json",
		},
		{
			name:   "register bad email",
			h:      register,
			body:   `{"email":"wrong","password":"verysecret","username":"user","userSurname":"Ivanova","birthday":"2004-01-12","cityId":"11111111-1111-1111-1111-111111111111"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
		{
			name:   "register short password",
			h:      register,
			body:   `{"email":"a@b.com","password":"short","username":"user","userSurname":"Ivanova","birthday":"2004-01-12","cityId":"11111111-1111-1111-1111-111111111111"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
		{
			name:   "register bad username",
			h:      register,
			body:   `{"email":"a@b.com","password":"verysecret","username":"имя пробел","userSurname":"Ivanova","birthday":"2004-01-12","cityId":"11111111-1111-1111-1111-111111111111"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
		{
			name:   "register bad city id",
			h:      register,
			body:   `{"email":"a@b.com","password":"verysecret","username":"user","userSurname":"Ivanova","birthday":"2004-01-12","cityId":"bad-id"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
		{
			name:   "register bad birthday",
			h:      register,
			body:   `{"email":"a@b.com","password":"verysecret","username":"user","userSurname":"Ivanova","birthday":"12-01-2004","cityId":"11111111-1111-1111-1111-111111111111"}`,
			status: http.StatusBadRequest,
			errMsg: "Validation failed",
		},
		{
			name:   "login invalid credentials",
			h:      login,
			body:   `{"email":"ghost@example.com","password":"verysecret"}`,
			status: http.StatusUnauthorized,
			errMsg: "Invalid credentials",
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
	email, username, userSurname, pass, birthday, cityID, err := authvalidation.ValidateRegister(
		" USER@Example.com ",
		"Юзер_1",
		"Иванова",
		"12345678",
		"2004-01-12",
		"11111111-1111-1111-1111-111111111111",
	)
	if err != nil {
		t.Fatalf("ValidateRegister: %v", err)
	}
	if birthday == nil || cityID == nil {
		t.Fatalf("expected optional values to be filled: birthday=%v cityID=%v", birthday, cityID)
	}
	if email != "user@example.com" || pass != "12345678" || username != "Юзер_1" || userSurname != "Иванова" || *cityID != "11111111-1111-1111-1111-111111111111" || birthday.Format("2006-01-02") != "2004-01-12" {
		t.Fatalf("unexpected normalized values: %q %q %q %q %q %s", email, pass, username, userSurname, *cityID, birthday.Format("2006-01-02"))
	}

	_, _, _, _, emptyBirthday, emptyCityID, err := authvalidation.ValidateRegister(
		" USER@Example.com ",
		"Юзер_1",
		"Иванова",
		"12345678",
		"",
		"",
	)
	if err != nil {
		t.Fatalf("ValidateRegister with optional empty fields: %v", err)
	}
	if emptyBirthday != nil || emptyCityID != nil {
		t.Fatalf("expected nil optional values, got birthday=%v cityID=%v", emptyBirthday, emptyCityID)
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
