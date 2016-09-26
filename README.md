# REST API
[![GoDoc](https://godoc.org/github.com/mdigger/rest?status.svg)](https://godoc.org/github.com/mdigger/rest)
[![Build Status](https://travis-ci.org/mdigger/rest.svg)](https://travis-ci.org/mdigger/rest)
[![Coverage Status](https://coveralls.io/repos/mdigger/rest/badge.svg?branch=master&service=github)](https://coveralls.io/github/mdigger/rest?branch=master)

Package rest is designed for creating RESTful APIs.

- makes it easy to give any objects that support serialization to JSON
- if desired, the output format can be overridden at any other
- supports the restriction of only one response to the request
- can be used with the standard http library
- have the ability to override the data format before output
- it is possible to set your headers to output
- separate support for handling redirects
- ready for logging responses

## Example
```go
package main

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

func main() {
	opts := &rest.Settings{
		Headers:      map[string]string{"X-API-Version": "2.1"},
		Preprocessor: preprocessor,
		OnError:      func(err error) { log.Println("Error:", err) },
		OnComplete: func(w http.ResponseWriter, r *http.Request,
			status int, data interface{}) {
			log.Printf("<- [%03d] %#v", status, data)
		},
	}
	http.Handle("/foo", opts.Handler(http.HandlerFunc(fooHandler)))
	http.ListenAndServe("http", nil)
}
```
