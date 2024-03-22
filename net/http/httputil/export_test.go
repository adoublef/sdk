package httputil

import (
	"context"
	"net/http"
)

func NewTestTimeoutHandler(handler http.Handler, ctx context.Context) http.Handler {
	return &timeoutHandler{
		handler:     handler,
		testContext: ctx,
		// (no body)
	}
}
