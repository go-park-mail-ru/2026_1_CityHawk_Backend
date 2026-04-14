package vk

import (
	"context"
	"testing"

	appconfig "cityhawk/backend/internal/config"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"golang.org/x/oauth2"
)

func TestVKHelpers(t *testing.T) {
	if _, err := NewOAuthConfig(appconfig.OAuthProviderConfig{}); err == nil {
		t.Fatal("expected missing config error")
	}

	cfg, err := NewOAuthConfig(appconfig.OAuthProviderConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
	})
	if err != nil {
		t.Fatalf("NewOAuthConfig() error = %v", err)
	}
	if got := AuthCodeURL(cfg, "state"); got == "" {
		t.Fatal("expected auth code url")
	}
	if _, err := ExchangeCode(context.Background(), cfg, ""); err == nil {
		t.Fatal("expected empty code error")
	}

	token := &oauth2.Token{}
	token = token.WithExtra(map[string]any{
		"email":   " user@example.com ",
		"user_id": float64(42),
	})
	if got := extractEmail(token); got != "user@example.com" {
		t.Fatalf("extractEmail() = %q", got)
	}
	if got := extractUserID(token); got != "42" {
		t.Fatalf("extractUserID() = %q", got)
	}

	gateway := NewGateway(cfg)
	if _, err := gateway.FetchIdentity(context.Background(), ""); err != platformerrors.ErrInvalidOAuthCode {
		t.Fatalf("error = %v, want ErrInvalidOAuthCode", err)
	}
}
