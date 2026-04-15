package common

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	platformcookies "cityhawk/backend/internal/platform/cookies"
)

func NewState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func SetStateCookie(w http.ResponseWriter, name, state string, ttl time.Duration) {
	platformcookies.Set(w, name, state, stateCookieOptions(ttl))
}

func ClearStateCookie(w http.ResponseWriter, name string) {
	platformcookies.Clear(w, name, stateCookieOptions(0))
}

func ReadStateCookie(r *http.Request, name string) string {
	return platformcookies.Read(r, name)
}

func stateCookieOptions(ttl time.Duration) platformcookies.Options {
	return platformcookies.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		TTL:      ttl,
	}
}
