package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestContext(t *testing.T) {
	r := httptest.NewRequest("POST", "/?test=queryParam", strings.NewReader("test=formParam"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// r.Header.Set("X-Real-IP", "127.0.0.1")
	r.SetBasicAuth("username", "password")
	w := httptest.NewRecorder()
	ctx := &Context{
		Response: &response{
			ResponseWriter: w,
			code:           http.StatusOK,
			writer:         w,
			request:        r,
		},
		Request:       r,
		AllowMultiple: true,
	}
	defer ctx.close()

	_, err := ctx.Cookie("test")
	if err != http.ErrNoCookie {
		t.Error("bad cookie")
	}
	ctx.SetCookie(&http.Cookie{
		Name:  "name",
		Value: "value",
	})

	if ctx.Data("key") != nil {
		t.Error("bad data")
	}
	ctx.SetData("key", "value")
	if ctx.Data("key").(string) != "value" {
		t.Error("bad data value")
	}
	ctx.RealIP()
	err = ctx.Write(JSON{"data": "test"})
	if err != nil {
		t.Error(err)
	}
	err = ctx.Write(`{"data": "test"}`)
	if err != nil {
		t.Error(err)
	}
	err = ctx.Write([]byte(`{"data": "test"}`))
	if err != nil {
		t.Error(err)
	}
	err = ctx.Write(strings.NewReader(`{"data": "test"}`))
	if err != nil {
		t.Error(err)
	}
	err = ctx.Write(errors.New("error"))
	if err != nil {
		t.Error(err)
	}
	err = ctx.Write(ErrLengthRequired)
	if err != nil {
		t.Error(err)
	}
	err = ctx.Write(os.ErrPermission)
	if err != nil {
		t.Error(err)
	}
	err = ctx.Write(os.ErrNotExist)
	if err != nil {
		t.Error(err)
	}
	err = ctx.Error(400, "test")
	if err.Error() != "test" {
		t.Error("bad error")
	}
	err = ctx.Write(nil)
	if err != nil {
		t.Error(err)
	}
	err = ctx.Redirect(301, "/test")
	if err != nil {
		t.Error(err)
	}
	if ctx.Query("test") != "queryParam" {
		t.Error("bad query")
	}
	if ctx.Form("test") != "formParam" {
		t.Error("bad form")
	}
	if ctx.Compressed() {
		t.Error("bad compression")
	}
	login, password, ok := ctx.BasicAuth()
	if !ok || login != "username" || password != "password" {
		t.Error("bad basic authorization")
	}
	_, _, err = ctx.FormFile("key")
	if err == nil {
		t.Error("bad file form")
	}
	var test = new(struct {
		Test string
	})
	err = ctx.Bind(test)
	if err != nil {
		t.Error(err)
	}
	if test.Test != "formParam" {
		t.Error("bad bind")
	}

	// dump, err := httputil.DumpResponse(w.Result(), true)
	// if err != nil {
	// 	t.Error(err)
	// }
	// fmt.Println(string(dump))
}

func TestContext_ServeFile(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("X-Real-IP", "127.0.0.1")
	w := httptest.NewRecorder()
	ctx := &Context{
		Response: &response{
			ResponseWriter: w,
			code:           http.StatusOK,
			writer:         w,
			request:        r,
		},
		Request: r,
	}
	defer ctx.close()
	ctx.RealIP()
	ctx.ServeFile("context_test.go")
	if w.Code != 200 {
		t.Error("bad response")
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Error("bad content type:", ct)
	}
	if ctx.Write(nil) != ErrMultipleResponse {
		t.Error("multiply responses")
	}
}
