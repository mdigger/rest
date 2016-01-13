package rest

import (
	"net/http"
	"path/filepath"
)

// Redirect возвращает Handler, который осуществляет постоянное перенаправление
// на указанный в параметрах URL.
func Redirect(url string) Handler {
	return func(c *Context) error {
		http.Redirect(c, c.Request, url, http.StatusMovedPermanently)
		return nil
	}
}

// File отдает на запрос содержимое файла с указанным именем.
func File(name string) Handler {
	return func(c *Context) error {
		http.ServeFile(c, c.Request, name)
		return nil
	}
}

// Files отдает файлы по имени из указанного каталога. Имя файла задается
// в пути в виде последнего именованного параметра.
func Files(dir string) Handler {
	return func(c *Context) error {
		filename := filepath.Join(dir, c.params[len(c.params)-1].Value)
		http.ServeFile(c, c.Request, filename)
		return nil
	}
}

// Data постоянно отдает указанные в параметрах данные в виде ответа на запрос.
func Data(data interface{}, contentType string) Handler {
	return func(c *Context) error {
		c.ContentType = contentType
		return c.Send(data)
	}
}
