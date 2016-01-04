package rest_test

import (
	"fmt"
	"net/http"
	"os"

	"github.com/geotrace/rest"
)

var c = new(rest.Context) // test context

func ExampleContex_DataSet() {
	type myType byte
	var myData myType = 1

	c.DataSet(myData, "Test data")
	str := c.DataGet(myData).(string)
	fmt.Println(str)
	// Output: Test data
}

func ExampleContext_Body() {
	file, err := os.Open("README.md")
	if err != nil {
		panic(err)
	}
	c.ContentType = "text/markdown; charset=UTF-8"
	c.Body(file) // отдаст содержимое файла
}

func ExampleContext_Code() {
	c.Code(404).Body(nil)
}

func ExampleContext_ParseBody() {
	obj := make(map[string]interface{})
	if err := c.ParseBody(&obj); err != nil {
		panic(err)
	}
}

func ExampleContext_SetHeader() {
	c.SetHeader("ETag", "ab0138")
}

func ExampleHandlers() {
	var mux rest.ServeMux
	mux.Handles(rest.Handlers{
		"/user/:id": {
			"GET": rest.HandlerFunc(func(c *rest.Context) {
				c.Body(rest.JSON{"user": c.Get("id")})
			}),
			"POST": rest.HandlerFunc(func(c *rest.Context) {
				var data = make(rest.JSON)
				if err := c.ParseBody(&data); err != nil {
					c.Code(500).Body(err)
					return
				}
				c.Body(rest.JSON{
					"user": c.Get("id"),
					"data": data,
				})
			}),
		},
		"/message/:text": {
			"GET": rest.HandlerFunc(func(c *rest.Context) {
				c.Body(rest.JSON{"message": c.Get("text")})
			}),
		},
	})
}

func ExampleServeMux_Handle() {
	var mux rest.ServeMux
	mux.Handle("GET", "/message/:text", rest.HandlerFunc(func(c *rest.Context) {
		c.Body(rest.JSON{"message": c.Get("text")})
	}))
}

func ExampleServeMux_Handler() {
	var mux rest.ServeMux
	mux.Handler("GET", "/tmpfiles/",
		http.StripPrefix("/tmpfiles/", http.FileServer(http.Dir("/tmp"))))
}

func ExampleServeMux_ServeHTTP() {
	var mux rest.ServeMux
	mux.Handle("GET", "/message/:text", rest.HandlerFunc(func(c *rest.Context) {
		c.Body(rest.JSON{"message": c.Get("text")})
	}))
	http.ListenAndServe(":8080", mux)
}
