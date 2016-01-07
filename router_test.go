package rest

import (
	"fmt"
	"strings"
	"testing"
)

var urls = []string{
	"///user/:id/param1/param2/:id/param3//",
	"/user/:id/param",
	"/user/:id",
	"/user/test",
	"/user/:id/param2",
	"/user/vova/param2",
	"/user/vova/:param",
	"/user",
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
<<<<<<< HEAD
=======
	handler, params = r.lookup("/user/test/mama/1/2/3/4/5/6/7/8/9/0/")
	if handler != nil {
		t.Error("Bad handler:", url)
	}
}

func TestLongRouter(t *testing.T) {
	var r router
	url := `1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/` +
		`1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/` +
		`1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/` +
		`1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/` +
		`1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/` +
		`1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/1/2/3/4/5/6/7/8/9/0/`
	if err := r.add(url, "long"); err == nil {
		t.Error("must be error")
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
>>>>>>> mdigger/master
}
