package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParams(t *testing.T) {
	testHandleFunc := func(w http.ResponseWriter, r *http.Request) (int, error) {
		params := Params(r)
		if params == nil || len(params) != 1 || params.Get("name") != "test" {
			t.Error("bad params")
		}
		fmt.Println(params)
		return Write(w, r, http.StatusOK, "test")
	}
	mux := new(ServeMux)
	mux.Handle("GET", "/:name", testHandleFunc)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	_, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Error(err)
	}
}

func TestParamsEmpty(t *testing.T) {
	r := httptest.NewRequest("", "/test", nil)
	if value := Params(r).Get("name"); value != "" {
		t.Error("bad empty params")
	}
}

func TestRealIP(t *testing.T) {
	r := httptest.NewRequest("", "/test", nil)
	addr := RealIP(r)
	fmt.Println(addr)
}
