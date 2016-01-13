package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	for i, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "CONNECT"} {
		r, err := http.NewRequest(method, "/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		c := newContext(w, r)
		if err := c.Status(100 * i).Send(nil); err != nil {
			t.Error(err)
		}
		c.close()
		// if w.Code != http.StatusNoContent {
		// 	t.Error("bad status")
		// }
	}
}

func TestLogger2(t *testing.T) {
	Debug = true
	SetLogger(os.Stderr)
	var mux ServeMux
	mux.Handle("GET", "/test", func(c *Context) error { return c.Send("test1") })
	mux.Handle("POST", "/test", func(c *Context) error { return c.Send(errors.New("test2")) })
	mux.Handle("PUT", "/test", func(c *Context) error { return c.Send([]byte("test3")) })
	mux.Handle("PATCH", "/test", func(c *Context) error { return c.Send(strings.NewReader("test4")) })
	mux.Handle("TEST", "/test", func(c *Context) error {
		return c.Send(struct {
			Test complex64
		}{
			Test: complex(3, 15),
		})
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "TEST"} {
		req, err := http.NewRequest(method, ts.URL+"/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
		}
		// dump, err := httputil.DumpResponse(resp, true)
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// fmt.Printf("%s\n%s\n", dump, strings.Repeat("-", 40))
		resp.Body.Close()
	}
}

func test1(c *Context) error { return errors.New("test error") }
func test2(c *Context) error { panic("panic error"); return nil }

func TestLoggerError(t *testing.T) {
	Debug = true
	SetLogger(os.Stderr)
	var mux ServeMux
	mux.Handle("GET", "/test", test1)
	mux.Handle("POST", "/test", test2)
	ts := httptest.NewServer(mux)
	defer ts.Close()
	for _, method := range []string{"GET", "POST"} {
		req, err := http.NewRequest(method, ts.URL+"/test", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
		}
		resp.Body.Close()
	}
}
