package rest

import (
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
type ServeMux struct {
	Headers map[string]string // additional http.Headers
	Encoder Encoder           // data Encoder (used default if nil)
	Logger  *log.Context      // access logger (if not nil)
	routers map[string]*router.Paths
}

// Handle registers the handler for the given method and pattern. If you specify
// multiple handlers, they will be run sequentially until one of them does not
// return a non-zero response status code or error. When you specify the path
// pattern, you can use named parameters.
func (mux *ServeMux) Handle(method, pattern string, handlers ...Handler) {
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
	if err := r.Add(pattern, Handlers(handlers...)); err != nil {
		panic(err) // the handler does not suit us for some reason
	}
}

// Handler is responsible for the selection of the handler and its
// implementation.
//
// If the handler for the given path and method was not found, but there are
// handlers for other methods, it returns the ErrMethodNotAllowed and the header
// is passed the list of methods that can be applied to the given path.
// Otherwise, returns the ErrNotFound.
func (mux *ServeMux) Handler(c *Context) (err error) {
	// add HTTP headers
	if len(mux.Headers) > 0 {
		for key, value := range mux.Headers {
			c.SetHeader(key, value)
		}
	}
	// lookup handler for method and path
	var (
		urlPath = c.Request.URL.Path
		method  = c.Request.Method
	)
	if routers := mux.routers[method]; routers != nil {
		if handler, params := routers.Lookup(urlPath); handler != nil {
			c.params = append(c.params, params...)
			return handler.(Handler)(c) // execute the request handler
		}

		// try add/remove slash at the end
		if strings.HasSuffix(urlPath, "/") {
			urlPath = strings.TrimSuffix(urlPath, "/")
		} else {
			urlPath += "/"
		}
		if handler, _ := routers.Lookup(urlPath); handler != nil {
			code := http.StatusMovedPermanently
			if method != "GET" && method != "HEAD" {
				code = http.StatusPermanentRedirect
			}
			return c.Redirect(code, urlPath)
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
		c.SetHeader("Allow", strings.Join(methods, ", "))
		return ErrMethodNotAllowed
	}
	return ErrNotFound
}

// ServeHTTP implements http.Handler interface.
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var started = time.Now()
	var context = newContext(w, r)
	context.Encoder = mux.Encoder
	err := mux.Handler(context)
	if !context.IsWrote() {
		context.Write(err)
	}
	context.close()
	// output information to the log
	if mux.Logger != nil {
		ctxlog := mux.Logger.WithFields(context.logFields)
		code := context.Status()
		ctxlog = ctxlog.WithFields(log.Fields{
			"code":     code,
			"duration": time.Since(started),
			"gzip":     context.Compressed(),
			"size":     context.ContentLength(),
			"ip":       context.RealIP(),
		})
		msg := fmt.Sprintf("%s %s",
			context.Request.Method, context.Request.RequestURI)
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
	}
}

type (
	// Paths allows to describe multiple handlers for different ways
	// and methods: the key for this dictionary are path queries.
	// Used as argument when calling the method ServeMux.Handles.
	Paths map[string]Methods
	// Methods allows one to describe handlers for methods.
	Methods map[string]Handler
)

// Handles adds from the list of handlers for multiple ways and methods.
// It is, in fact, just a convenient way to immediately identify a large number
// of handlers, without causing every time ServeMux.Handle.
//
// Optionally, you can specify the list of handlers to be executed before
// performance of the specified.
func (mux *ServeMux) Handles(paths Paths, middleware ...Handler) {
	for path, methods := range paths {
		for method, handler := range methods {
			// add middleware for all handlers if they are defined
			if len(middleware) > 0 {
				handler = Handlers(append(middleware, handler)...)
			}
			mux.Handle(method, path, handler)
		}
	}
}
