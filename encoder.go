package rest

import "encoding/json"

// Encoder describes an function capable of encoding a response.
type Encoder func(*Context, interface{}) error

// defaultEncoder is used as the default Encoder.
func defaultEncoder(c *Context, v interface{}) error {
	c.SetContentType("application/json; charset=utf-8")
	enc := json.NewEncoder(c.Response)
	enc.SetIndent("", "    ")
	if err, ok := v.(error); ok {
		return enc.Encode(&struct {
			Error string `json:"error,omitempty"`
		}{
			Error: err.Error(),
		})
	}
	return enc.Encode(v)
}
