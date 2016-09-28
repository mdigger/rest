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

func TestHandler(t *testing.T) {
	r := httptest.NewRequest("", "/test", nil)
	w := httptest.NewRecorder()
	Handler(errorTest).ServeHTTP(w, r)
}
