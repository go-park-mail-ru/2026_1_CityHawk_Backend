package main

import (
	"net/http"
	"strings"
	"time"
)

func setOAuthStateCookie(w http.ResponseWriter, state, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().UTC().Add(10 * time.Minute),
		MaxAge:   600,
	})
}

func clearOAuthStateCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func readOAuthStateCookie(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(c.Value)
}
