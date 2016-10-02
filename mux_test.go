package rest

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/mdigger/log"
)

var _ http.Handler = new(ServeMux)

func TestServeMux(t *testing.T) {
	testHandleFunc := func(w http.ResponseWriter, r *http.Request) (int, error) {
		return Write(w, r, 200, "OK")
	}

	mux := &ServeMux{
		Headers:     map[string]string{"test": "test"},
		NotCompress: false,
		Options: &Options{
			Encoder:       JSONEncoder(false),
			DataAdapter:   Adapter,
			AllowMultiple: false,
		},
		Logger: log.Default,
	}
	// Handle("", "", nil)
	mux.Handle("", "/:name", testHandleFunc)
	mux.Handle("GET", "/:name/test/", testHandleFunc)
	mux.Handle("POST", "/:name", testHandleFunc)

	for _, req := range []struct {
		method, url string
		status      int
	}{
		{"GET", "/test/", 301},
		{"GET", "/test/test", 301},
		{"POST", "/test/", 301},
		{"TEST", "/test/", 404},
		{"GET", "/test", 200},
		{"POST", "/test", 200},
		{"TEST", "/test", 405},
	} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(req.method, req.url, nil)
		mux.ServeHTTP(w, r)
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("%s", dump)
		resp := w.Result()
		dump, err = httputil.DumpResponse(resp, true)
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("%s", dump)
		if resp.StatusCode != req.status {
			t.Errorf("bad status: %v - %03d", req, resp.StatusCode)
		}
		fmt.Println(strings.Repeat("-", 60))
	}
}

func TestServeMuxError(t *testing.T) {
	defer func() {
		p := recover()
		if err, ok := p.(error); !ok ||
			err.Error() != "path parts overflow: 50000" {
			t.Error(p)
		}
	}()
	path := strings.Repeat("/test", 50000)
	mux := new(ServeMux)
	mux.Logger = log.Default
	mux.Handle("GET", path, func(w http.ResponseWriter, r *http.Request) (int, error) {
		return Write(w, r, 200, "OK")
	})
}

func TestError(t *testing.T) {
	mux := new(ServeMux)
	mux.Logger = log.Default
	mux.Handle("GET", "/test", func(w http.ResponseWriter, r *http.Request) (int, error) {
		return 500, errors.New("test error")
	})
	r := httptest.NewRequest("", "/test", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
}
