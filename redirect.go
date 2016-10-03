package rest

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

// RedirectURL describes the URL to redirect. The Redirect function is used for
// transmission of the URL to switch to Write.
type RedirectURL struct {
	Redirect string `json:"redirect"`
}

// Redirect causes the URL to absolute relative to the current request and
// invokes Write with RedirectURL(urlStr).
func Redirect(w http.ResponseWriter, r *http.Request,
	code int, urlStr string) (int, error) {
	// trying to set relative path
	if u, err := url.Parse(urlStr); err == nil {
		if u.Scheme == "" && u.Host == "" {
			oldpath := r.URL.Path
			if oldpath == "" { // should not happen, but avoid a crash if it does
				oldpath = "/"
			}
			// no leading http://server
			if urlStr == "" || urlStr[0] != '/' {
				// make relative path absolute
				olddir, _ := path.Split(oldpath)
				urlStr = olddir + urlStr
			}
			var query string
			if i := strings.Index(urlStr, "?"); i != -1 {
				urlStr, query = urlStr[:i], urlStr[i:]
			}
			// clean up but preserve trailing slash
			trailing := strings.HasSuffix(urlStr, "/")
			urlStr = path.Clean(urlStr)
			if trailing && !strings.HasSuffix(urlStr, "/") {
				urlStr += "/"
			}
			urlStr += query
		}
	}
	w.Header().Set("Location", urlStr)
	return Write(w, r, code, &RedirectURL{urlStr})
}

// RedirectHandler returns a Handler which gives back to the jump at the
// specified URL.
func RedirectHandler(code int, urlStr string) Handler {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		return Redirect(w, r, code, urlStr)
	}
}
