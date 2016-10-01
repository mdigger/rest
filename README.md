# REST API
[![GoDoc](https://godoc.org/github.com/mdigger/rest?status.svg)](https://godoc.org/github.com/mdigger/rest)
[![Build Status](https://travis-ci.org/mdigger/rest.svg)](https://travis-ci.org/mdigger/rest)
[![Coverage Status](https://coveralls.io/repos/github/mdigger/rest/badge.svg?branch=master)](https://coveralls.io/github/mdigger/rest?branch=master)

Package rest is designed for creating RESTful APIs.

- supports the restriction of only one response to the request
- have the ability to override the data format before output
- it is possible to set your headers to output
- separate support for handling redirects
- ready for logging responses
- support for named parameters in the path

## Example
```go
package main

import (
	"net/http"

	"github.com/mdigger/log"
	"github.com/mdigger/rest"
)

func getUser(w http.ResponseWriter, r *http.Request) (int, error) {
	name := rest.Params(r).Get("name")
	data := rest.JSON{"name": name}
	return rest.Write(w, r, http.StatusOK, data)
}

func getFile(w http.ResponseWriter, r *http.Request) (int, error) {
	name := rest.Params(r).Get("filename")
	return rest.ServeFile(name)
}

func main() {
	mux := &rest.ServeMux{
		Headers: map[string]string{
			"X-API-Version": "1.0",
		},
		NotCompress: false,
		Options: &rest.Options{
			Encoder:       rest.JSONEncoder(true),
			DataAdapter:   rest.Adapter,
			AllowMultiple: false,
		},
		Debug:  true,
		Logger: log.Default,
	}
	mux.Handle("GET", "/user/:name", getUser)
	mux.Handle("GET", "/files/*filename", getFile)

	http.ListenAndServe(":8080", mux)
}
```
