package rest

import (
	"context"
	"net/http"
)

// Write gives the data with the specified status. Use either the default
// settings or out of context if they were used.
func Write(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	var settings *Settings
	// get settings from request context
	if contextSettings, ok := r.Context().Value(keySettings).(*Settings); ok {
		settings = contextSettings
	} else {
		settings = Default // default settings
	}
	// check double response
	if !settings.AllowMultiple && r.Context().Value(keyResponded) != nil {
		panic("rest: multiple responses")
	}
	// add the response headers if they are defined
	var responseHeader = w.Header() // response header
	if len(settings.Headers) > 0 {
		for key, value := range settings.Headers {
			responseHeader.Set(key, value)
		}
	}
	// preprocess response data
	if settings.Preprocessor != nil {
		status, data = settings.Preprocessor(w, r, status, data)
	}
	// set encoder
	var encoder Encoder
	if settings.Encoder != nil {
		encoder = settings.Encoder
	} else {
		encoder = JSONEncoder(true) // JSON by default
	}
	// set response headers
	responseHeader.Set("Content-Type", encoder.ContentType(w, r))
	w.WriteHeader(status)
	// response with data
	if err := encoder.Encode(w, r, data); err != nil {
		if settings.OnError != nil {
			settings.OnError(err)
		} else {
			panic(err)
		}
	}
	// response complete
	if settings.OnComplete != nil {
		settings.OnComplete(w, r, status, data)
	}
	// disallow double response
	if !settings.AllowMultiple {
		ctx := context.WithValue(r.Context(), keyResponded, true)
		nr := r.WithContext(ctx)
		*r = *nr
	}
}
