package rest

import (
	"context"
	"net/http"
)

// Default contains the default Settings.
var Default = &Settings{
	Headers:      map[string]string{"X-API-Version": "1.0"},
	Preprocessor: Preprocessor,
	Encoder:      JSONEncoder(true),
}

// Settings describes the possible settings of the response processing.
type Settings struct {
	// Headers can contain a list of additional headers that will be
	// automatically added to the response.
	Headers map[string]string

	// Preprocessor allows you to change the data and status before they will be
	// given in response.
	Preprocessor func(w http.ResponseWriter, r *http.Request,
		status int, data interface{}) (newStatus int, newData interface{})

	// Encoder includes an Encoder, which is used to return data.
	Encoder Encoder

	// OnError is called in case of error, returns data.
	OnError func(err error)

	// OnComplete is called after the return data. The status and data as well
	// adds the processing time of the query.
	OnComplete func(w http.ResponseWriter, r *http.Request,
		status int, data interface{})

	// AllowMultiple indicates that multiple responses are allowed. Otherwise,
	// multiple calls to With will panic.
	AllowMultiple bool
}

// Handler attach Settings to the context of the request and handle it.
func (s *Settings) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add settings to request context
		ctx := context.WithValue(r.Context(), keySettings, s)
		h.ServeHTTP(w, r.WithContext(ctx)) // serve handler
	})
}

type contextKey byte // context key type
const (
	keySettings contextKey = iota
	keyResponded
)
