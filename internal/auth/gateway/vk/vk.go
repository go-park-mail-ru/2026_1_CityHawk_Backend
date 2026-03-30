package vk

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	authmodel "cityhawk/backend/internal/auth/model"
	appconfig "cityhawk/backend/internal/config"
	platformerrors "cityhawk/backend/internal/platform/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

const StateCookieName = "vk_oauth_state"

type Gateway struct {
	cfg *oauth2.Config
}

func NewGateway(cfg *oauth2.Config) *Gateway {
	return &Gateway{cfg: cfg}
}

func NewOAuthConfig(cfg appconfig.OAuthProviderConfig) (*oauth2.Config, error) {
	if cfg.ClientID == "" {
		return nil, errors.New("VK_OAUTH_CLIENT_ID is required")
	}
	if cfg.ClientSecret == "" {
		return nil, errors.New("VK_OAUTH_CLIENT_SECRET is required")
	}
	if cfg.RedirectURL == "" {
		return nil, errors.New("VK_OAUTH_REDIRECT_URL is required")
	}

	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     vk.Endpoint,
		Scopes:       []string{"email"},
	}, nil
}

func AuthCodeURL(cfg *oauth2.Config, state string) string {
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code string) (*oauth2.Token, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("oauth code is required")
	}
	return cfg.Exchange(ctx, code)
}

func (g *Gateway) FetchIdentity(ctx context.Context, code string) (authmodel.OAuthIdentity, error) {
	token, err := ExchangeCode(ctx, g.cfg, code)
	if err != nil {
		return authmodel.OAuthIdentity{}, platformerrors.ErrInvalidOAuthCode
	}

	email := extractEmail(token)
	userID := extractUserID(token)
	username := ""
	if userID != "" {
		username = "vk_" + userID
	}

	return authmodel.OAuthIdentity{
		Provider:  "vk",
		SubjectID: userID,
		Email:     email,
		Username:  username,
	}, nil
}

func extractEmail(token *oauth2.Token) string {
	if token == nil {
		return ""
	}
	email, _ := token.Extra("email").(string)
	return strings.TrimSpace(email)
}

func extractUserID(token *oauth2.Token) string {
	if token == nil {
		return ""
	}

	raw := token.Extra("user_id")
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == math.Trunc(v) {
			return strconv.FormatInt(int64(v), 10)
		}
		return fmt.Sprintf("%.0f", v)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
