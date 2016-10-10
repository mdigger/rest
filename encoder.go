package rest

import "encoding/json"

// Encoder describes an function capable of encoding a response.
type Encoder func(*Context, interface{}) error

// defaultEncoder is used as the default Encoder.
func defaultEncoder(c *Context, v interface{}) error {
	var code = c.Status()
	var response = &struct {
		Code     int         `json:"code"`
		Success  bool        `json:"success"`
		Error    string      `json:"error,omitempty"`
		Location string      `json:"location,omitempty"`
		Data     interface{} `json:"data,omitempty"`
	}{
		Code:    code,
		Success: code < 400,
	}
	switch data := v.(type) {
	case error:
		response.Error = data.Error()
	case *RedirectURL:
		response.Location = data.URL
	default:
		response.Data = data
	}
	c.SetContentType("application/json; charset=utf-8")
	enc := json.NewEncoder(c.Response)
	enc.SetIndent("", "    ")
	return enc.Encode(response)
}
