// Package xsrf provides methods for generating and validating secure XSRF tokens.
package xsrf

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/xsrftoken"
)

var (
	xsrfKey = &contextKey{"xsrf"}
)

type contextKey struct {
	s string
}

func (k contextKey) String() string {
	return "xsrf: " + k.s
}

// Template is a template helper for html/template that provides an <input> field
// populated with a CSRF token.
//
// Example:
//
//	// The following tag in our index.html template:
//	{{ .xsrf }}
//
//	// ... becomes:
//	<input type="hidden" name="_xsrf" value="<token>">
func Template(ctx context.Context) template.HTML {
	s, ok := ctx.Value(xsrfKey).(string)
	if !ok {
		return ""
	}
	fragment := fmt.Sprintf(`<input type="hidden" name="_xsrf" value="%s">`, s)
	return template.HTML(fragment)
}

// Generate
func Generate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("_xsrf")
		if err != nil {
			c = &http.Cookie{
				Name:     "_xsrf",
				Value:    xsrftoken.Generate("hmac_value", "", ""),
				Path:     "/",
				Expires:  time.Now().Add(xsrftoken.Timeout),
				MaxAge:   int(xsrftoken.Timeout.Seconds()),
				HttpOnly: true,
				Secure:   r.URL.Scheme == "https",
				// https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#samesite-cookie-attribute
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, c)
		}

		r = r.WithContext(context.WithValue(r.Context(), xsrfKey, c.Value))
		h.ServeHTTP(w, r)
	})
}

// ValidFunc
func ValidFunc(h http.HandlerFunc) http.HandlerFunc {
	var (
		serve400 = serveErr(http.StatusBadRequest)
		serve401 = serveErr(http.StatusUnauthorized)
		serve403 = serveErr(http.StatusForbidden)
	)
	var serveHTTP = func(w http.ResponseWriter, r *http.Request) {
		c, _ := r.Cookie("_xsrf")
		if c != nil {
			c.Expires = time.Now().Add(-1)
			c.MaxAge = -1
			http.SetCookie(w, c)
		}
		h.ServeHTTP(w, r)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			h.ServeHTTP(w, r)
			return
		}
		// https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#identifying-source-o-via-originreferer-header
		if u, err := url.ParseRequestURI(r.Header.Get("Origin")); err == nil && u.Host != r.Host {
			serve400(w, r, "Bad Header")
			return
		} else if u, err := url.ParseRequestURI(r.Referer()); err == nil && u.Host != r.Host {
			serve400(w, r, "Bad Header")
			return
		}
		// https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html#user-interaction-based-csrf-defense
		c, err := r.Cookie("_xsrf")
		if err != nil {
			serve401(w, r, "Token Not Found") // todo bad request
			return
		}
		if !xsrftoken.Valid(c.Value, "hmac_value", "", "") {
			serve403(w, r, "Token Not Valid")
			return
		}

		serveHTTP(w, r)
	}
}

var serveErr = func(code int) func(w http.ResponseWriter, r *http.Request, s string) {
	return func(w http.ResponseWriter, r *http.Request, s string) {
		c, _ := r.Cookie("_xsrf")
		if c != nil {
			c.Expires = time.Now().Add(-1)
			c.MaxAge = -1
			http.SetCookie(w, c)
		}
		if s == "" {
			s = http.StatusText(code)
		}
		http.Error(w, s, code)
	}
}
