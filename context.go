package rest

import (
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/mdigger/log"
	"github.com/mdigger/router"
)

// Context describes the context of processing an http request.
type Context struct {
	Response      http.ResponseWriter         // original http.Response
	Request       *http.Request               // original http.Request
	Encoder       Encoder                     // data encoder
	AllowMultiple bool                        // allow multiple response
	params        router.Params               // path named params
	data          map[interface{}]interface{} // request context data
	query         url.Values                  // url query values
	logFields     []log.Field                 // additional log fields
}

// newContext return new initialized request context.
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Response: &response{
			ResponseWriter: w,
			code:           http.StatusOK,
			writer:         w,
			request:        r,
		},
		Request: r,
	}
}

// close terminates the output of the response and frees gzip.Writer if it has
// been initialized for compression response.
func (c *Context) close() {
	c.Response.(*response).Close()
}

// Header return request header value.
func (c *Context) Header(key string) string {
	return c.Request.Header.Get(key)
}

// SetHeader sets a response header with a given value.
func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

// SetContentType sets response content-type header.
func (c *Context) SetContentType(contentType string) {
	c.SetHeader("Content-Type", contentType)
}

// SetStatus sets response status code.
func (c *Context) SetStatus(code int) {
	c.Response.WriteHeader(code)
}

// Status returns response status code.
func (c *Context) Status() int {
	return c.Response.(*response).code
}

// Compressed returns true if response compression supported.
func (c *Context) Compressed() bool {
	return c.Response.(*response).compressed
}

// ContentLength returns the uncompressed size of the response data.
func (c *Context) ContentLength() int64 {
	return c.Response.(*response).written
}

// IsWrote returns true if the header has already been sent in response to the
// request.
func (c *Context) IsWrote() bool {
	return c.Response.(*response).wroteHeader
}

// SetCookie adds the Cookie in the response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

// Cookie returns the named cookie provided in the request or http.ErrNoCookie
// if not found.
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// Form returns the first value for the named component of the POST or PUT
// request body. URL query parameters are ignored.
func (c *Context) Form(key string) string {
	return c.Request.PostFormValue(key)
}

// Query returns the first value for the named component of the URL query.
func (c *Context) Query(key string) string {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query.Get(key)
}

// FormFile returns the first file for the provided form key.
func (c *Context) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return c.Request.FormFile(key)
}

// BasicAuth returns the username and password provided in the request's
// Authorization header, if the request uses HTTP Basic Authentication.
func (c *Context) BasicAuth() (username, password string, ok bool) {
	return c.Request.BasicAuth()
}

// Bind parses the request and populates the received data specified structure.
// Supported parsing of JSON, XML and HTTP form. For HTTP form in the structure,
// you can use the tag `form:` to specify the name.
func (c *Context) Bind(v interface{}) error {
	return bind(c.Request, v)
}

// JSON is just a quick way to describe data structures.
type JSON = map[string]interface{}

// RedirectURL is used for output redirect.
type RedirectURL struct {
	Code int    `json:"code"`
	URL  string `json:"location"`
}

// Write replies to the request with the specified using Encoder.
func (c *Context) Write(data interface{}) (err error) {
	if c.Response.(*response).wroteHeader && !c.AllowMultiple {
		return ErrMultipleResponse // not supports multiple responses
	}
	// get data Encoder
	var encoder = c.Encoder
	if encoder == nil {
		encoder = defaultEncoder
	}
	// response data
	switch data := data.(type) {
	case nil:
		if c.Status() == http.StatusOK {
			c.SetStatus(http.StatusNoContent)
		}
		_, err = c.Response.Write(nil)
	case []byte:
		_, err = c.Response.Write(data)
	case string:
		_, err = io.WriteString(c.Response, data)
	case io.Reader:
		_, err = io.Copy(c.Response, data)
	case *RedirectURL:
		c.SetStatus(data.Code)
		if newURL, err := c.Request.URL.Parse(data.URL); err == nil {
			data.URL = newURL.String()
		}
		c.SetHeader("Location", data.URL)
		err = encoder(c, data)
	case error:
		var code = http.StatusInternalServerError
		if httperror, ok := data.(*Error); ok {
			code = httperror.Code
		} else if os.IsNotExist(data) {
			code = http.StatusNotFound
		} else if os.IsPermission(data) {
			code = http.StatusForbidden
		} else if timeout, ok := data.(net.Error); ok && timeout.Timeout() {
			code = http.StatusRequestTimeout
		}
		c.SetStatus(code)
		err = encoder(c, data)
	default:
		err = encoder(c, data)
	}
	return err
}

// Error replies to the request with the specified error message and HTTP code.
// The error message should be plain text.
func (c *Context) Error(code int, message string) error {
	var httpError = &Error{Code: code, Message: message}
	if err := c.Write(httpError); err != nil {
		return err
	}
	return httpError
}

// Redirect replies to the request with a redirect to url, which may be a path
// relative to the request path.
//
// The provided code should be in the 3xx range and is usually
// http.StatusMovedPermanently, http.StatusFound or http.StatusSeeOther.
func (c *Context) Redirect(code int, url string) error {
	return c.Write(&RedirectURL{Code: code, URL: url})
}

// ServeContent replies to the request using the content in the provided
// ReadSeeker. The main benefit of ServeContent over io.Copy is that it handles
// Range requests properly, sets the MIME type, and handles If-Modified-Since
// requests.
func (c *Context) ServeContent(name string, modtime time.Time, content io.ReadSeeker) error {
	// TODO: перехватить запись ошибки
	http.ServeContent(c.Response, c.Request, name, modtime, content)
	if code := c.Response.(*response).code; code >= 400 {
		return &Error{Code: code, Message: http.StatusText(code)}
	}
	return nil
}

// ServeFile replies to the request with the contents of the named file.
func (c *Context) ServeFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	fi, _ := file.Stat()
	return c.ServeContent(name, fi.ModTime(), file)
}

// Param returns the value of the named parameter.
func (c *Context) Param(key string) string {
	for _, param := range c.params {
		if param.Key == key {
			return param.Value
		}
	}
	return ""
}

// Params return all param names
func (c *Context) Params() []string {
	list := make([]string, len(c.params))
	for i, param := range c.params {
		list[i] = param.Key
	}
	return list
}

// Data returns the user data stored in the request context with specified key.
// Usually this information is used when you want to pass them between multiple
// processors.
func (c *Context) Data(key interface{}) interface{} {
	if c.data == nil {
		return nil
	}
	return c.data[key]
}

// SetData saves the user data in the request context with the specified key.
//
// Recommended as a key to use a private type and its the value to avoid
// accidental overwrites of the data to other processors: this will certainly
// protect against casual access to them. But strings too supported. :)
func (c *Context) SetData(key, value interface{}) {
	if c.data == nil {
		c.data = make(map[interface{}]interface{})
	}
	c.data[key] = value
}

// AddLogField add named filed to context log.
func (c *Context) AddLogField(key string, value interface{}) {
	c.logFields = append(c.logFields, log.Field{Name: key, Value: value})
}

// RealIP returns a real IP address from headers. To this end, we use the
// headers of the http request "X-Forwarded-For" and "X-Real-IP" or a real
// address from the connection.
func (c *Context) RealIP() string {
	addr := c.Request.RemoteAddr
	if ip := c.Header("X-Forwarded-For"); ip != "" {
		addr = ip
	} else if ip := c.Header("X-Real-IP"); ip != "" {
		addr = ip
	} else {
		addr, _, _ = net.SplitHostPort(addr)
	}
	return addr
}
