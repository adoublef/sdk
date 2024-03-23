package httputil_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.adoublef.dev/is"
	"go.adoublef.dev/sdk/io/fs"
	"go.adoublef.dev/sdk/io/iotest"
	. "go.adoublef.dev/sdk/net/http/httputil"
)

func Test_SetTimeout(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)

		var hf http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			_, err := io.Copy(w, r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		timeout := 300 * time.Millisecond
		cf := ChainFunc(hf, SetTimeout(timeout, ""))

		ts := httptest.NewServer(cf)
		t.Cleanup(func() { ts.Close() })

		c := ts.Client()

		rs, err := c.Post(ts.URL, "", iotest.NewReader(10*fs.GB.Int()))
		is.NoErr(err) // (http.Client).Post
		t.Cleanup(func() { rs.Body.Close() })
		_, err = io.Copy(io.Discard, rs.Body)
		is.NoErr(err)
	})
}
