package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"cityhawk/backend/internal/platform/httpx"
)

type AppHandler func(http.ResponseWriter, *http.Request) error

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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, "+httpx.RequestIDHeader+", "+httpx.CSRFHeader)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Expose-Headers", httpx.RequestIDHeader+", "+httpx.CSRFHeader)
			w.Header().Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CSRFMiddleware(next http.Handler, readCSRFToken func(*http.Request) string) http.Handler {
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
		if !isUnsafeMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" && !isAllowedOrigin(origin, allowedOrigins) {
			httpx.WriteJSON(w, http.StatusForbidden, httpx.NewErrorResponse("Invalid origin", nil))
			return
		}

		cookieToken := strings.TrimSpace(readCSRFToken(r))
		headerToken := strings.TrimSpace(r.Header.Get(httpx.CSRFHeader))
		if cookieToken == "" || headerToken == "" || cookieToken != headerToken {
			httpx.WriteJSON(w, http.StatusForbidden, httpx.NewErrorResponse("CSRF token mismatch", nil))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CSRFTokenHeaderMiddleware(next http.Handler, readCSRFToken func(*http.Request) string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := strings.TrimSpace(readCSRFToken(r)); token != "" {
			w.Header().Set(httpx.CSRFHeader, token)
		}
		next.ServeHTTP(w, r)
	})
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
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

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get(httpx.RequestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}

		w.Header().Set(httpx.RequestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), httpx.RequestIDContextKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(lrw, r)

		requestID, _ := r.Context().Value(httpx.RequestIDContextKey).(string)
		log.Printf(
			"access request_id=%s method=%s path=%s status=%d bytes=%d duration=%s remote_addr=%s",
			requestID,
			r.Method,
			r.URL.Path,
			lrw.status,
			lrw.bytes,
			time.Since(start).Round(time.Millisecond),
			requestRemoteAddr(r),
		)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec != nil {
				requestID, _ := r.Context().Value(httpx.RequestIDContextKey).(string)
				log.Printf("panic recovered: %v, request_id=%s, path=%s, method=%s\n%s", rec, requestID, r.URL.Path, r.Method, debug.Stack())
				httpx.WriteJSON(w, http.StatusInternalServerError, httpx.NewErrorResponse("internal server error", nil))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func ErrorMiddleware(next AppHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			httpx.WriteMappedError(w, err)
		}
	}
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
			return httpx.NewHTTPError(http.StatusUnauthorized, "missing access token")
		}

		subject, tokenType, err := parseAccessToken(token)
		if err != nil || tokenType != "access" {
			return httpx.NewHTTPError(http.StatusUnauthorized, "invalid access token")
		}
		if strings.TrimSpace(subject) == "" {
			return httpx.NewHTTPError(http.StatusUnauthorized, "invalid access token")
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, subject)
		next.ServeHTTP(w, r.WithContext(ctx))
		return nil
	})
}

func OptionalAuthMiddleware(
	next http.Handler,
	readAccessToken func(*http.Request) string,
	parseAccessToken func(string) (subject string, tokenType string, err error),
	userIDContextKey any,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := readAccessToken(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		subject, tokenType, err := parseAccessToken(token)
		if err != nil || tokenType != "access" || strings.TrimSpace(subject) == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req-fallback"
	}

	var dst [32]byte
	hex.Encode(dst[:], b[:])
	return "req-" + string(dst[:])
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		w.ResponseWriter.WriteHeader(status)
		return
	}

	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func requestRemoteAddr(r *http.Request) string {
	if r == nil {
		return ""
	}

	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
