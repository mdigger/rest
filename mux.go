package rest

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mdigger/log"
	"github.com/mdigger/router"
)

// ServeMux is an HTTP request multiplexer. It matches the URL of each incoming
// request against a list of registered patterns and calls the handler for the
// pattern that most closely matches the URL.
//
// You can use the named path elements:
// 	/user/:name
// 	/user/:name/files
// 	/user/:name/files/*filename
//
// If a subtree has been registered and a request is received naming the subtree
// root without its trailing slash, ServeMux redirects that request to the
// subtree root (adding the trailing slash). This behavior can be overridden
// with a separate registration for the path without the trailing slash. For
// example, registering "/images/" causes ServeMux to redirect a request for
// "/images" to "/images/", unless "/images" has been registered separately.
type ServeMux struct {
	Headers     map[string]string // additional http headers
	*Options                      // write options
	NotCompress bool              // disallow compression of the response
	Logger      *log.Context      // logger
	routers     map[string]*router.Paths
}

// Handle registers the handler for the given method and pattern.
func (mux *ServeMux) Handle(method, pattern string, handler Handler) {
	if handler == nil {
		panic("rest: nil handler")
	}
	if method == "" {
		method = "GET"
	}
	if mux.routers == nil {
		// typically no more than 9 of HTTP methods
		mux.routers = make(map[string]*router.Paths, 9)
	}
	method = strings.ToUpper(method)
	r := mux.routers[method]
	if r == nil {
		r = new(router.Paths)
		mux.routers[method] = r
	}
	if err := r.Add(pattern, handler); err != nil {
		panic(err) // the handler does not suit us for some reason
	}
}

// ServeHTTP implements http.Handler interface.
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ctxlog *log.Context // log request context
		code   int          // status code
		err    error        // handler error
	)
	// initialize logging
	if mux.Logger != nil {
		started := time.Now()
		ctxlog = mux.Logger
		defer func() {
			ctxlog = ctxlog.WithFields(log.Fields{
				"code":     code,
				"duration": time.Since(started),
			})
			msg := fmt.Sprintf("%s %s", r.Method, r.RequestURI)
			switch {
			case err != nil:
				ctxlog.WithError(err).Error(msg)
			case code < 400:
				ctxlog.Info(msg)
			case code < 500:
				ctxlog.Warning(msg)
			default:
				ctxlog.Error(msg)
			}

		}()
	}

	// add HTTP headers
	if len(mux.Headers) > 0 {
		responseHeader := w.Header()
		for key, value := range mux.Headers {
			responseHeader.Set(key, value)
		}
	}

	// add gzip compression
	if !mux.NotCompress &&
		strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		// Delete this header so gzipping isn't repeated later in the chain
		r.Header.Del("Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()
		w = gzipResponseWriter{Writer: gzipWriter, ResponseWriter: w}
		if ctxlog != nil {
			ctxlog = ctxlog.WithField("gzip", true)
		}
	}

	// add write options to request context
	if mux.Options != nil {
		ctx := context.WithValue(r.Context(), keyOptions, mux.Options)
		r = r.WithContext(ctx)
	}

	// lookup handler for method and path
	var urlPath = r.URL.Path
	if routers := mux.routers[r.Method]; routers != nil {
		if handler, params := routers.Lookup(urlPath); handler != nil {
			if len(params) > 0 { // add params to request context
				ctx := context.WithValue(r.Context(), keyParams, params)
				r = r.WithContext(ctx)
			}
			// execute handler
			fnHandler := handler.(Handler)
			code, err = fnHandler(w, r)
			if code == 0 {
				code = http.StatusOK
				Write(w, r, code, nil)
			} else if code >= 400 {
				Write(w, r, code, nil)
			}
			// if err != nil && mux.Debug && ctxlog != nil {
			// 	// add handler name to log
			// 	name := runtime.FuncForPC(
			// 		reflect.ValueOf(fnHandler).Pointer()).Name()
			// 	ctxlog = ctxlog.WithField("handler", name)
			// }
			return
		}
		// try add slash at the end
		if !strings.HasSuffix(urlPath, "/") {
			urlPath += "/"
			if handler, _ := routers.Lookup(urlPath); handler != nil {
				code, err = Redirect(w, r, http.StatusMovedPermanently, urlPath)
				if ctxlog != nil {
					ctxlog = ctxlog.WithField("url", urlPath)
				}
				return
			}
		}
	}

	// handler for request method not found
	var methods = make([]string, 0, len(mux.routers))
	for method, handlers := range mux.routers {
		if handler, _ := handlers.Lookup(urlPath); handler != nil {
			methods = append(methods, method)
		}
	}
	if len(methods) > 0 {
		// allowed other methods
		w.Header().Set("Allow", strings.Join(methods, ", "))
		code, err = Write(w, r, http.StatusMethodNotAllowed, nil)
		if ctxlog != nil {
			ctxlog = ctxlog.WithField("allowed", methods)
		}
		return
	}

	// handler not found for all methods
	code, err = Write(w, r, http.StatusNotFound, nil)
}
