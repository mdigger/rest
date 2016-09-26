package rest

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestJSONBind(t *testing.T) {
	data, err := json.Marshal(JSON{"test": true})
	if err != nil {
		t.Fatal(err)
	}
	body := bytes.NewReader(data)

	r := httptest.NewRequest("POST", "/", body)
	var test = make(JSON)
	err = JSONBind(r, &test)
	if err.Error() != "unknown content-type" {
		t.Error("empty content-type")
	}

	r.Header.Set("Content-type", "application/json")
	err = JSONBind(r, &test)
	if err != nil {
		t.Error(err)
	}

	r.Header.Set("Content-type", `application/json; charset="windows-1251"`)
	err = JSONBind(r, &test)
	if err.Error() != "unsupported charset windows-1251" {
		t.Error("unsupported charset")
	}

	r.Header.Set("Content-type", "application/xml")
	err = JSONBind(r, &test)
	if err.Error() != "unsupported content-type application/xml" {
		t.Error("unsupported charset")
	}
}
