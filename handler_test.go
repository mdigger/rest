package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

var _ http.Handler = new(Handler)

func errorTest(w http.ResponseWriter, r *http.Request) (int, error) {
	return 500, errors.New("test error")
}

func info(w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, nil
}

func TestHandler(t *testing.T) {
	r := httptest.NewRequest("", "/test", nil)
	w := httptest.NewRecorder()
	Handler(errorTest).ServeHTTP(w, r)
}

func TestMutipleHandlers(t *testing.T) {
	r := httptest.NewRequest("", "/test", nil)
	w := httptest.NewRecorder()
	Handlers(info, errorTest).ServeHTTP(w, r)
}

func TestServeFileHandler(t *testing.T) {
	mux := new(ServeMux)
	mux.Handle("GET", "/", ServeFileHandler("index.html"))
	mux.Handle("GET", "/license", ServeFileHandler("LICENSE"))
	r := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	r = httptest.NewRequest("", "/license", nil)
	mux.ServeHTTP(w, r)
}
