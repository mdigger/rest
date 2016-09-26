package rest

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSettings(t *testing.T) {
	settings := &Settings{
		OnComplete: func(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
			// log.Printf("<- [%03d] %#v", status, data)
			if status != 200 || data == nil {
				t.Error("bad on complete")
			}
		},
	}
	ts := httptest.NewServer(settings.Handler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Write(w, r, 200, JSON{"data": true})
		})))
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	// data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("%s", data)
}

func TestSettingsOnError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("", "/test", nil)

	settings := &Settings{
		OnError: func(err error) {
			if err.Error() != "json: unsupported type: complex128" {
				t.Error("bad on error")
			}
		},
	}
	ctx := context.WithValue(r.Context(), keySettings, settings)
	r = r.WithContext(ctx)

	Write(w, r, 200, complex(1, 76))
	resp := w.Result()
	if status := resp.StatusCode; status != 200 {
		t.Errorf("bad status code: %v", status)
	}

}
