package rest

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

// RedirectURL describes the type of the string containing the URL to redirect.
type RedirectURL string

// Redirect causes the URL to absolute relative to the current request and
// invokes Write with RedirectURL(urlStr).
func Redirect(w http.ResponseWriter, r *http.Request, status int, urlStr string) {
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
	Write(w, r, status, RedirectURL(urlStr))
}