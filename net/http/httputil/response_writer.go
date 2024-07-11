package httputil

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	// Status returns the status code of the response or 0 if the response has not been written.
	Status() int
	// Written returns whether or not the ResponseWriter has been written.
	Written() bool
	// Size returns the size of the response body.
	Size() int
	Unwrap() http.ResponseWriter
	// Tee allows the caller to write to multiple writers.
	Tee(w io.Writer)
}

type response struct {
	http.ResponseWriter
	method string
	status int
	size   int
	mu     sync.Mutex
	tee    io.Writer
}

// Tee implements ResponseWriter.
func (r *response) Tee(w io.Writer) {
	r.mu.Lock()
	r.tee = w
	r.mu.Unlock()
}

// Size implements ResponseWriter.
func (rw *response) Size() int {
	return rw.size
}

// Unwrap implements ResponseWriter.
func (rw *response) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// Written implements ResponseWriter.
func (rw *response) Written() bool {
	return rw.status != 0
}

// Status implements ResponseWriter.
func (rw *response) Status() int {
	return rw.status
}

// Write implements ResponseWriter.
// Subtle: this method shadows the method (ResponseWriter).Write of response.ResponseWriter.
func (r *response) Write(b []byte) (size int, err error) {
	if !r.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		r.WriteHeader(http.StatusOK)
	}
	if r.method != http.MethodHead {
		size, err = r.ResponseWriter.Write(b)
		r.mu.Lock()
		if r.tee != nil {
			_, err2 := r.tee.Write(b[:size])
			if err == nil {
				err = err2
			}
		}
		r.mu.Unlock()
		r.size += size
	}
	return size, err
}

// WriteHeader implements ResponseWriter.
// Subtle: this method shadows the method (ResponseWriter).WriteHeader of response.ResponseWriter.
func (rw *response) WriteHeader(s int) {
	// Avoid panic if status code is not a valid HTTP status code
	if s < 100 || s > 999 {
		rw.ResponseWriter.WriteHeader(500)
		rw.status = 500
		return
	}

	rw.ResponseWriter.WriteHeader(s)
	rw.status = s
}

func (rw *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, ErrHijackUnsupported
	}

	conn, brw, err := hijacker.Hijack()
	if err == nil {
		rw.status = -1
	}

	return conn, brw, err
}

func (rw *response) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Wrap http.ResponseWriter into a ResponseWriter.
func Wrap(w http.ResponseWriter, r *http.Request) ResponseWriter {
	if rw, ok := w.(ResponseWriter); ok {
		return rw
	}
	return &response{ResponseWriter: w, method: r.Method}
}

var (
	ErrHijackUnsupported = errors.New("the ResponseWriter doesn't support the Hijacker interface")
)
