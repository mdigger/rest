package rest

import "net/http"

// statusResponseWriter is used to catch the response status.
//
// Sometimes you need to call some third-party handler http request, which knows
// what there actually doing. In order to capture the status that he actually
// gives up, and made this wrapper.
type statusResponseWriter struct {
	http.ResponseWriter
	code int
}

// WriteHeader writes the header status code of the response.
func (w *statusResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *statusResponseWriter) Write(b []byte) (int, error) {
	if w.code == 0 {
		w.code = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// Status returns the sent response status code.
func (w statusResponseWriter) Status() int {
	return w.code
}
