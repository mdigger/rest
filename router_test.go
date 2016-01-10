package rest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kr/pretty"
)

var urls = []string{
	"///user/:id/param1/param2/:id/param3//",
	"/user/:id/param",
	"/user/:id",
	"/user/test",
	"/user/vova/:param",
	"/user/:vova/param3",
	"/user/:id/param2",
	"/user",
	"/user/*test",
	"/user/:id/*param2",
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
	pretty.Println(r)
	for _, url := range urls {
		handler, params := r.lookup(url)
		if handler == nil {
			t.Error("Nil handler:", url)
		}
		fmt.Println(handler, params)
	}
	url := "/user/:id/param1/"
	handler, params := r.lookup(url)
	if handler == nil {
		t.Error("Bad handler:", url)
	}
	fmt.Println(handler, params)
	handler, params = r.lookup("/user/test/mama/1/2/3/4/5/6/7/8/9/0/")
	if handler == nil {
		t.Error("Bad handler:", url)
	}
}

func TestOnlyStaticRouter(t *testing.T) {
	var r router
	var urls = []string{
		"test/url",
		"test2/url",
		"test2/add",
		"test/add",
		"test",
	}
	for _, url := range urls {
		if err := r.add(url, "long"); err != nil {
			t.Error(err)
		}
	}
	handler, _ := r.lookup("/test")
	if handler == nil {
		t.Error("Bad handler:", "/test")
	}
	handler, _ = r.lookup("/test2")
	if handler != nil {
		t.Error("Bad handler:", "/test2")
	}
}

func TestRouterSort(t *testing.T) {
	var r router
	var urls = []string{
		"/1/2/*3/",
		"/1/:2/*3/",
		"/:1/2/*3/",
		"/:1/:2/*3/",
		"/1/2/3/",
		"/:1/2/3/",
		"/1/:2/3/",
		"/1/2/:3/",
		"/:1/:2/3/",
		"/:1/2/:3/",
		"/1/:2/:3/",
	}
	for _, url := range urls {
		if err := r.add(url, url); err != nil {
			t.Error(err)
		}
	}
	pretty.Println(r)
}

func TestRouterBad(t *testing.T) {
	var r router
	if err := r.add("/*test/:test", "bad"); err == nil {
		t.Error("bad * param in path")
	}
	if h, _ := r.lookup("/1/2/3/4/5/"); h != nil {
		t.Error("bad handler")
	}
	if err := r.add(strings.Repeat("/:test", 1<<15+1), "bad long"); err == nil {
		t.Error("bad long handler")
	}
	if err := r.add("/:test/:test", "bad"); err != nil {
		t.Error(err)
	}
	if h, _ := r.lookup("/1/2/3/4/5/"); h != nil {
		t.Error("bad handler")
	}
}

func TestRouterDynamic(t *testing.T) {
	var r router
	var urls = []string{
		"/1/2/*3/",
		"/1/:2/*3/",
		"/:1/2/*3/",
		"/1/2/3/",
		"/:1/2/3/",
		"/1/:2/3/",
		"/1/2/:3/",
		"/:1/:2/3/",
		"/:1/2/:3/",
		"/1/:2/:3/",
		// "/1/*2/",
	}
	for _, url := range urls {
		if err := r.add(url, url); err != nil {
			t.Error(err)
		}
	}
	pretty.Println(r)
	pretty.Println(r.lookup("/1/2/3/4/5/6"))
	pretty.Println(r.lookup("/1/2/3/4/5"))
	pretty.Println(r.lookup("/1/2/3/"))
	pretty.Println(r.lookup("/1/2/"))
	pretty.Println(r.lookup("/1/"))
}
