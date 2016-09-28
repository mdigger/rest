package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestWrite(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/test", nil)

	Write(w, r, 200, JSON{"test": true})
	resp := w.Result()
	if status := resp.StatusCode; status != 200 {
		t.Errorf("bad status code: %v", status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Error(err)
	}
	var obj = make(JSON)
	err = json.Unmarshal(data, &obj)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiWrite(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/test", nil)

	Write(w, r, 200, JSON{"test": true})
	defer func() {
		p := recover()
		if str, ok := p.(string); !ok || str != "rest: multiple responses" {
			t.Error("bad double response check")
		}
	}()
	// must be panic
	Write(w, r, 200, JSON{"test": true})
	t.Error("after panic")
}

func TestWriteEncodeError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/test", nil)

	_, err := Write(w, r, 200, complex(1, 76))
	if err == nil {
		t.Error("bad encoder error check")
	}
}
