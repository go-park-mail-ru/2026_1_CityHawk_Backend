package config

import (
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	Server ServerConfig
	Auth   AuthConfig
	OAuth  OAuthConfig
}

type ServerConfig struct {
	Port string
}

type AuthConfig struct {
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type OAuthConfig struct {
	Google OAuthProviderConfig
	VK     OAuthProviderConfig
	Yandex OAuthProviderConfig
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func LoadFromEnv() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-insecure-secret-change-me"
		log.Println("JWT_SECRET is empty, using insecure development secret")
	}

	return Config{
		Server: ServerConfig{
			Port: port,
		},
		Auth: AuthConfig{
			JWTSecret:  secret,
			AccessTTL:  parseDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTTL: parseDurationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		},
		OAuth: OAuthConfig{
			Google: OAuthProviderConfig{
				ClientID:     strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID")),
				ClientSecret: strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")),
				RedirectURL:  strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_REDIRECT_URL")),
			},
			VK: OAuthProviderConfig{
				ClientID:     strings.TrimSpace(os.Getenv("VK_OAUTH_CLIENT_ID")),
				ClientSecret: strings.TrimSpace(os.Getenv("VK_OAUTH_CLIENT_SECRET")),
				RedirectURL:  strings.TrimSpace(os.Getenv("VK_OAUTH_REDIRECT_URL")),
			},
			Yandex: OAuthProviderConfig{
				ClientID:     strings.TrimSpace(os.Getenv("YANDEX_OAUTH_CLIENT_ID")),
				ClientSecret: strings.TrimSpace(os.Getenv("YANDEX_OAUTH_CLIENT_SECRET")),
				RedirectURL:  strings.TrimSpace(os.Getenv("YANDEX_OAUTH_REDIRECT_URL")),
			},
		},
	}
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		log.Printf("invalid %s=%q, using fallback %s", key, raw, fallback)
		return fallback
	}

	return d
}
