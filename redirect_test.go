package rest

import (
	"net/http/httptest"
	"testing"
)

func TestRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/test", nil)
	Redirect(w, r, 302, "/new/?test")
	resp := w.Result()
	if status := resp.StatusCode; status != 302 {
		t.Errorf("bad status code: %v", status)
	}
	if location := resp.Header.Get("Location"); location != "/new/?test" {
		t.Errorf("bad location: %v", location)
	}
}

func TestRedirect2(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/test", nil)
	r.URL.Path = ""
	Redirect(w, r, 302, "")
	resp := w.Result()
	if status := resp.StatusCode; status != 302 {
		t.Errorf("bad status code: %v", status)
	}
	if location := resp.Header.Get("Location"); location != "/" {
		t.Errorf("bad location: %v", location)
	}
}
