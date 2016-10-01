package rest

import "net/http"

// DataAdapter describes the function interface to change the data before it
// writes to response.
//
// The Adapter function is an example implementation of this interface.
type DataAdapter func(w http.ResponseWriter, r *http.Request,
	code int, data interface{}) (newCode int, newData interface{})

// Response used by Adapter for generating data response.
type Response struct {
	Code    int         `json:"code"`
	Status  string      `json:"status,omitempty"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Adapter implements DataAdapter interface.
//
// This function represents the responses transmitted by the Write function in
// the form Response:
// 	{
//	    "code": 200,
// 	    "status": "OK",
// 	    "success": true,
// 	    "data": {
//	        ...
// 	    }
//	}
func Adapter(w http.ResponseWriter, r *http.Request,
	code int, data interface{}) (int, interface{}) {
	var resp = &Response{
		Code:   code,
		Status: http.StatusText(code),
	}
	switch data := data.(type) {
	case error:
		resp.Error = data.Error()
		resp.Success = false
	case nil:
		resp.Success = code < 400
	default:
		resp.Data = data
		resp.Success = true
	}
	return code, resp
}
