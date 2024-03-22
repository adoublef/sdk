package httputil_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"go.adoublef.dev/is"
	"go.adoublef.dev/sdk/io/fs"
	"go.adoublef.dev/sdk/io/iotest"
	. "go.adoublef.dev/sdk/net/http/httputil"
)

// cancellableTimeoutContext overwrites the error message to DeadlineExceeded
type cancellableTimeoutContext struct {
	context.Context
}

func (c cancellableTimeoutContext) Err() error {
	if c.Context.Err() != nil {
		return context.DeadlineExceeded
	}
	return nil
}

// Export https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/export_test.go;l=89
func Test_TimeoutHandler(t *testing.T) {
	// Source https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/serve_test.go;l=2294
	t.Run("OK", func(t *testing.T) {
		is := is.NewRelaxed(t)

		sendHi := make(chan bool, 1)
		writeErrors := make(chan error, 1)
		sayHi := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-sendHi
			_, werr := w.Write([]byte("hi"))
			writeErrors <- werr
		})
		ctx, cancel := context.WithCancel(context.Background())
		ts := httptest.NewServer(NewTestTimeoutHandler(sayHi, cancellableTimeoutContext{ctx}))

		// succeed without a timing out:
		sendHi <- true
		rs, err := ts.Client().Get(ts.URL)
		is.NoErr(err) // (http.Client).Get
		is.Equal(rs.StatusCode, http.StatusOK)

		body, _ := io.ReadAll(rs.Body)
		is.Equal(string(body), "hi")
		is.True(errors.Is(<-writeErrors, nil))

		// Times out:
		cancel()

		rs, err = ts.Client().Get(ts.URL)
		is.NoErr(err) // (http.Client).Get
		is.Equal(rs.StatusCode, http.StatusServiceUnavailable)

		body, _ = io.ReadAll(rs.Body)
		is.True(strings.Contains(string(body), "<title>Timeout</title>"))
		is.Equal(rs.Header.Get("Content-Type"), "text/html; charset=utf-8")

		// Now make the previously-timed out handler speak again,
		// which verifies the panic is handled:
		sendHi <- true
		is.True(errors.Is(<-writeErrors, http.ErrHandlerTimeout))
	})

	// Issues 8209 and 8414.
	// Source https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/serve_test.go;l=2351
	t.Run("Race", func(t *testing.T) {
		delayHi := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ms, _ := strconv.Atoi(r.URL.Path[1:])
			if ms == 0 {
				ms = 1
			}
			for i := 0; i < ms; i++ {
				w.Write([]byte("hi"))
				time.Sleep(time.Millisecond)
			}
		})

		ts := httptest.NewServer(TimeoutHandler(delayHi, 20*time.Millisecond, ""))
		c := ts.Client()

		var wg sync.WaitGroup
		gate := make(chan bool, 10)
		n := 50
		if testing.Short() {
			n = 10
			gate = make(chan bool, 3)
		}
		for i := 0; i < n; i++ {
			gate <- true
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-gate }()
				res, err := c.Get(fmt.Sprintf("%s/%d", ts.URL, rand.Intn(50)))
				if err == nil {
					io.Copy(io.Discard, res.Body)
					res.Body.Close()
				}
			}()
		}
		wg.Wait()
	})

	// Source https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/serve_test.go;l=2393
	t.Run("RaceHeader", func(t *testing.T) {
		delay204 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(204)
		})

		ts := httptest.NewServer(TimeoutHandler(delay204, time.Nanosecond, ""))

		var wg sync.WaitGroup
		gate := make(chan bool, 50)
		n := 500
		if testing.Short() {
			n = 10
		}

		c := ts.Client()
		for i := 0; i < n; i++ {
			gate <- true
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-gate }()
				res, err := c.Get(ts.URL)
				if err != nil {
					// We see ECONNRESET from the connection occasionally,
					// and that's OK: this test is checking that the server does not panic.
					t.Log(err)
					return
				}
				defer res.Body.Close()
				io.Copy(io.Discard, res.Body)
			}()
		}
		wg.Wait()
	})

	// Issue 9162.
	// Source https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/serve_test.go;l=2430
	t.Run("RaceHeaderTimeout", func(t *testing.T) {
		is := is.NewRelaxed(t)

		sendHi := make(chan bool, 1)
		writeErrors := make(chan error, 1)
		sayHi := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			<-sendHi
			_, werr := w.Write([]byte("hi"))
			writeErrors <- werr
		})
		ctx, cancel := context.WithCancel(context.Background())
		// NewTestTimeoutHandler(sayHi, cancellableTimeoutContext{ctx})
		ts := httptest.NewServer(NewTestTimeoutHandler(sayHi, cancellableTimeoutContext{ctx}))

		// Succeed without timing out:
		sendHi <- true
		rs, err := ts.Client().Get(ts.URL)
		is.NoErr(err)                          // (http.Client).Get
		is.Equal(rs.StatusCode, http.StatusOK) // 200

		body, _ := io.ReadAll(rs.Body)
		is.Equal(string(body), "hi")
		is.True(errors.Is(<-writeErrors, nil))

		// Times out:
		cancel()

		rs, err = ts.Client().Get(ts.URL)
		is.NoErr(err)                                          // (http.Client).Get
		is.Equal(rs.StatusCode, http.StatusServiceUnavailable) // 503

		body, _ = io.ReadAll(rs.Body)
		is.True(strings.Contains(string(body), "<title>Timeout</title>"))

		// Now make the previously-timed out handler speak again,
		// which verifies the panic is handled:
		sendHi <- true
		is.True(errors.Is(<-writeErrors, http.ErrHandlerTimeout))
	})

	// Issue 14568
	// Source https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/serve_test.go;l=2485
	t.Run("StartTimerWhenServing", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping sleeping test in -short mode")
		}
		is := is.NewRelaxed(t)

		var handler http.HandlerFunc = func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}
		timeout := 300 * time.Millisecond
		ts := httptest.NewServer(TimeoutHandler(handler, timeout, ""))
		t.Cleanup(func() { ts.Close() })

		c := ts.Client()

		// Issue was caused by the timeout handler starting the timer when
		// was created, not when the request. So wait for more than the timeout
		// to ensure that's not the case.
		time.Sleep(2 * timeout)
		rs, err := c.Get(ts.URL)
		is.NoErr(err) // (http.Client).Get

		t.Cleanup(func() { rs.Body.Close() })
		is.Equal(rs.StatusCode, http.StatusNoContent) // 204
	})

	// Canceled https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/serve_test.go;l=2515
	t.Run("ContextCanceled", func(t *testing.T) {
		is := is.NewRelaxed(t)

		writeErrors := make(chan error, 1)
		sayHi := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			var err error
			// The request context has already been canceled, but
			// retry the write for a while to give the timeout handler
			// a chance to notice.
			for i := 0; i < 100; i++ {
				_, err = w.Write([]byte("a"))
				if err != nil {
					break
				}
				time.Sleep(1 * time.Millisecond)
			}
			writeErrors <- err
		})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		// NewTestTimeoutHandler(sayHi, ctx)
		ts := httptest.NewServer(NewTestTimeoutHandler(sayHi, ctx))
		t.Cleanup(func() { ts.Close() })

		rs, err := ts.Client().Get(ts.URL)
		is.NoErr(err) // (http.Client).Get
		is.Equal(rs.StatusCode, http.StatusServiceUnavailable)

		body, _ := io.ReadAll(rs.Body)
		is.True(strings.Contains(string(body), ""))

		is.True(errors.Is(<-writeErrors, context.Canceled))
	})

	// Issue 15948
	// Source https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/serve_test.go;l=2556
	t.Run("EmptyResponse", func(t *testing.T) {
		is := is.NewRelaxed(t)

		var handler http.HandlerFunc = func(w http.ResponseWriter, _ *http.Request) {
			// No response.
		}
		timeout := 300 * time.Millisecond
		ts := httptest.NewServer(TimeoutHandler(handler, timeout, ""))
		c := ts.Client()

		rs, err := c.Get(ts.URL)
		is.NoErr(err) // (http.Client).Get
		t.Cleanup(func() { rs.Body.Close() })
		is.Equal(rs.StatusCode, http.StatusOK)
	})

	// Issue 22084
	// Source https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/serve_test.go;l=2577
	t.Run("PanicRecovery", func(t *testing.T) {
		is := is.NewRelaxed(t)

		pr, pw := io.Pipe()
		t.Cleanup(func() {
			pw.Close()
		})

		var h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("intentional death for testing")
		})

		ts := httptest.NewUnstartedServer(TimeoutHandler(h, time.Second, ""))
		ts.Config.ErrorLog = log.New(pw, "", 0)
		ts.Start()

		// Do a blocking read on the log output pipe.
		done := make(chan bool, 1)
		go func() {
			buf := make([]byte, 4<<10)
			_, err := pr.Read(buf)
			pr.Close()
			is.NoErr(err) // (io.PipeReader).Read
			is.True(err != io.EOF)
			done <- true
		}()

		_, err := ts.Client().Get(ts.URL)
		is.True(err != nil) // (http.Client).Get expected error

		<-done
	})

	// Issue 30803
	// Source https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/net/http/serve_test.go;l=6336
	t.Run("SuperfluousLog", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping in short mode")
		}

		pc, curFile, _, _ := runtime.Caller(0)
		curFileBaseName := filepath.Base(curFile)
		testFuncName := runtime.FuncForPC(pc).Name()

		timeoutMsg := "timed out here!"

		tt := map[string]struct {
			mustTimeout bool
			wantResp    string
		}{
			// "return before timeout": {
			// 	wantResp: "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n",
			// },
			"return after timeout": {
				mustTimeout: true,
				wantResp: fmt.Sprintf("HTTP/1.1 503 Service Unavailable\r\nContent-Length: %d\r\n\r\n%s",
					len(timeoutMsg), timeoutMsg),
			},
		}

		for name, tc := range tt {
			t.Run(name, func(t *testing.T) {
				is.NewRelaxed(t)

				exitHandler := make(chan bool, 1)
				t.Cleanup(func() { close(exitHandler) })
				lastLine := make(chan int, 1)

				sh := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(404)
					w.WriteHeader(404)
					w.WriteHeader(404)
					w.WriteHeader(404)
					_, _, line, _ := runtime.Caller(0)
					lastLine <- line
					<-exitHandler
				})

				if !tc.mustTimeout {
					exitHandler <- true
				}

				logBuf := new(strings.Builder)
				srvLog := log.New(logBuf, "", 0)
				// When expecting to timeout, we'll keep the duration short.
				dur := 20 * time.Millisecond
				if !tc.mustTimeout {
					// Otherwise, make it arbitrarily long to reduce the risk of flakes.
					dur = 10 * time.Second
				}
				th := TimeoutHandler(sh, dur, timeoutMsg)
				ts := httptest.NewUnstartedServer(th)
				ts.Config.ErrorLog = srvLog
				ts.Start()
				t.Cleanup(func() { ts.Close() })

				rs, err := ts.Client().Get(ts.URL)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				// Deliberately removing the "Date" header since it is highly ephemeral
				// and will cause failure if we try to match it exactly.
				rs.Header.Del("Date")
				rs.Header.Del("Content-Type")

				// Match the response.
				blob, _ := httputil.DumpResponse(rs, true)
				if g, w := string(blob), tc.wantResp; g != w {
					t.Errorf("Response mismatch\nGot\n%q\n\nWant\n%q", g, w)
				}

				// Given 4 w.WriteHeader calls, only the first one is valid
				// and the rest should be reported as the 3 spurious logs.
				logEntries := strings.Split(strings.TrimSpace(logBuf.String()), "\n")
				if g, w := len(logEntries), 3; g != w {
					blob, _ := json.MarshalIndent(logEntries, "", "  ")
					t.Fatalf("Server logs count mismatch\ngot %d, want %d\n\nGot\n%s\n", g, w, blob)
				}

				lastSpuriousLine := <-lastLine
				firstSpuriousLine := lastSpuriousLine - 3
				// Now ensure that the regexes match exactly.
				//      "http: superfluous response.WriteHeader call from <fn>.func\d.\d (<curFile>:lastSpuriousLine-[1, 3]"
				for i, logEntry := range logEntries {
					wantLine := firstSpuriousLine + i
					pat := fmt.Sprintf("^http: superfluous response.WriteHeader call from %s.\\d+.\\d+ \\(%s:%d\\)$",
						testFuncName, curFileBaseName, wantLine)
					re := regexp.MustCompile(pat)
					if !re.MatchString(logEntry) {
						t.Errorf("Log entry mismatch\n\t%s\ndoes not match\n\t%s", logEntry, pat)
					}
				}
			})
		}
	})

	// Set
	t.Run("ResponseController", func(t *testing.T) {
		is := is.NewRelaxed(t)

		var handler http.HandlerFunc = func(w http.ResponseWriter, _ *http.Request) {
			rc := http.NewResponseController(w)

			// ... and set deadline of 2 seconds for writing request body
			if err := rc.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			io.Copy(w, iotest.NewReader(fs.GB.Int()))
		}
		timeout := 10 * time.Second
		ts := httptest.NewServer(TimeoutHandler(handler, timeout, ""))
		c := ts.Client()

		rs, err := c.Get(ts.URL)
		is.NoErr(err) // (http.Client).Get
		t.Cleanup(func() { rs.Body.Close() })
		n, err := io.Copy(io.Discard, rs.Body)
		is.NoErr(err) // io.Copy
		t.Logf("written: %d", n)
	})
}

func Benchmark_TimeoutHandler(b *testing.B) {
	var handler http.HandlerFunc = func(w http.ResponseWriter, _ *http.Request) {
		rc := http.NewResponseController(w)

		// ... and set deadline of 2 seconds for reading request body
		if err := rc.SetReadDeadline(time.Now().Add(time.Minute * 2)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		io.Copy(w, iotest.NewReader(fs.GB.Int()))
	}
	timeout := 300 * time.Millisecond
	ts := httptest.NewServer(TimeoutHandler(handler, timeout, ""))
	c := ts.Client()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rs, err := c.Get(ts.URL)
		if err != nil {
			b.Fatal(err)
		}
		defer rs.Body.Close()
		io.Copy(io.Discard, rs.Body)
	}
}
