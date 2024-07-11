package hlog

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"go.adoublef.dev/sdk/net/http/httputil"
)

type contextKey struct{ string }

func (k *contextKey) String() string { return "hlog: context value " + k.string }

var (
	LoggerContextKey = &contextKey{"http-log"}
)

func LogHandler(h http.Handler, sl *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := newLogger(sl, r)
		ww := httputil.Wrap(w, r)

		buf := newLimitBuffer(512)
		ww.Tee(buf)

		t1 := time.Now()
		defer func() {
			var error []byte
			if ww.Status() >= 400 {
				error, _ = io.ReadAll(buf) // lorem is 3091
			}
			l.Write(ww.Status(), ww.Size(), ww.Header(), time.Since(t1), error)
		}()
		ctx := context.WithValue(r.Context(), LoggerContextKey, l)
		h.ServeHTTP(ww, r.WithContext(ctx))
	})
}

// Logger returns the in-context Logger for a request.
func Logger(r *http.Request) *slog.Logger {
	entry, ok := r.Context().Value(LoggerContextKey).(*logger)
	if !ok || entry == nil {
		opts := &slog.HandlerOptions{
			AddSource: true,
			// LevelError+1 will be higher than all levels
			// hence logs would be skipped
			Level: slog.LevelError + 1,
		}
		return slog.New(slog.NewTextHandler(os.Stderr, opts))
	} else {
		return entry.l
	}
}

type logger struct {
	l *slog.Logger
}

func newLogger(l *slog.Logger, r *http.Request) *logger {
	// request can be passed via context
	l = l.With(
		slog.String("path", r.URL.Path),
		slog.String("method", r.Method),
		slog.String("remoteIP", r.RemoteAddr),
	)
	return &logger{l: l}
}

func (l *logger) Write(status, bytes int, header http.Header, elapsed time.Duration, body []byte) {
	// response
	l.l = l.l.With(
		slog.Int("status", status),
		slog.Int("bytes", bytes),
		slog.Duration("elapsed", elapsed),
	)
	if status >= 400 {
		l.l = l.l.With(slog.String("body", strings.TrimSpace(string(body))))
	}
	l.l.Log(context.Background(), statusLevel(status), statusLabel(status))
}

func statusLevel(status int) slog.Level {
	switch {
	case status <= 0:
		return slog.LevelWarn
	case status < 400: // for codes in 100s, 200s, 300s
		return slog.LevelInfo
	case status >= 400 && status < 500:
		// switching to info level to be less noisy
		return slog.LevelInfo
	case status >= 500:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func statusLabel(status int) string {
	switch {
	case status >= 100 && status < 300:
		return "OK"
	case status >= 300 && status < 400:
		return "Redirect"
	case status >= 400 && status < 500:
		return "Client Error"
	case status >= 500:
		return "Server Error"
	default:
		return "Unknown"
	}
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("err", err)
}
