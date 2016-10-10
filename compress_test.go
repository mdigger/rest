package rest

import "testing"

func TestCompress(t *testing.T) {
	for _, data := range []struct {
		mime     string
		compress bool
	}{
		{"text/plain", true},
		{"text/css", true},
		{"text/javascript", true},
		{"text/xml", true},
		{"text/html", true},
		{"application/javascript; charset=utf-8", true},
		{"application/javascript; charset=utf-8", true},
		{"application/manifest+json", true},
		{"application/rdf+xml", true},
		{"application/rss+xml", true},
		{"application/atom+xml", true},
		{"application/vnd.geo+json", true},
		{"application/x-javascript", true},
		{"application/x-web-app-manifest+json", true},
		{"application/xhtml+xml", true},
		{"application/xml", true},
		{"font/eot", true},
		{"font/opentype", true},
		{"image/bmp", true},
		{"image/svg+xml", true},
		{"image/vnd.microsoft.icon", true},
		{"image/x-icon", true},
		{"application/pdf", false},
		{"application/octet-stream", false},
		{"application/ogg", false},
		{"application/postscript", false},
		{"application/soap+xml", true},
		{"application/zip", false},
		{"application/gzip", false},
		{"application/x-bittorrent", false},
		{"audio/basic", true},
		{"audio/mp4", false},
		{"audio/aac", false},
		{"audio/mpeg", false},
		{"audio/vorbis", false},
		{"image/gif", false},
		{"image/jpeg", false},
		{"image/png", false},
		{"image/tiff", false},
		{"video/mpeg", false},
		{"application/x-www-form-urlencoded", false},
	} {
		if isCompress(data.mime) != data.compress {
			t.Error("bad compress flag:", data.mime)
		}
	}

	AddCompressMimeType("test", "*pattern")
	if !isCompress("test/x-pattern") {
		t.Error("bad registering new mime-type pattern")
	}
}
