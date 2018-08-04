package rest

import (
	"net/http"
	"path/filepath"
)

// Handler describes an HTTP request handler.
type Handler func(*Context) error

// ServeHTTP implements http.Handler interface.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	if err := h(c); !c.IsWrote() {
		c.Write(err)
	}
	c.close()
}

// Handlers combines multiple query processors in the queue. They will be
// executed in the order in which were added one after another until they are
// all done or until you return the first error, that interrupts the process
// further processing. As well further processing is interrupted if the handler
// gave the response to the client. Using this feature allows you to combine
// multiple processors into one.
func Handlers(handlers ...Handler) Handler {
	switch len(handlers) {
	case 0:
		return nil
	case 1:
		return handlers[0]
	default:
		return func(c *Context) error {
			for _, h := range handlers {
				if h == nil {
					continue
				}
				if err := h(c); err != nil {
					return err
				}
				if c.IsWrote() {
					break
				}
			}
			return nil
		}
	}
}

// Redirect returns a Handler which performs a permanent redirect specified in
// the URL parameters.
func Redirect(url string) Handler {
	return func(c *Context) error {
		return c.Redirect(http.StatusMovedPermanently, url)
	}
}

// File returns the file content.
func File(name string) Handler {
	return func(c *Context) error { return c.ServeFile(name) }
}

// Files gives the files from the specified directory. The file name is set
// in the way as the last named parameter. Does not display the list of files
// if the request is for a file directory in contrast to standard functions
// http.FileServer.
func Files(dir string) Handler {
	return func(c *Context) error {
		if len(c.params) == 0 {
			return ErrNotFound
		}
		filename := filepath.Join(dir, c.params[len(c.params)-1].Value)
		return c.ServeFile(filename)
	}
}

// Data constantly gives specified in the settings data in response to the
// request.
func Data(data interface{}, contentType string) Handler {
	return func(c *Context) error {
		c.SetContentType(contentType)
		return c.Write(data)
	}
}

// ErrorHandler returns the handler of http requests, which always returns
// the specified error.
func ErrorHandler(err *Error) Handler {
	return func(*Context) error { return err }
}

// Predefined error handlers.
var (
	NotFound       = ErrorHandler(ErrNotFound)
	NotImplemented = ErrorHandler(ErrNotImplemented)
)

// HTTPHandler преобразует http.Handler в Handler.
func HTTPHandler(h http.Handler) Handler {
	return func(c *Context) error {
		h.ServeHTTP(c.Response, c.Request)
		return nil
	}
}

// HTTPFiles обеспечивает отдачу файлов с помощью http.Dir и аналогичных
// вещей, которые поддерживают интерфейс http.FileSystem.
func HTTPFiles(files http.FileSystem, index string) Handler {
	return func(c *Context) error {
		if len(c.params) == 0 {
			return ErrNotFound
		}
		// получаем последний параметр, определенный в пути
		var name = c.params[len(c.params)-1].Value
		// подставляем имя для корневого элемента
		if name == "" {
			name = index
		}
		// отдаем содержимое файла
		file, err := files.Open(name)
		if err != nil {
			return err
		}
		defer file.Close()
		fi, err := file.Stat()
		if err != nil {
			return err
		}
		return c.ServeContent(name, fi.ModTime(), file)
	}
}
