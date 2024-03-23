package httputil

import (
	"net/http"
	"time"
)

// SetTimeout wraps a [http.HandlerFunc] with a [http.TimeoutHandler] with a given time limit.
//
// TimeoutHandler supports the [http.Pusher] interface but does not support the [http.Hijacker] or [http.Flusher] interfaces.
func SetTimeout(dt time.Duration, msg string) func(http.HandlerFunc) http.HandlerFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			http.TimeoutHandler(hf, dt, msg).ServeHTTP(w, r)
		}
	}
}
