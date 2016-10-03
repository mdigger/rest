package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/test", nil)
	Redirect(w, r, 302, "/new/?test")
	resp := w.Result()
	if code := resp.StatusCode; code != 302 {
		t.Errorf("bad status code: %v", code)
	}
	if location := resp.Header.Get("Location"); location != "/new/?test" {
		t.Errorf("bad location: %v", location)
	}
}

func TestRedirectEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/test", nil)
	r.URL.Path = ""
	Redirect(w, r, 302, "")
	resp := w.Result()
	if code := resp.StatusCode; code != 302 {
		t.Errorf("bad status code: %v", code)
	}
	if location := resp.Header.Get("Location"); location != "/" {
		t.Errorf("bad location: %v", location)
	}
}

func TestRedirectHndler(t *testing.T) {
	mux := new(ServeMux)
	mux.Handle("GET", "/", RedirectHandler(301, "https://www.connector73.com/en#softphone"))
	ts := httptest.NewServer(mux)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp
	// dump, err := httputil.DumpResponse(resp, true)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%s", dump)
}
