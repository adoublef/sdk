package httputil

import "net/http"

// Chain applies middlewares to a http.Handler
func Chain(h http.Handler, ff ...func(http.Handler) http.Handler) http.Handler {
	for _, f := range ff {
		h = f(h)
	}
	return h
}

// ChainFunc applies middlewares to a http.HandlerFunc
func ChainFunc(hf http.HandlerFunc, ff ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, f := range ff {
		hf = f(hf)
	}
	return hf
}
