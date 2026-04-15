package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gatewaygoogle "cityhawk/backend/internal/auth/gateway/google"
	authmodel "cityhawk/backend/internal/auth/model"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"golang.org/x/oauth2"
)

type stubOAuthUsecase struct {
	loginWithGoogle func(context.Context, string) (authmodel.TokenPair, error)
	loginWithYandex func(context.Context, string) (authmodel.TokenPair, error)
	loginWithVK     func(context.Context, string) (authmodel.TokenPair, error)
}

func (s stubOAuthUsecase) LoginWithGoogle(ctx context.Context, code string) (authmodel.TokenPair, error) {
	return s.loginWithGoogle(ctx, code)
}
func (s stubOAuthUsecase) LoginWithYandex(ctx context.Context, code string) (authmodel.TokenPair, error) {
	return s.loginWithYandex(ctx, code)
}
func (s stubOAuthUsecase) LoginWithVK(ctx context.Context, code string) (authmodel.TokenPair, error) {
	return s.loginWithVK(ctx, code)
}

func TestGoogleLoginHandlerRedirects(t *testing.T) {
	cfg := &oauth2.Config{
		ClientID:    "client",
		RedirectURL: "http://localhost/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL: "https://accounts.example.com/auth",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/login", nil)
	rec := httptest.NewRecorder()
	GoogleLoginHandler(cfg).ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if location := rec.Header().Get("Location"); location == "" {
		t.Fatal("missing redirect location")
	}
	if len(rec.Result().Cookies()) == 0 || rec.Result().Cookies()[0].Name != gatewaygoogle.StateCookieName {
		t.Fatalf("unexpected cookies: %+v", rec.Result().Cookies())
	}
}

func TestGoogleCallbackHandlerSuccessAndErrors(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		state := "state-1"
		req := httptest.NewRequest(http.MethodGet, "/callback?state="+state+"&code=code-1", nil)
		req.AddCookie(&http.Cookie{Name: gatewaygoogle.StateCookieName, Value: state})
		rec := httptest.NewRecorder()

		uc := stubOAuthUsecase{
			loginWithGoogle: func(context.Context, string) (authmodel.TokenPair, error) {
				return authmodel.TokenPair{AccessToken: "access", RefreshToken: "refresh"}, nil
			},
		}
		GoogleCallbackHandler(uc, time.Minute, time.Hour).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		if rec.Header().Get("X-CSRF-Token") == "" {
			t.Fatal("missing csrf header")
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/callback?state=bad&code=code-1", nil)
		rec := httptest.NewRecorder()
		GoogleCallbackHandler(stubOAuthUsecase{}, time.Minute, time.Hour).ServeHTTP(rec, req)
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
				req.AddCookie(&http.Cookie{Name: gatewaygoogle.StateCookieName, Value: state})
				rec := httptest.NewRecorder()

				uc := stubOAuthUsecase{
					loginWithGoogle: func(context.Context, string) (authmodel.TokenPair, error) {
						return authmodel.TokenPair{}, tc.err
					},
				}
				GoogleCallbackHandler(uc, time.Minute, time.Hour).ServeHTTP(rec, req)
				if rec.Code != tc.want {
					t.Fatalf("status = %d, want %d", rec.Code, tc.want)
				}
			})
		}
	})
}
