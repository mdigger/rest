package rest

import (
	"fmt"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
)

func TestHandler(t *testing.T) {
	var r = httptest.NewRequest("", "/", nil)
	var w = httptest.NewRecorder()
	Handler(func(*Context) error {
		return nil
	}).ServeHTTP(w, r)
	NotImplemented.ServeHTTP(w, r)
	NotFound.ServeHTTP(w, r)
	Redirect("/").ServeHTTP(w, r)
	File("handler_test.go").ServeHTTP(w, r)
	File("bad_file").ServeHTTP(w, r)
	Files(".").ServeHTTP(w, r)
	Data("OK", "text/plain").ServeHTTP(w, r)
}

func TestHandlers(t *testing.T) {
	var r = httptest.NewRequest("", "/", nil)
	var w = httptest.NewRecorder()
	Handlers(
		func(*Context) error { return nil },
		func(*Context) error { return ErrNotImplemented },
	).ServeHTTP(w, r)
	dump, err := httputil.DumpResponse(w.Result(), true)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n-------------\n", dump)

	r = httptest.NewRequest("", "/", nil)
	w = httptest.NewRecorder()
	Handlers(
		func(c *Context) error { return c.Write("OK") },
		func(*Context) error { return ErrNotImplemented },
	).ServeHTTP(w, r)
	dump, err = httputil.DumpResponse(w.Result(), true)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s\n-------------\n", dump)
}
