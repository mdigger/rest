package rest

import (
	"net/http/httptest"
	"testing"
)

func TestEncoder(t *testing.T) {
	var enc Encoder = JSONEncoder(true)
	if enc.ContentType(nil, nil) != "application/json; charset=utf-8" {
		t.Error("bad JSON content type")
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/", nil)
	data := JSON{"test": true}
	err := enc.Encode(w, r, data)
	if err != nil {
		t.Error(err)
	}
}
