package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gatewayyandex "cityhawk/backend/internal/auth/gateway/yandex"
	authmodel "cityhawk/backend/internal/auth/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"golang.org/x/oauth2"
)

type stubOAuthUsecase struct {
	loginWithYandex func(context.Context, string) (authmodel.TokenPair, error)
	loginWithVK     func(context.Context, string) (authmodel.TokenPair, error)
}

func (s stubOAuthUsecase) LoginWithYandex(ctx context.Context, code string) (authmodel.TokenPair, error) {
	return s.loginWithYandex(ctx, code)
}
func (s stubOAuthUsecase) LoginWithVK(ctx context.Context, code string) (authmodel.TokenPair, error) {
	return s.loginWithVK(ctx, code)
}

func TestYandexLoginHandlerRedirects(t *testing.T) {
	cfg := &oauth2.Config{
		ClientID:    "client",
		RedirectURL: "http://localhost/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL: "https://accounts.example.com/auth",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/yandex/login", nil)
	rec := httptest.NewRecorder()
	YandexLoginHandler(cfg).ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if location := rec.Header().Get("Location"); location == "" {
		t.Fatal("missing redirect location")
	}
	if len(rec.Result().Cookies()) == 0 || rec.Result().Cookies()[0].Name != gatewayyandex.StateCookieName {
		t.Fatalf("unexpected cookies: %+v", rec.Result().Cookies())
	}
}

func TestYandexCallbackHandlerSuccessAndErrors(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Setenv("FRONTEND_ORIGIN", "http://frontend.local")
		state := "state-1"
		req := httptest.NewRequest(http.MethodGet, "/callback?state="+state+"&code=code-1", nil)
		req.AddCookie(&http.Cookie{Name: gatewayyandex.StateCookieName, Value: state})
		rec := httptest.NewRecorder()

		uc := stubOAuthUsecase{
			loginWithYandex: func(context.Context, string) (authmodel.TokenPair, error) {
				return authmodel.TokenPair{AccessToken: "access", RefreshToken: "refresh"}, nil
			},
		}
		YandexCallbackHandler(uc, time.Minute, time.Hour).ServeHTTP(rec, req)

		if rec.Code != http.StatusFound {
			t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusFound, rec.Body.String())
		}
		if location := rec.Header().Get("Location"); location != "http://frontend.local/" {
			t.Fatalf("Location = %q, want frontend home", location)
		}
		if rec.Header().Get("X-CSRF-Token") == "" {
			t.Fatal("missing csrf header")
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/callback?state=bad&code=code-1", nil)
		rec := httptest.NewRecorder()
		YandexCallbackHandler(stubOAuthUsecase{}, time.Minute, time.Hour).ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("error mapping", func(t *testing.T) {
		cases := []struct {
			name string
			err  error
			want int
		}{
			{name: "invalid code", err: platformerrors.ErrInvalidOAuthCode, want: http.StatusUnauthorized},
			{name: "profile fetch", err: platformerrors.ErrOAuthProfileFetch, want: http.StatusUnauthorized},
			{name: "issue tokens", err: platformerrors.ErrIssueTokens, want: http.StatusInternalServerError},
			{name: "internal", err: errors.New("boom"), want: http.StatusInternalServerError},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				state := "state-1"
				req := httptest.NewRequest(http.MethodGet, "/callback?state="+state+"&code=code-1", nil)
				req.AddCookie(&http.Cookie{Name: gatewayyandex.StateCookieName, Value: state})
				rec := httptest.NewRecorder()

				uc := stubOAuthUsecase{
					loginWithYandex: func(context.Context, string) (authmodel.TokenPair, error) {
						return authmodel.TokenPair{}, tc.err
					},
				}
				YandexCallbackHandler(uc, time.Minute, time.Hour).ServeHTTP(rec, req)
				if rec.Code != tc.want {
					t.Fatalf("status = %d, want %d", rec.Code, tc.want)
				}
			})
		}
	})
}
