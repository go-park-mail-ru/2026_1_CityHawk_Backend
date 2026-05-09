package config

import (
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	GRPC     GRPCConfig
	Metrics  MetricsConfig
	Auth     AuthConfig
	OAuth    OAuthConfig
	Database DatabaseConfig
	Photon   PhotonConfig
}

type ServerConfig struct {
	Port string
}

type GRPCConfig struct {
	AuthAddr    string
	ProfileAddr string
	EventsAddr  string
	SupportAddr string
	SocialAddr  string
}

type MetricsConfig struct {
	AuthAddr    string
	ProfileAddr string
	EventsAddr  string
	SupportAddr string
	SocialAddr  string
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

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type PhotonConfig struct {
	Enabled         bool
	BaseURL         string
	RequestTimeout  time.Duration
	DefaultCountry  string
	DefaultTimezone string
}

func LoadFromEnv() Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, use system env")
	}

	port := getEnv("PORT", "8080")
	secret := getEnv("JWT_SECRET", "")
	if secret == "" {
		secret = "dev-insecure-secret-change-me"
		log.Println("JWT_SECRET is empty, using insecure development secret")
	}

	return Config{
		Server: ServerConfig{
			Port: port,
		},
		GRPC: GRPCConfig{
			AuthAddr:    getEnv("AUTH_GRPC_ADDR", ":50051"),
			ProfileAddr: getEnv("PROFILE_GRPC_ADDR", ":50052"),
			EventsAddr:  getEnv("EVENTS_GRPC_ADDR", ":50053"),
			SupportAddr: getEnv("SUPPORT_GRPC_ADDR", ":50054"),
			SocialAddr:  getEnv("SOCIAL_GRPC_ADDR", ":50055"),
		},
		Metrics: MetricsConfig{
			AuthAddr:    getEnv("AUTH_METRICS_ADDR", ":9101"),
			ProfileAddr: getEnv("PROFILE_METRICS_ADDR", ":9102"),
			EventsAddr:  getEnv("EVENTS_METRICS_ADDR", ":9103"),
			SupportAddr: getEnv("SUPPORT_METRICS_ADDR", ":9104"),
			SocialAddr:  getEnv("SOCIAL_METRICS_ADDR", ":9105"),
		},
		Auth: AuthConfig{
			JWTSecret:  secret,
			AccessTTL:  parseDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTTL: parseDurationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		},
		OAuth: OAuthConfig{
			Google: OAuthProviderConfig{
				ClientID:     getEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
				RedirectURL:  normalizeOAuthRedirectURL(getEnv("GOOGLE_OAUTH_REDIRECT_URL", "")),
			},
			VK: OAuthProviderConfig{
				ClientID:     getEnv("VK_OAUTH_CLIENT_ID", ""),
				ClientSecret: getEnv("VK_OAUTH_CLIENT_SECRET", ""),
				RedirectURL:  normalizeOAuthRedirectURL(getEnv("VK_OAUTH_REDIRECT_URL", "")),
			},
			Yandex: OAuthProviderConfig{
				ClientID:     getEnv("YANDEX_OAUTH_CLIENT_ID", ""),
				ClientSecret: getEnv("YANDEX_OAUTH_CLIENT_SECRET", ""),
				RedirectURL:  normalizeOAuthRedirectURL(getEnv("YANDEX_OAUTH_REDIRECT_URL", "")),
			},
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "cityhawk"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "cityhawk"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Photon: PhotonConfig{
			Enabled:         parseBoolEnv("PHOTON_ENABLED", false),
			BaseURL:         getEnv("PHOTON_BASE_URL", "http://localhost:2322"),
			RequestTimeout:  parseDurationEnv("PHOTON_REQUEST_TIMEOUT", 5*time.Second),
			DefaultCountry:  getEnv("PHOTON_DEFAULT_COUNTRY", "Russia"),
			DefaultTimezone: getEnv("PHOTON_DEFAULT_TIMEZONE", "UTC"),
		},
	}
}

func normalizeOAuthRedirectURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}

	u, err := url.Parse(value)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return value
	}
	if strings.HasPrefix(u.Path, "/auth/") {
		u.Path = "/api" + u.Path
		u.RawPath = ""
	}
	return u.String()
}

func (c DatabaseConfig) DSN() string {
	query := url.Values{}
	query.Set("sslmode", c.SSLMode)

	dsn := &url.URL{
		Scheme:   "postgres",
		Host:     net.JoinHostPort(c.Host, c.Port),
		Path:     "/" + c.Name,
		RawPath:  "/" + url.PathEscape(c.Name),
		RawQuery: query.Encode(),
	}

	if c.Password != "" {
		dsn.User = url.UserPassword(c.User, c.Password)
	} else if c.User != "" {
		dsn.User = url.User(c.User)
	}

	return dsn.String()
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

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseBoolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		log.Printf("invalid %s=%q, using fallback %t", key, raw, fallback)
		return fallback
	}

	return value
}

func parseIntEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		log.Printf("invalid %s=%q, using fallback %d", key, raw, fallback)
		return fallback
	}

	return value
}
