package rest

import (
	"context"
	"net/http"
)

// Options describes the settings for returns information.
type Options struct {
	// Encoder includes an Encoder, which is used to write data.
	Encoder Encoder

	// The DataAdapter provides data for the response to the common view.
	DataAdapter DataAdapter

	// AllowMultiple indicates that multiple responses are allowed. Otherwise,
	// multiple calls to Write will panic.
	AllowMultiple bool
}

var defaultOptions = new(Options)

// Write gives the data with the specified status. Use either the default
// settings or out of context if they were used.
//
// Note: errors are a first class concern in Go, with most methods optionally
// returning one as the second argument — but if you try to encode an error
// object into JSON, you will receive an empty response. This is because error
// is an interface type, and there are no exported fields. The easiest way to
// override the error in function of preprocessing.
func Write(w http.ResponseWriter, r *http.Request,
	code int, data interface{}) (int, error) {
	opt, ok := r.Context().Value(keyOptions).(*Options)
	if !ok || opt == nil {
		opt = defaultOptions
	}
	// check double response
	if !opt.AllowMultiple && r.Context().Value(keyResponded) != nil {
		panic("rest: multiple responses")
	}
	// preprocess response data
	if opt.DataAdapter != nil {
		code, data = opt.DataAdapter(w, r, code, data)
	}
	// set encoder
	var encoder Encoder
	if opt.Encoder != nil {
		encoder = opt.Encoder
	} else {
		encoder = defaultEncoder // JSON by default
	}
	// set response headers
	w.Header().Set("Content-Type", encoder.ContentType(w, r))
	w.WriteHeader(code)
	// write data
	err := encoder.Encode(w, r, data)
	if err != nil {
		return code, err
	}
	// disallow double response
	if !opt.AllowMultiple {
		ctx := context.WithValue(r.Context(), keyResponded, true)
		nr := r.WithContext(ctx)
		*r = *nr
	}
	return code, nil
}
