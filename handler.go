package rest

import (
	"log"
	"net/http"
)

// Handler describes an HTTP request handler.
//
// A simple contract guarantees robust request handling: if a handler returns an
// error status (>= 400), then the response has not yet been written and the
// client must still be shown an error message. If the error value is not nil,
// then something went wrong on the server and it should be logged/reported.
//
// In other words, the first return value is for the client's benefit, and the
// second return value is for the server. They are completely independent; an
// error status doesn't always mean the error will be non-nil. (For example,
// 404 Not Found is not usually a server error.)
type Handler func(w http.ResponseWriter, r *http.Request) (code int, err error)

// ServeHTTP implements http.Handler interface.
func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := fn(w, r)
	if code >= 400 {
		http.Error(w, http.StatusText(code), code)
	}
	if err != nil {
		log.Println("handler error:", err)
	}
}
