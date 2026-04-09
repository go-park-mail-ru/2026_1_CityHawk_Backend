package cookies

import (
	"net/http"
	"strings"
	"time"
)

type Options struct {
	Path     string
	HttpOnly bool
	Secure   bool
	SameSite http.SameSite
	TTL      time.Duration
}

func Set(w http.ResponseWriter, name, value string, opts Options) {
	path := opts.Path
	if strings.TrimSpace(path) == "" {
		path = "/"
	}

	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: opts.HttpOnly,
		Secure:   opts.Secure,
		SameSite: opts.SameSite,
	}

	if opts.TTL > 0 {
		cookie.Expires = time.Now().UTC().Add(opts.TTL)
		cookie.MaxAge = int(opts.TTL.Seconds())
	}

	http.SetCookie(w, cookie)
}

func Clear(w http.ResponseWriter, name string, opts Options) {
	path := opts.Path
	if strings.TrimSpace(path) == "" {
		path = "/"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		HttpOnly: opts.HttpOnly,
		Secure:   opts.Secure,
		SameSite: opts.SameSite,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func Read(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}
