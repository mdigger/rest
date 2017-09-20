package rest

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mdigger/log"
)

func TestServeMux_HandleErrror(t *testing.T) {
	defer func() {
		p := recover()
		if err, ok := p.(error); !ok ||
			err.Error() != "path parts overflow: 50000" {
			t.Error(p)
		}
	}()
	path := strings.Repeat("/test", 50000)
	mux := new(ServeMux)
	mux.Handle("", path, func(c *Context) error {
		return c.Write("OK")
	})
}

func TestServeMux(t *testing.T) {
	mux := new(ServeMux)
	mux.Headers = map[string]string{
		"X-Server-API": "1.0",
	}
	mux.Logger = log.New("http")
	mux.Handles(Paths{
		"/": {
			"GET": func(c *Context) error {
				return c.Write("OK")
			},
		},
		"/files/*filename": {
			"GET": Files("."),
		},
	}, func(*Context) error { return nil })
	var r = httptest.NewRequest("", "/", nil)
	var w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	r = httptest.NewRequest("", "/files/mux_test.go", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	r = httptest.NewRequest("POST", "/files/mux_test.go", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	r = httptest.NewRequest("GET", "/files/mux_test.go/", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	r = httptest.NewRequest("", "/bad/path/", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
}
