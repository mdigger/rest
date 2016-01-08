package rest

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestContext(t *testing.T) {
	var mux ServeMux
	mux.Handles(Paths{
		"/test/": {
			"GET": func(c *Context) {
				c.Data("fff")
				c.SetHeader("Use", "")
				c.Send(c.Get("id"))
			},
		},
		"/error/": {
			"GET": func(c *Context) {
				c.Error(errors.New("test error"))
			},
		},
		"/bytes/": {
			"GET": func(c *Context) {
				c.Send([]byte("OK"))
			},
		},
		"/stream/": {
			"GET": func(c *Context) {
				file, err := os.Open("README.md")
				if err != nil {
					c.Error(err)
					return
				}
				c.ContentType = `text/markdown`
				c.Send(file)
			},
		},
		"/object/": {
			"POST": func(c *Context) {
				var o = make(JSON)
				if err := c.Parse(o); err != nil {
					c.Error(err)
					return
				}
				c.Send(o)
			},
		},
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	res, err := http.Get(ts.URL + mux.BasePath + "/test?id=test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = http.Get(ts.URL + mux.BasePath + "/error/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = http.Get(ts.URL + mux.BasePath + "/bytes/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = http.Get(ts.URL + mux.BasePath + "/stream/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println()
	fmt.Println(strings.Repeat("-", 40))

	res, err = http.Post(ts.URL+mux.BasePath+"/object/", "application/json",
		strings.NewReader(`{test:"message"}`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))
}
