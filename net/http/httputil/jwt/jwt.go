package jwt

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	Authorization = "Authorization"
)

var (
	jwtKey = &contextKey{"jwt"}
)

type contextKey struct {
	s string
}

func (k contextKey) String() string {
	return "jwt: " + k.s
}

type Parser interface {
	Parse(h http.HandlerFunc) http.HandlerFunc
}

// parser
type parser struct {
	c   *jwk.Cache
	url string
}

// New [Parser] is returned.
func New(ctx context.Context, endpoint string) (Parser, error) {
	c := jwk.NewCache(ctx)
	err := c.Register(endpoint)
	if err != nil {
		return nil, err
	}
	// ping to be sure the url is valid
	_, err = c.Refresh(ctx, endpoint)
	if err != nil {
		return nil, errors.Join(err, ErrInvalidURL)
	}
	return &parser{c: c, url: endpoint}, nil
}

// Parse
func (p *parser) Parse(h http.HandlerFunc) http.HandlerFunc {
	var (
		serve401 = serveErr(http.StatusUnauthorized)
		serve404 = serveErr(http.StatusNotFound)
		serve503 = serveErr(http.StatusServiceUnavailable)
	)
	return func(w http.ResponseWriter, r *http.Request) {
		_, raw, found := strings.Cut(r.Header.Get(Authorization), " ")
		if !found {
			serve401(w, r, "")
			return
		}
		jwks, err := p.c.Refresh(r.Context(), p.url)
		if err != nil {
			serve503(w, r, "")
			return
		}
		tk, err := jwt.ParseString(raw, jwt.WithKeySet(jwks))
		if err != nil {
			serve404(w, r, "")
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), jwtKey, tk))
		h.ServeHTTP(w, r)
	}
}

// Token
func Token(ctx context.Context) jwt.Token {
	tk, ok := ctx.Value(jwtKey).(jwt.Token)
	if !ok {
		return nil
	}
	return tk
}

// Subtle: this type shadows the type jwt.Vistor
type Visitor = jwt.Visitor

// TODO function for self-generated tokens

var serveErr = func(code int) func(w http.ResponseWriter, _ *http.Request, s string) {
	return func(w http.ResponseWriter, _ *http.Request, s string) {
		if s == "" {
			s = http.StatusText(code)
		}
		http.Error(w, s, code)
	}
}

var (
	ErrInvalidURL = errors.New("jwt: invalid url")
)
