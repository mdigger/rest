package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPreprocessor(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/test", nil)

	_, obj := Preprocessor(w, r, 200, JSON{"test": true})
	data, ok := obj.(*Response)
	if !ok {
		t.Fatal("bad response type")
	}
	if data.Code != 200 {
		t.Errorf("bad status: %v", data.Code)
	}
	if data.Data == nil {
		t.Error("bad data")
	}

	_, obj = Preprocessor(w, r, 200, errors.New("test error"))
	data, ok = obj.(*Response)
	if !ok {
		t.Fatal("bad response type")
	}
	if data.Code != 200 {
		t.Errorf("bad status: %v", data.Code)
	}
	if data.Error != "test error" {
		t.Errorf("bad error data:", data.Error)
	}

	_, obj = Preprocessor(w, r, 200, nil)
	data, ok = obj.(*Response)
	if !ok {
		t.Fatal("bad response type")
	}
	if data.Code != 200 {
		t.Errorf("bad status: %v", data.Code)
	}
	if data.Data != nil {
		t.Errorf("bad data:", data.Data)
	}
	if data.Error != "" {
		t.Errorf("bad error data:", data.Error)
	}
	if data.Status != http.StatusText(200) {
		t.Errorf("bad status:", data.Status)
	}

	_, obj = Preprocessor(w, r, 301, RedirectURL("/new"))
	data, ok = obj.(*Response)
	if !ok {
		t.Fatal("bad response type")
	}
	if data.Code != 301 {
		t.Errorf("bad status: %v", data.Code)
	}
	if data.Data != nil {
		t.Errorf("bad data:", data.Data)
	}
	if data.Error != "" {
		t.Errorf("bad error data:", data.Error)
	}
	if data.Status != http.StatusText(301) {
		t.Errorf("bad status:", data.Status)
	}
	if string(data.Redirect) != "/new" {
		t.Errorf("bad redirect:", data.Redirect)
	}
}
