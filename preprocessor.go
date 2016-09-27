package rest

import "net/http"

// Response used by standard Preprocessor for generating response.
type Response struct {
	Code    int         `json:"code"`
	Status  string      `json:"status,omitempty"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Preprocessor is a default data preprocessor for data response.
func Preprocessor(w http.ResponseWriter, r *http.Request, status int,
	data interface{}) (int, interface{}) {
	var resp = &Response{
		Code:   status,
		Status: http.StatusText(status),
	}
	switch data := data.(type) {
	case error:
		resp.Error = data.Error()
		resp.Success = false
	case nil:
		resp.Success = status < 400
	default:
		resp.Data = data
		resp.Success = true
	}
	return status, resp
}
