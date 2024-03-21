package httputil

import (
	"net/http"
)

// MethodOverride is a higher order [http.handlerFunc] which allows the client
// to override a POST request with a PUT, PATCH or DELETE.
//
// A search query ("_method") can be used or header ("X-HTTP-Method-Override").
func MethodOverride(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			method := r.URL.Query().Get("_method")
			if method == "" {
				method = r.Header.Get("X-HTTP-Method-Override")
			}
			if method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
				r2 := r.Clone(r.Context())
				r2.Method = method
				h.ServeHTTP(w, r2)
				return
			}
		}

		h.ServeHTTP(w, r)
	})
}
