package hlog_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"go.adoublef.dev/is"
	. "go.adoublef.dev/sdk/net/http/httputil/hlog"
)

func Test_LogHandler(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var (
			c, buf, ctx = newClient(t)
			is          = is.NewRelaxed(t)
		)

		// three request
		// <key=value> textProto?
		r, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
		is.NoErr(err)

		_, err = c.Do(r)
		is.NoErr(err)

		// check all the fields (somehow)
		t.Logf("%s\n", buf)
	})
}

func newClient(tb testing.TB) (*http.Client, *bytes.Buffer, context.Context) {
	tb.Helper()
	var (
		ctx = context.Background()
	)
	var buf bytes.Buffer
	// info level is default
	lh := slog.NewTextHandler(&buf, &slog.HandlerOptions{})

	mux := http.NewServeMux()
	{
		// populate
	}
	ts := httptest.NewServer(LogHandler(mux, slog.New(lh)))
	tb.Cleanup(func() { ts.Close() })

	tc := ts.Client()
	tc.Transport = &RoundTripper{must(url.Parse(ts.URL)), tc.Transport}

	return tc, &buf, ctx
}

type RoundTripper struct {
	u *url.URL
	http.RoundTripper
}

func (rt *RoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	// if relative then use a resolve
	if r.URL.RawPath != "http" {
		r.URL = rt.u.ResolveReference(r.URL)
	}
	// can probably include some other shit here too
	return rt.RoundTripper.RoundTrip(r)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
