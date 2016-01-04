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
	mux.Handles(Handlers{
		"/login": Method{
			"GET": HandlerFunc(func(c *Context) {
				c.Body("login GET")
			}),
			"POST": HandlerFunc(func(c *Context) {
				c.Body("login POST")
			}),
		},
		"/login/:user-id": Method{
			"GET": HandlerFunc(func(c *Context) {
				c.Body(map[string]string{"user": c.Get("user-id")})
			}),
		},
	})
	ts := httptest.NewTLSServer(mux)
	defer ts.Close()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	res, err := client.Get(ts.URL + "/login")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Post(ts.URL+"/login", "", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

	fmt.Println(strings.Repeat("-", 40))

	res, err = client.Get(ts.URL + "/login/dmitrys")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s ", res.Request.Method, res.Request.URL.Path)
	res.Write(os.Stdout)

}
