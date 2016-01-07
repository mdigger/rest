package rest

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	var mux ServeMux
	handlers := Paths{
		"/login": {
			"GET":  func(c *Context) { c.Send("login GET") },
			"POST": func(c *Context) { c.Send("login POST") },
		},
		"/login/:user-id": {
			"GET": func(c *Context) { c.Send(JSON{"user": c.Get("user-id")}) },
		},
		"/test-query": {
			"GET": func(c *Context) { c.Send(JSON{"query": c.Get("param")}) },
		},
	}
	mux.Handles(handlers)
	mux.Handle("GET", "/test", func(c *Context) { c.Send("OK") })
	mux.Handler("GET", "/test/:name", http.NotFound)
	mux.BasePath = "/api/v1"
	mux.Middleware = func(h Handler) Handler {
		return h
	}
	ts := httptest.NewTLSServer(mux)
	defer ts.Close()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	res, err := client.Get(ts.URL + mux.BasePath + "/login")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Post(ts.URL+mux.BasePath+"/login", "", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Get(ts.URL + mux.BasePath + "/login/dmitrys")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Get(ts.URL + mux.BasePath + "/test-query?param=param-name")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Get(ts.URL + mux.BasePath + "/test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Get(ts.URL + mux.BasePath + "/test/test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Get(ts.URL + mux.BasePath + "/test/test/test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Get(ts.URL + "/test/test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Post(ts.URL+mux.BasePath+"/test", "", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

}
