package rest

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
)

func TestResponse(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "deflate,gzip")
	w := httptest.NewRecorder()
	resp := &response{
		ResponseWriter: w,
		code:           http.StatusOK,
		writer:         w,
		request:        r,
	}
	resp.WriteHeader(200)
	_, err := resp.Write([]byte(`<html>test</html>`))
	if err != nil {
		t.Error(err)
	}
	if _, ok := resp.writer.(*gzip.Writer); !ok {
		t.Error("bad compression")
	}
	resp.Close()

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Error("bad compression header")
	}
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Error("bad content-type header")
	}
	if w.Code != 200 {
		t.Error("bad status code:", w.Code)
	}

	r = httptest.NewRequest("HEAD", "/", nil)
	r.Header.Set("Accept-Encoding", "deflate,gzip")
	w = httptest.NewRecorder()
	resp = &response{
		ResponseWriter: w,
		code:           http.StatusOK,
		writer:         w,
		request:        r,
	}
	resp.WriteHeader(201)
	_, err = resp.Write([]byte(`<html>test</html>`))
	if err != nil {
		t.Error(err)
	}
	if _, ok := resp.writer.(*gzip.Writer); ok {
		t.Error("bad compression on head")
	}
	resp.Close()
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Error("bad compression header")
	}
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Error("bad content-type header")
	}
	if w.Code != 201 {
		t.Error("bad status code:", w.Code)
	}

	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "deflate,gzip")
	w = httptest.NewRecorder()
	resp = &response{
		ResponseWriter: w,
		code:           http.StatusOK,
		writer:         w,
		request:        r,
	}
	_, err = resp.Write(nil)
	if err != nil {
		t.Error(err)
	}
	if _, ok := resp.writer.(*gzip.Writer); ok {
		t.Error("bad compression on nil")
	}
	resp.Close()
	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Error("bad compression header")
	}
	if w.Header().Get("Content-Type") != "" {
		t.Error("bad content-type header")
	}
	if w.Code != 200 {
		t.Error("bad status code:", w.Code)
	}

	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "deflate,gzip")
	w = httptest.NewRecorder()
	resp = &response{
		ResponseWriter: w,
		code:           http.StatusOK,
		writer:         w,
		request:        r,
	}
	resp.Close()
	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Error("bad compression header")
	}
	if w.Header().Get("Content-Type") != "" {
		t.Error("bad content-type header")
	}
	if w.Code != 204 {
		t.Error("bad status code:", w.Code)
	}

	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "deflate,gzip")
	w = httptest.NewRecorder()
	resp = &response{
		ResponseWriter: w,
		code:           http.StatusOK,
		writer:         w,
		request:        r,
	}
	_, err = resp.Write([]byte(`<html>`))
	if err != nil {
		t.Error(err)
	}
	resp.Flush()
	_, err = resp.Write(bytes.Repeat([]byte(`test `), 1000))
	if err != nil {
		t.Error(err)
	}
	resp.Flush()
	_, err = resp.Write(bytes.Repeat([]byte("<p>test</p>\n"), 1000))
	if err != nil {
		t.Error(err)
	}
	resp.Flush()
	_, err = resp.Write([]byte(`</html>`))
	if err != nil {
		t.Error(err)
	}
	resp.Close()
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Error("bad compression header")
	}
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Error("bad content-type header")
	}
	if w.Code != 200 {
		t.Error("bad status code:", w.Code)
	}

	// dump, err := httputil.DumpRequest(r, true)
	// if err != nil {
	// 	t.Error(err)
	// }
	// fmt.Println(string(dump))
	dump, err := httputil.DumpResponse(w.Result(), true)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(dump))

}
