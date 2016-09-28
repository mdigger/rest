package rest_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/mdigger/rest"
)

var (
	w = &Recorder{httptest.NewRecorder()}
	r = httptest.NewRequest("", "/test", nil)
)

type Recorder struct {
	*httptest.ResponseRecorder
}

func (rec *Recorder) Write(buf []byte) (int, error) {
	recorder := rec.ResponseRecorder
	n, err := recorder.Write(buf)
	io.Copy(os.Stdout, recorder.Body)
	w = &Recorder{httptest.NewRecorder()}
	r = httptest.NewRequest("", "/test", nil)
	return n, err
}

func ExampleWrite() {
	rest.Write(w, r, http.StatusOK, rest.JSON{
		"string": "test",
		"int":    10,
		"bool":   true,
	})
	// Output: {"bool":true,"int":10,"string":"test"}
}

func ExampleRedirect() {
	rest.Redirect(w, r, http.StatusFound, fmt.Sprintf("/obj/%d", 468))
	// Output: {"redirect":"/obj/468"}
}

func ExampleParams() (int, error) {
	name := rest.Params(r).Get("name")
	data := rest.JSON{"name": name}
	return rest.Write(w, r, http.StatusOK, data)
}

func ExampleDataAdapter() {
	opt := &rest.Options{
		DataAdapter: func(w http.ResponseWriter, r *http.Request,
			status int, data interface{}) (int, interface{}) {
			dataEnvelope := rest.JSON{"code": status}
			if err, ok := data.(error); ok {
				dataEnvelope["error"] = err.Error()
				dataEnvelope["success"] = false
			} else {
				dataEnvelope["data"] = data
				dataEnvelope["success"] = true
			}
			return status, dataEnvelope
		},
		Encoder:       rest.JSONEncoder(true),
		AllowMultiple: false,
	}
	mux := new(rest.ServeMux)
	mux.Options = opt
}
