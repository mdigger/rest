package rest

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"testing"
)

var _ DataAdapter = Adapter

func jsonDataOutput(data interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

func TestAdapter(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/", nil)
	for _, testdata := range []struct {
		code int
		data interface{}
	}{
		{200, "string"},
		{301, RedirectURL{"http://example.com/"}},
		{302, RedirectURL{"/example"}},
		{404, nil},
		{500, errors.New("error")},
	} {
		code, data := Adapter(w, r, testdata.code, testdata.data)
		if code != testdata.code {
			t.Error("bad returned code:", testdata.code)
		}
		if d, ok := data.(*Response); !ok || d == nil {
			t.Error("bad returned data:", data)
		}
		jsonDataOutput(data)
	}
}
