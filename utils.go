package rest

import (
	"net"
	"net/http"

	"github.com/mdigger/router"
)

// JSON is just a quick way to describe data structures.
type JSON map[string]interface{}

// RealIP returns a real IP address from headers.
func RealIP(r *http.Request) string {
	addr := r.RemoteAddr
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		addr = ip
	} else if ip := r.Header.Get("X-Real-IP"); ip != "" {
		addr = ip
	} else {
		addr, _, _ = net.SplitHostPort(addr)
	}
	return addr
}

// Params returns a list of named path parameters stored in the request context.
func Params(r *http.Request) router.Params {
	if params, ok := r.Context().Value(keyParams).(router.Params); ok {
		return params
	}
	return nil
}

type contextKey byte // context key type
const (
	keyParams    contextKey = iota // params key
	keyOptions                     // write options key
	keyResponded                   // double response key
)

// statusResponseWriter is used to catch the response status.
//
// Sometimes you need to call some third-party handler http request, which knows
// what there actually doing. In order to capture the status that he actually
// gives up, and made this wrapper.
type statusResponseWriter struct {
	http.ResponseWriter
	code int
}

// WriteHeader writes the header status code of the response.
func (w *statusResponseWriter) WriteHeader(status int) {
	w.code = status
	w.ResponseWriter.WriteHeader(status)
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *statusResponseWriter) Write(b []byte) (int, error) {
	if w.code == 0 {
		w.code = 200
	}
	return w.ResponseWriter.Write(b)
}

// Status returns the sent response status code.
func (w statusResponseWriter) Status() int {
	return w.code
}
