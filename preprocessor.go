package rest

import "net/http"

// Response used by standard Preprocessor for generating response.
type Response struct {
	Code     int         `json:"code"`
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
	Status   string      `json:"status,omitempty"`
	Redirect RedirectURL `json:"redirect,omitempty"`
}

// Preprocessor is a default data preprocessor for data response.
func Preprocessor(w http.ResponseWriter, r *http.Request, status int,
	data interface{}) (int, interface{}) {
	var resp = &Response{Code: status}
	switch data := data.(type) {
	case error:
		resp.Error = data.Error()
		resp.Success = false
	case RedirectURL:
		resp.Status = http.StatusText(status)
		resp.Redirect = data
		resp.Success = true
	case nil:
		resp.Status = http.StatusText(status)
		resp.Success = status < 400
	default:
		resp.Data = data
		resp.Success = true
	}
	return status, resp
}
