package rest

import (
	"net/http"
	"os"
	"path/filepath"
)

// HTTPError устанавливает статус ответа и возвращает ошибку.
func HTTPError(err string, code int) Handler {
	return func(c *Context) error {
		return NewError(code, err)
	}
}

// NotFound возвращает в ответ ошибку NotFound.
func NotFound() Handler {
	return HTTPError(http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

// Redirect возвращает перенаправляющий ответ.
func Redirect(url string, code int) Handler {
	return func(c *Context) error {
		http.Redirect(c, c.Request, url, code)
		return nil
	}
}

// StaticData описывает обработчик запроса со статическим ответом.
func StaticData(data interface{}, contentType string) Handler {
	return func(c *Context) error {
		if contentType != "" {
			c.ContentType = contentType
		}
		return c.Send(data)
	}
}

// ServeFile отдает статический файл в ответ на запрос. Если файл с таким именем
// не найден, то отдается ошибка.
func ServeFile(filename string) Handler {
	fi, err := os.Stat(filename) // проверяем, что файл существует и доступен
	if err == nil && fi.IsDir() {
		err = os.ErrNotExist // не позволяем обращаться к каталогам
	}
	return func(c *Context) error {
		if err != nil {
			return err
		}
		http.ServeFile(c, c.Request, filename)
		return nil
	}
}

// ServeParamFile отдает файл с указанным в параметре именем. В качестве аргумента передается
// имя параметра и путь к каталогу с файлами.
func ServeParamFile(param, path string) Handler {
	return func(c *Context) error {
		filename := c.Param(param)
		// проверяем, что файл существует и доступен
		fi, err := os.Stat(filepath.Join(path, filename))
		if err != nil {
			return err
		}
		if fi.IsDir() { // не позволяем обращаться к каталогам
			return os.ErrNotExist
		}
		http.ServeFile(c, c.Request, filename) // отдаем содержимое файла
		return nil
	}
}
