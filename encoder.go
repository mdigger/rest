package rest

import (
	"encoding/json"
	"net/http"
)

// JSON is just a quick way to describe data structures.
type JSON map[string]interface{}

// Encoder describes an object capable of encoding a response.
type Encoder interface {
	// ContentType returns a string with the Content-Type to return in response
	// header.
	ContentType(w http.ResponseWriter, r *http.Request) string

	// Encode writes a serialization of v to http.ResponseWriter, optionally
	// using additional information from the http.Request to do so.
	Encode(w http.ResponseWriter, r *http.Request, v interface{}) error
}

// JSONEncoder is responsible for return information in JSON format. If true,
// the data will be formatted with indentation.
type JSONEncoder bool

// ContentType returns a string with the Content-Type to return in response
// header.
func (JSONEncoder) ContentType(w http.ResponseWriter, r *http.Request) string {
	return "application/json; charset=utf-8"
}

// Encode writes a serialization of v to http.ResponseWriter in JSON format.
func (indent JSONEncoder) Encode(w http.ResponseWriter, r *http.Request,
	v interface{}) error {
	enc := json.NewEncoder(w)
	if indent {
		enc.SetIndent("", "    ")
	}
	return enc.Encode(v)
}
