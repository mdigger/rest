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
			"GET":  Func(func(c *Context) { c.Body("login GET") }),
			"POST": Func(func(c *Context) { c.Body("login POST") }),
		},
		"/login/:user-id": {
			"GET": Func(func(c *Context) { c.Body(JSON{"user": c.Get("user-id")}) }),
		},
	}
	mux.Handles(handlers)
	mux.BasePath = "/api/v1"
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

}
