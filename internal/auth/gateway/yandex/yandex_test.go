package yandex

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	appconfig "cityhawk/backend/internal/config"
	platformerrors "cityhawk/backend/internal/platform/errors"
)

type yandexRoundTrip func(*http.Request) (*http.Response, error)

func (f yandexRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestNewOAuthConfigAndFetchProfile(t *testing.T) {
	if _, err := NewOAuthConfig(appconfig.OAuthProviderConfig{}); err == nil {
		t.Fatal("expected missing config error")
	}

	cfg, err := NewOAuthConfig(appconfig.OAuthProviderConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
	})
	if err != nil || cfg.Endpoint.AuthURL == "" {
		t.Fatalf("NewOAuthConfig() error = %v cfg=%+v", err, cfg)
	}

	original := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: yandexRoundTrip(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"id":"42","login":"user","default_email":"user@example.com"}`)),
			Header:     make(http.Header),
		}, nil
	})}
	defer func() { http.DefaultClient = original }()

	profile, err := FetchProfile(context.Background(), "token")
	if err != nil || profile.ID != "42" {
		t.Fatalf("FetchProfile() error = %v profile=%+v", err, profile)
	}
}

func TestGatewayFetchIdentityAndProfileErrors(t *testing.T) {
	gateway := NewGateway(nil)
	if _, err := gateway.FetchIdentity(context.Background(), ""); err != platformerrors.ErrInvalidOAuthCode {
		t.Fatalf("error = %v, want ErrInvalidOAuthCode", err)
	}
	if _, err := FetchProfile(context.Background(), ""); err == nil {
		t.Fatal("expected empty token error")
	}
}
