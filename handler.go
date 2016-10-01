package rest

import (
	"log"
	"net/http"
	"os"
)

// Handler describes an HTTP request handler.
//
// A simple contract guarantees robust request handling: if a handler returns an
// error status (>= 400), then the response has not yet been written and the
// client must still be shown an error message. If the error value is not nil,
// then something went wrong on the server and it should be logged/reported.
//
// In other words, the first return value is for the client's benefit, and the
// second return value is for the server. They are completely independent; an
// error status doesn't always mean the error will be non-nil. (For example,
// 404 Not Found is not usually a server error.)
//
// Use code 0 if the response is not sent. Or -1, if the response is sent, but
// it is made by a third-party processor and the status code of the sent
// response is not known.
type Handler func(w http.ResponseWriter, r *http.Request) (code int, err error)

// ServeHTTP implements http.Handler interface.
func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := fn(w, r)
	if code == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else if code >= 400 {
		http.Error(w, http.StatusText(code), code)
	}
	if err != nil {
		log.Println("handler error:", err)
	}
}

// Handlers Handler.
func Handlers(handlers ...Handler) Handler {
	if len(handlers) == 1 {
		return handlers[0]
	}
	return func(w http.ResponseWriter, r *http.Request) (code int, err error) {
		for _, handler := range handlers {
			code, err = handler(w, r)
			if code != 0 || err != nil {
				break
			}
		}
		return
	}
}

// ServeFile replies to the request with the contents of the named file.
func ServeFile(filename string) Handler {
	return func(w http.ResponseWriter, r *http.Request) (code int, err error) {
		switch file, err := os.Open(filename); {
		case err == nil:
			fi, _ := file.Stat()
			http.ServeContent(w, r, filename, fi.ModTime(), file)
			file.Close()
			return http.StatusOK, nil
		case os.IsNotExist(err):
			return http.StatusNotFound, err
		case os.IsPermission(err):
			return http.StatusForbidden, err
		default:
			return http.StatusInternalServerError, err
		}
	}
}
