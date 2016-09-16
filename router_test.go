package rest

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/kr/pretty"
)

func TestRouter3(t *testing.T) {
	log.SetFlags(log.Lshortfile)
	var r router
	var err error
	h1 := func(c *Context) error { return nil }
	err = r.add("/:name/test", h1)
	if err != nil {
		t.Error(err)
	}
	h2 := func(c *Context) error { return nil }
	err = r.add("/:name/store/*filename", h2)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(r.path(h1))
	fmt.Println(r.path(h2))
	pretty.Println(r)
	handler, params := r.lookup("/name1/test")
	// fmt.Println(r.path(handler), params)
	fmt.Println(params)
	handler, params = r.lookup("/name2/store")
	// fmt.Println(r.path(handler), params)
	fmt.Println(params)
	handler, params = r.lookup("/name3/store/test")
	// fmt.Println(r.path(handler), params)
	fmt.Println(params)
	_ = handler
}

func TestRouter(t *testing.T) {
	tests := []struct {
		url     string
		bad     bool
		handler Handler
	}{
		{"/test/:p1/:p2", false, func(c *Context) error { return nil }},
		{"/test/:p1/:p2/*p3", false, func(c *Context) error { return nil }},
		{"/:p1/:p2/*p3", false, func(c *Context) error { return nil }},
		{"/p1/p2/p3/p4", false, func(c *Context) error { return nil }},
		{"p2/*p1", false, func(c *Context) error { return nil }},
		{"p2/*p1/:bad", true, func(c *Context) error { return nil }},
	}
	// список урлов и номер обработчика
	urls := map[string]int{
		"/test/param1/param2":               0,
		"/test/param1/param2/param3/param4": 1,
		"/test/param1/param2/param3":        1,
		"/p1/p2/p3/p4":                      3,
		"/p2/p3/p4":                         2,
		"/p2/p3/p4/p5":                      2,
	}
	var r router
	for i, test := range tests {
		err := r.add(test.url, test.handler)
		if test.bad && err == nil {
			t.Errorf("error in add bad test %d", i)
		} else if !test.bad && err != nil {
			t.Errorf("error in add test %d", i)
		}
	}
	for i, test := range tests {
		path := r.path(test.handler)
		if test.bad && path != nil {
			t.Errorf("error in path bad test %d", i)
		} else if !test.bad && path == nil {
			t.Errorf("error in path test %d", i)
		}
	}
	for url, num := range urls {
		handler, _ := r.lookup(url)
		if num < 0 && handler != nil {
			t.Errorf("bad test lookup for url %q", url)
		} else {
			h1 := runtime.FuncForPC(reflect.ValueOf(tests[num].handler).Pointer()).Entry()
			h2 := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Entry()
			if h1 != h2 {
				t.Errorf("bad handler for test lookup url %q", url)
			}
		}
	}

	// pretty.Println(r)
}

func TestRouter2(t *testing.T) {
	urls := []string{
		"/p1/p2/p3/:p4",
		"/p1/p2/:p3/p4",
		"/p1/:p2/p3/p4",
		"/:p1/p2/p3/p4",
		"/:p1/p2/p3/:p4",
		"/:p1/p2/:p3/p4",
		"/:p1/p2/:p3/:p4",
		"/:p1/:p2/p3/p4",
		"/:p1/:p2/p3/:p4",
		"/:p1/:p2/:p3/p4",
		"/:p1/:p2/:p3/:p4",
	}
	var r router
	if h, _ := r.lookup("/test"); h != nil {
		t.Error("bad lookup")
	}
	for _, url := range urls {
		if err := r.add(url, func(*Context) error { return nil }); err != nil {
			t.Error(err)
		}
	}
	if h, _ := r.lookup("/1/2/3/4/5"); h != nil {
		t.Error("bad lookup")
	}
	if err := r.add("/p1/p2/p3/p4/*p5", func(*Context) error { return nil }); err != nil {
		t.Error(err)
	}
	if h, _ := r.lookup("/p1/p2/p3/p4/p5/p6/p7"); h == nil {
		t.Error("bad catch all lookup")
	}
	if err := r.add("/p1/p2/*p5", func(*Context) error { return nil }); err != nil {
		t.Error(err)
	}
	if h, _ := r.lookup("/p1/p2/p3/p4/p5/p6/p7"); h == nil {
		t.Error("bad catch all lookup")
	}
	if h, _ := r.lookup("/p0/p2/p3/p4/p5/p6/p7"); h != nil {
		t.Error("bad lookup")
	}
	if err := r.add("/a1/a2/*a3", func(*Context) error { return nil }); err != nil {
		t.Error(err)
	}
	if h, _ := r.lookup("/a1/a2/a3/a4/a5/a6/a7"); h == nil {
		t.Error("bad lookup a")
	}
	if err := r.add("/a1/a2/a3/a4/a5/a6/:a7", func(*Context) error { return nil }); err != nil {
		t.Error(err)
	}
	if h, _ := r.lookup("/a1/a2/a3/a4/a5/a6"); h == nil {
		t.Error("bad lookup a")
	}
	if h, _ := r.lookup("/a1/a2/a3/a4/a5/a6/a7"); h == nil {
		t.Error("bad lookup a")
	}
	// pretty.Println(r)
}

func TestRouterMax(t *testing.T) {
	var r router
	if err := r.add(strings.Repeat("/:param", 1<<15), func(*Context) error { return nil }); err == nil {
		t.Error("bad max path length")
	}
	if err := r.add("/", nil); err == nil {
		t.Error("bad add nil handler")
	}
}
