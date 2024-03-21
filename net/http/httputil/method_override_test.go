package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.adoublef.dev/is"
	. "go.adoublef.dev/sdk/net/http/httputil"
)

func Test_MethodOverride(t *testing.T) {
	t.Run("Query", func(t *testing.T) {
		is := is.NewRelaxed(t)

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /{$}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		s := httptest.NewServer(MethodOverride(mux))
		t.Cleanup(func() { s.Close() })

		rs, err := s.Client().Post(s.URL+"?_method=DELETE", "", nil)
		is.NoErr(err) // (http.Client).Post
		is.Equal(rs.StatusCode, http.StatusOK)
	})

	t.Run("Header", func(t *testing.T) {
		is := is.NewRelaxed(t)

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /{$}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		s := httptest.NewServer(MethodOverride(mux))
		t.Cleanup(func() { s.Close() })

		r, err := http.NewRequest(http.MethodPost, s.URL, nil)
		is.NoErr(err) // http.NewRequest

		r.Header.Set("X-HTTP-Method-Override", http.MethodDelete)

		rs, err := s.Client().Do(r)
		is.NoErr(err) // (http.Client).Post
		is.Equal(rs.StatusCode, http.StatusOK)
	})
}
