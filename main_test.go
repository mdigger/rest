package rest_test

import (
	"log"
	"net/http"
	"time"

	rest "github.com/mdigger/rest"
)

func preprocessor(w http.ResponseWriter, r *http.Request,
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
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	if time.Now().Weekday() == time.Monday {
		rest.Redirect(w, r, http.StatusFound, "http://golang.org/")
		return
	}

	if r.Method == http.MethodPost {
		var obj = make(rest.JSON)
		err := rest.JSONBind(r, &obj)
		if err != nil {
			rest.Write(w, r, http.StatusBadRequest, err)
			return
		}
	}

	data := rest.JSON{
		"string": "test",
		"int":    10,
		"bool":   true,
	}
	rest.Write(w, r, http.StatusOK, data)
}

func Example() {
	opts := &rest.Settings{
		Headers:      map[string]string{"X-API-Version": "2.1"},
		Preprocessor: preprocessor,
		OnError:      func(err error) { log.Println("Error:", err) },
		OnComplete: func(w http.ResponseWriter, r *http.Request,
			status int, data interface{}) {
			log.Printf("<- [%03d] %@v", status, data)
		},
	}
	http.Handle("/foo", opts.Handler(http.HandlerFunc(fooHandler)))
	http.ListenAndServe("http", nil)
}
