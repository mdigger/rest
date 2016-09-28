package rest

import (
	"compress/gzip"
	"io"
	"net/http/httptest"
	"testing"
)

func TestGzip(t *testing.T) {
	w := httptest.NewRecorder()
	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close()
	gzw := gzipResponseWriter{Writer: gzipWriter, ResponseWriter: w}
	gzw.WriteHeader(200)
	io.WriteString(gzw, "<html>test</html")

}
