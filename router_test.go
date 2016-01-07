package rest

import (
	"fmt"
	"strings"
	"testing"
)

var urls = []string{
	"/user",
	"/user/test",
	"/user/:id",
	"/user/:id/param",
	"///user/:id/param1/param2/:id/param3//",
}

func TestSplit(t *testing.T) {
	for _, url := range urls {
		splitted := split(url)
		fmt.Printf("%q\n", splitted)
		if strings.Join(splitted, "/") != strings.Trim(url, "/") {
			t.Error(url, splitted)
		}
	}
}

func TestRouter(t *testing.T) {
	var r router
	for _, url := range urls {
		if err := r.add(url, url); err != nil {
			t.Error(err)
		}
	}
	for _, url := range urls {
		handler, params := r.lookup(url)
		if handler == nil {
			t.Error("Nil handler:", url)
		}
		fmt.Println(handler, params)
	}
	url := "/user/:id/param1/"
	handler, params := r.lookup(url)
	if handler != nil {
		t.Error("Bad handler:", url)
	}
	fmt.Println(handler, params)
}
