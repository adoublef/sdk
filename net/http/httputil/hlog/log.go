package hlog

import (
	"log/slog"
	"net/http"
	"time"

	"go.adoublef.dev/sdk/net/http/httputil"
)

// Log wraps a http.Handler with a request logger.
func Log(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := httputil.Wrap(w, r)

		start := time.Now()
		defer func() {
			duration := time.Since(start)

			slog.LogAttrs(
				r.Context(),
				statusLevel(rw.Status()),
				"http request",
				slog.String("method", r.Method),
				slog.Int64("time_ms", int64(duration/time.Millisecond)),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.Status()),
				slog.String("duration", duration.String()),
			)
		}()

		h.ServeHTTP(rw, r)
	})
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
