package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"strings"
	"testing"
)

func TestContext(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)
	r, err := http.NewRequest("POST", "http://example.com/foo?param=name", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")
	// r.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	c := newContext(w, r)
	defer c.close()
	if c.Data("test") != nil {
		t.Error("bad get data")
	}
	if c.Param("param") != r.URL.Query().Get("param") {
		t.Error("bad param")
	}
	c.SetData("test", "test")
	if c.Data("test").(string) != "test" {
		t.Error("bad set data")
	}
	data := make(map[string]interface{})
	if err := c.Parse(&data); err != nil {
		t.Error(err)
	}
	c.ContentType = "application/json; encoding=utf-8"
	if err := c.Status(200).Send(data); err != nil {
		t.Error(err)
	}
	fmt.Println("HTTP/1.1", w.Code)
	w.HeaderMap.Write(os.Stdout)
	fmt.Println()
	w.Body.WriteTo(os.Stdout)
	fmt.Print("\n", strings.Repeat("-", 40), "\n")
}

func TestContext2(t *testing.T) {
	ts := httptest.NewServer(Handler(func(c *Context) error {
		data := make(JSON)
		if err := c.Parse(&data); err != nil {
			return err
		}
		// return c.Send(data)
		c.ContentType = "text/plain"
		c.SetHeader("X-Powered-By", "")
		err := c.Send("OK")
		c.Flush()
		c.Send("OK")
		return err
	}))
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL, strings.NewReader(`{"key":"value"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(dump))
	fmt.Println()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		t.Error("Dump error:", err)
	}
	fmt.Println(string(dump))
	fmt.Println(strings.Repeat("-", 40))

	req, err = http.NewRequest("POST", ts.URL, strings.NewReader(`{"key":"value"}`))
	if err != nil {
		t.Fatal(err)
	}
	dump, err = httputil.DumpRequest(req, true)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(dump))
	fmt.Println()
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		t.Error("Dump error:", err)
	}
	fmt.Println(string(dump))
	fmt.Println(strings.Repeat("-", 40))
}

func TestContextLoop(t *testing.T) {
	// CompressResponse = true
	for i := 0; i < 10; i++ {
		TestContext2(t)
	}
}

func TestContext4(t *testing.T) {
	ts := httptest.NewServer(Handler(func(c *Context) error {
		c.ContentType = "text/plain"
		return c.Send([]byte("OK"))
	}))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(dump))
	// fmt.Println()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(dump))
	fmt.Println(strings.Repeat("-", 40))
}

func TestContextReader(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)
	r, err := http.NewRequest("POST", "http://example.com/foo?param=name", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c := newContext(w, r)
	defer c.close()

	file, err := os.Open("context_test.go")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Send(file); err != nil {
		t.Fatal(err)
	}
}

func TestContextNil(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)
	r, err := http.NewRequest("POST", "http://example.com/foo?param=name", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c := newContext(w, r)
	defer c.close()

	if err := c.Send(nil); err != nil {
		t.Fatal(err)
	}
}

func TestContextErrorNotExists(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)
	r, err := http.NewRequest("POST", "http://example.com/foo?param=name", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c := newContext(w, r)
	defer c.close()

	if err := c.Send(os.ErrNotExist); err != nil {
		t.Fatal(err)
	}
}

func TestContextErrorPermission(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)
	r, err := http.NewRequest("POST", "http://example.com/foo?param=name", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c := newContext(w, r)
	defer c.close()

	if err := c.Send(os.ErrPermission); err != nil {
		t.Fatal(err)
	}
}

func TestContextError(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)
	r, err := http.NewRequest("POST", "http://example.com/foo?param=name", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c := newContext(w, r)
	defer c.close()

	if err := c.Send(os.ErrInvalid); err != nil {
		t.Fatal(err)
	}
}

func TestContextHTTPError(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)
	r, err := http.NewRequest("POST", "http://example.com/foo?param=name", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c := newContext(w, r)
	defer c.close()

	if err := c.Send(&Error{}); err != nil {
		t.Fatal(err)
	}
}

func TestContext404Nil(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)
	r, err := http.NewRequest("POST", "http://example.com/foo?param=name", body)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c := newContext(w, r)
	defer c.close()

	if err := c.Status(404).Send(nil); err != nil {
		t.Fatal(err)
	}
}
