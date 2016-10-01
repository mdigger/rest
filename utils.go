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
