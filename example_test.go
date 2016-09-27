package rest_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/mdigger/rest"
)

var (
	settings = rest.Standard
	w        = &Recorder{httptest.NewRecorder()}
	r        = httptest.NewRequest("", "/test", nil)
)

type Recorder struct {
	*httptest.ResponseRecorder
}

func (rec *Recorder) Write(buf []byte) (int, error) {
	recorder := rec.ResponseRecorder
	n, err := recorder.Write(buf)
	// result := recorder.Result()
	io.Copy(os.Stdout, recorder.Body)
	w = &Recorder{httptest.NewRecorder()}
	r = httptest.NewRequest("", "/test", nil)
	// restore default settings
	rest.Default = new(rest.Settings)
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

func ExampleSettings() {
	var settings = &rest.Settings{
		Headers:      map[string]string{"X-API-Version": "2.1"},
		Encoder:      rest.JSONEncoder(true),
		Preprocessor: rest.Preprocessor,
		OnError:      func(err error) { log.Println("Error:", err) },
		OnComplete: func(w http.ResponseWriter, r *http.Request,
			status int, data interface{}) {
			log.Printf("<- [%03d] %#v", status, data)
		},
	}

	http.Handle("/rest", settings.Handler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			rest.Write(w, r, http.StatusOK, rest.JSON{"date": time.Now()})
		})))
}

func ExampleSettings_Handler() {
	http.Handle("/rest", rest.Standard.Handler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			rest.Write(w, r, http.StatusOK, rest.JSON{"date": time.Now()})
		})))
}

func ExampleSettings_standard() {
	rest.Standard.Headers["X-My-Header"] = "My header"
	rest.Standard.OnError = func(err error) { log.Println("Error:", err) }
	rest.Standard.OnComplete = func(w http.ResponseWriter, r *http.Request,
		status int, data interface{}) {
		log.Printf("<- [%03d] %#v", status, data)
	}
	http.Handle("/rest", rest.Standard.Handler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			rest.Write(w, r, http.StatusOK, rest.JSON{"date": time.Now()})
		})))
}

func ExampleSettings_default() {
	rest.Default = rest.Standard

	rest.Write(w, r, http.StatusOK, rest.JSON{
		"string": "test",
		"int":    10,
		"bool":   true,
	})
	// Output:
	// {
	//     "code": 200,
	//     "status": "OK",
	//     "success": true,
	//     "data": {
	//         "bool": true,
	//         "int": 10,
	//         "string": "test"
	//     }
	// }
}

func ExamplePreprocessor() {
	rest.Default.Preprocessor = func(w http.ResponseWriter, r *http.Request,
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

	// ...
	// func(w http.ResponseWriter, r *http.Request) {
	rest.Write(w, r, http.StatusOK, rest.JSON{
		"string": "test",
		"int":    10,
		"bool":   true,
	})
	// }

	// Output:
	// {"code":200,"data":{"bool":true,"int":10,"string":"test"},"success":true}
}

func ExampleJSON() {
	rest.Write(w, r, http.StatusOK, rest.JSON{
		"string": "test",
		"int":    10,
		"bool":   true,
	})

	// Output: {"bool":true,"int":10,"string":"test"}
}

func ExampleJSONEncoder() {
	// indent JSON
	rest.Default.Encoder = rest.JSONEncoder(true)

	rest.Write(w, r, http.StatusOK, rest.JSON{
		"string": "test",
		"int":    10,
		"bool":   true,
	})
	// Output:
	// {
	//     "bool": true,
	//     "int": 10,
	//     "string": "test"
	// }
}

func ExampleJSONBind() {
	if r.Method == http.MethodPost {
		var obj = make(rest.JSON)
		err := rest.JSONBind(r, &obj)
		if err != nil {
			rest.Write(w, r, http.StatusBadRequest, err)
			return
		}
	}
}
