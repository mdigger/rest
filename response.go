package rest

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// response implements http.ResponseWriter interface, adding support for some
// methods.
type response struct {
	http.ResponseWriter
	code        int
	request     *http.Request
	writer      io.Writer
	wroteHeader bool
	compressed  bool
	written     int64
}

// WriteHeader sets the response status code.
func (rw *response) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.code = code
	}
}

// writeHeader really writes the response status code in the response.
func (rw *response) writeHeader() {
	rw.ResponseWriter.WriteHeader(rw.code) // real writing header status
	rw.wroteHeader = true                  // block rewriting headers
}

// setDataHeaders sets the Content-Type header and configures compression
// support response.
func (rw *response) setDataHeaders(data []byte) {
	if len(data) == 0 {
		return
	}
	var headers = rw.Header() // response headers
	// writing Content-Type
	contentType := headers.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(data)
		headers.Set("Content-Type", contentType)
	}
	// check compression support in request header
	if !strings.Contains(rw.request.Header.Get("Accept-Encoding"), "gzip") ||
		!isCompress(contentType) {
		return
	}
	// remove the header compression support, not to install it again
	rw.request.Header.Del("Accept-Encoding")
	headers.Add("Vary", "Accept-Encoding")
	headers.Set("Content-Encoding", "gzip")
	headers.Del("Content-Length")
	if rw.request.Method == "HEAD" {
		return
	}
	// set gzip writer to response
	var gzw = gzips.Get().(*gzip.Writer)
	gzw.Reset(rw.writer)
	rw.writer = gzw
	rw.compressed = true
}

// Write is responsible for return data in response to the request.
func (rw *response) Write(data []byte) (int, error) {
	if !rw.wroteHeader {
		rw.setDataHeaders(data) // set Content-Type & compression
		rw.writeHeader()        // real writing header status
	}
	if rw.request.Method == "HEAD" {
		return len(data), nil // skip writing on HEAD
	}
	// writing data to response
	n, err := rw.writer.Write(data)
	rw.written += int64(n)
	return n, err
}

// Close terminates the output of the response and frees gzip.Writer if it has
// been initialized for compression response.
func (rw *response) Close() {
	if !rw.wroteHeader {
		rw.code = http.StatusNoContent
		rw.writeHeader()
	}
	if gzw, ok := rw.writer.(*gzip.Writer); ok {
		gzw.Close()
		gzips.Put(gzw)
	}
}

// Flush supports the http.Flusher interface.
func (rw *response) Flush() {
	if !rw.wroteHeader {
		rw.writeHeader()
	}
	if gzw, ok := rw.writer.(*gzip.Writer); ok {
		gzw.Flush()
	}
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// gzips contains pool of gzip.Writer.
var gzips = sync.Pool{New: func() interface{} { return gzip.NewWriter(nil) }}
