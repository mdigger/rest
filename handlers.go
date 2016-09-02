package rest

import (
	"fmt"
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
		return c.ServeFile(name)
	}
}

// Files отдает файлы по имени из указанного каталога. Имя файла задается
// в пути в виде последнего именованного параметра. Не выводит список файлов,
// если запрос направлен на каталог файлов, в отличии от стандартной функции
// http.FileServer.
func Files(dir string) Handler {
	return func(c *Context) error {
		filename := filepath.Join(dir, c.params[len(c.params)-1].Value)
		return c.ServeFile(filename)
	}
}

// Data постоянно отдает указанные в параметрах данные в виде ответа на запрос.
func Data(data interface{}, contentType string) Handler {
	return func(c *Context) error {
		c.ContentType = contentType
		return c.Send(data)
	}
}

// NotImplemented возвращает ошибку ErrNotImplemented.
//
// Иногда при разработке руки сразу не доходят до того, чтобы написать
// полноценный обработчик какого нибудь запроса. В этом случае очень выручает
// данная функция, которую можно использовать вместо временной "заплатки".
func NotImplemented(*Context) error {
	return ErrNotImplemented
}

// BasicAuth проверяет HTTP Basic авторизацию пользователя. В качестве
// аргумента передается функция, принимающая значения логина и пароля
// пользователя, и возвращающая true, если пользователь успешно авторизован.
// Вторым параметром передается строка, которая будет использоваться в
// заголовке авторизации для обозначения раздела.
func BasicAuth(auth func(login, password string) bool, realm string) Handler {
	return func(c *Context) error {
		login, password, ok := c.BasicAuth()
		if auth(login, password) {
			c.SetData(restDataAuthLogin, login)
			return nil
		}
		if ok {
			return c.Send(ErrForbidden)
		}
		if realm == "" {
			realm = "Restricted"
		}
		c.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q", realm))
		return c.Send(ErrUnauthorized)
	}
}

// GetAuthLogin возвращает сохраненный при проверке авторизации с помощью
// BasicAuth логин пользователя. В противном случае возвращает пустую строку.
func GetAuthLogin(c *Context) string {
	return c.Data(restDataAuthLogin).(string)
}

// restDataType используется для сохранения данных в контексте запроса.
type restDataType byte

const (
	_ restDataType = iota
	// используется для сохранения логина авторизованного пользователя
	restDataAuthLogin
)
