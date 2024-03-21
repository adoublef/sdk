package httputil

import (
	"net/http"
)

func MethodOverride(h http.Handler) http.Handler {
	// https://learn.microsoft.com/en-us/dotnet/api/microsoft.aspnetcore.builder.httpmethodoverrideextensions.usehttpmethodoverride?view=aspnetcore-8.0
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
