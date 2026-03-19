package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"cityhawk/backend/internal/platform/httpx"
)

type AppHandler func(http.ResponseWriter, *http.Request) error

type HTTPError struct {
	status  int
	message string
}

func (e HTTPError) Error() string {
	return e.message
}

func NewHTTPError(status int, message string) error {
	return HTTPError{status: status, message: message}
}

func CorsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := parseAllowedOrigins(os.Getenv("FRONTEND_ORIGIN"))
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{
			"http://cityhawk.ru",
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedOrigin(origin, allowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseAllowedOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			origins = append(origins, v)
		}
	}
	return origins
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if strings.TrimSpace(origin) == "" {
		return false
	}
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec != nil {
				log.Printf("panic recovered: %v, path=%s, method=%s\n%s", rec, r.URL.Path, r.Method, debug.Stack())
				httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func ErrorMiddleware(next AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			WriteMappedError(w, err)
		}
	}
}

func WriteMappedError(w http.ResponseWriter, err error) {
	var he HTTPError
	if errors.As(err, &he) {
		httpx.WriteJSON(w, he.status, map[string]string{"error": he.message})
		return
	}
	httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

func AuthMiddleware(
	next http.Handler,
	readAccessToken func(*http.Request) string,
	parseAccessToken func(string) (subject string, tokenType string, err error),
	userIDContextKey any,
) http.Handler {
	return ErrorMiddleware(func(w http.ResponseWriter, r *http.Request) error {
		token := readAccessToken(r)
		if token == "" {
			return NewHTTPError(http.StatusUnauthorized, "missing access token")
		}

		subject, tokenType, err := parseAccessToken(token)
		if err != nil || tokenType != "access" {
			return NewHTTPError(http.StatusUnauthorized, "invalid access token")
		}
		if strings.TrimSpace(subject) == "" {
			return NewHTTPError(http.StatusUnauthorized, "invalid access token")
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, subject)
		next.ServeHTTP(w, r.WithContext(ctx))
		return nil
	})
}
