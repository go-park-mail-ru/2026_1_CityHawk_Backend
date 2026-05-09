package config

import "testing"

func TestLoadFromEnvAndParsers(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("ACCESS_TOKEN_TTL", "20m")
	t.Setenv("REFRESH_TOKEN_TTL", "48h")
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "name")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("AUTH_GRPC_ADDR", ":51051")
	t.Setenv("PROFILE_GRPC_ADDR", ":51052")
	t.Setenv("EVENTS_GRPC_ADDR", ":51053")
	t.Setenv("SUPPORT_GRPC_ADDR", ":51054")
	t.Setenv("SOCIAL_GRPC_ADDR", ":51055")
	t.Setenv("AUTH_METRICS_ADDR", ":52051")
	t.Setenv("PROFILE_METRICS_ADDR", ":52052")
	t.Setenv("EVENTS_METRICS_ADDR", ":52053")
	t.Setenv("SUPPORT_METRICS_ADDR", ":52054")
	t.Setenv("SOCIAL_METRICS_ADDR", ":52055")
	t.Setenv("PHOTON_ENABLED", "true")
	t.Setenv("PHOTON_BASE_URL", "http://photon")
	t.Setenv("PHOTON_REQUEST_TIMEOUT", "7s")
	t.Setenv("PHOTON_DEFAULT_COUNTRY", "Russia")
	t.Setenv("PHOTON_DEFAULT_TIMEZONE", "Europe/Moscow")
	t.Setenv("GOOGLE_OAUTH_REDIRECT_URL", "http://cityhawk.ru:8080/auth/google/callback")
	t.Setenv("VK_OAUTH_REDIRECT_URL", "http://cityhawk.ru:8080/api/auth/vk/callback")
	t.Setenv("YANDEX_OAUTH_REDIRECT_URL", "http://cityhawk.ru:8080/auth/yandex/callback")

	cfg := LoadFromEnv()
	if cfg.Server.Port != "9090" || cfg.Auth.JWTSecret != "secret" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.Database.Host != "db" || cfg.Database.Port != "5433" || cfg.Database.SSLMode != "require" {
		t.Fatalf("unexpected db config: %+v", cfg.Database)
	}
	if cfg.GRPC.AuthAddr != ":51051" ||
		cfg.GRPC.ProfileAddr != ":51052" ||
		cfg.GRPC.EventsAddr != ":51053" ||
		cfg.GRPC.SupportAddr != ":51054" ||
		cfg.GRPC.SocialAddr != ":51055" {
		t.Fatalf("unexpected grpc config: %+v", cfg.GRPC)
	}
	if cfg.Metrics.AuthAddr != ":52051" ||
		cfg.Metrics.ProfileAddr != ":52052" ||
		cfg.Metrics.EventsAddr != ":52053" ||
		cfg.Metrics.SupportAddr != ":52054" ||
		cfg.Metrics.SocialAddr != ":52055" {
		t.Fatalf("unexpected metrics config: %+v", cfg.Metrics)
	}
	if !cfg.Photon.Enabled || cfg.Photon.BaseURL != "http://photon" {
		t.Fatalf("unexpected photon config: %+v", cfg.Photon)
	}
	if cfg.OAuth.Google.RedirectURL != "http://cityhawk.ru:8080/api/auth/google/callback" {
		t.Fatalf("unexpected google redirect url: %q", cfg.OAuth.Google.RedirectURL)
	}
	if cfg.OAuth.VK.RedirectURL != "http://cityhawk.ru:8080/api/auth/vk/callback" {
		t.Fatalf("unexpected vk redirect url: %q", cfg.OAuth.VK.RedirectURL)
	}
	if cfg.OAuth.Yandex.RedirectURL != "http://cityhawk.ru:8080/api/auth/yandex/callback" {
		t.Fatalf("unexpected yandex redirect url: %q", cfg.OAuth.Yandex.RedirectURL)
	}
}
