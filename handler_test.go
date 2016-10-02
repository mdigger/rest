package rest

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"testing"

	"github.com/mdigger/log"
)

var _ http.Handler = new(Handler)

func errorTest(w http.ResponseWriter, r *http.Request) (int, error) {
	return 500, errors.New("test error")
}

func info(w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, nil
}

func info2(w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, &os.PathError{"open", "test", os.ErrPermission}
}

func TestHandler(t *testing.T) {
	r := httptest.NewRequest("", "/test", nil)
	w := httptest.NewRecorder()
	Handler(errorTest).ServeHTTP(w, r)
}

func TestMutipleHandlers(t *testing.T) {
	r := httptest.NewRequest("", "/test", nil)
	w := httptest.NewRecorder()
	Handlers(info, info2).ServeHTTP(w, r)
}

func TestServeFileHandler(t *testing.T) {
	mux := new(ServeMux)
	mux.Logger = log.Default
	mux.Handle("GET", "/", ServeFileHandler("index.html"))
	mux.Handle("GET", "/license", ServeFileHandler("LICENSE"))
	mux.Handle("GET", "/error",
		func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 200, &os.PathError{"open", "test", os.ErrPermission}
		})
	mux.Handle("GET", "/error2",
		func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 500, nil
		})
	mux.Handle("GET", "/redirect",
		func(w http.ResponseWriter, r *http.Request) (int, error) {
			return Redirect(w, r, 301, "/")
		})
	mux.Handle("GET", "/null",
		func(w http.ResponseWriter, r *http.Request) (int, error) {
			return 0, nil
		})
	mux.Handle("GET", "/unknown",
		func(w http.ResponseWriter, r *http.Request) (int, error) {
			w.WriteHeader(404)
			return -1, nil
		})
	for i, url := range []string{
		"/", "/license", "/error", "/error2", "/redirect", "/null", "/unknown",
	} {
		fmt.Printf("%d: %s\n", i, url)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("", url, nil)
		mux.ServeHTTP(w, r)
		resp := w.Result()
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.WithError(err).Error("dump error")
		}
		fmt.Printf("%s--------\n", dump)
	}
}
