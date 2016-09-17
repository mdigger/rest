package rest

import (
	"errors"
	"net/http"
	"os"
)

// Эти ошибки обрабатываются при передаче их в метод Context.Send и
// устанавливают соответствующий статус ответа.
//
// Кроме указанных здесь ошибок, так же проверяется, что ошибка отвечает на
// os.IsNotExist (в этом случае статус станет 404) или os.IsPermission (статус
// 403). Все остальные ошибки устанавливают статус 500.
//
// Если вам нет необходимости указывать собственное сообщение для вывода ошибки,
// то проще всего воспользоваться этим предопределенными, использовав их в
// context.Send():
// 	return c.Send(ErrNotfound)
var (
	ErrDataAlreadySent       = errors.New("data already sent")
	ErrBadRequest            = errors.New("bad request")              // 400
	ErrUnauthorized          = errors.New("unauthorized")             // 401
	ErrForbidden             = errors.New("forbidden")                // 403
	ErrNotFound              = errors.New("not found")                // 404
	ErrLengthRequired        = errors.New("length required")          // 411
	ErrRequestEntityTooLarge = errors.New("request entity too large") // 413
	ErrUnsupportedMediaType  = errors.New("unsupported media type")   // 415
	ErrInternalServerError   = errors.New("internal server error")    // 500
	ErrNotImplemented        = errors.New("not implemented")          // 501
	ErrServiceUnavailable    = errors.New("service unavailable")      // 503
)

// setErrorStatus устанавливает статус ответа в зависимости от ошибки и ее типа.
func (c *Context) setErrorStatus(err error) {
	if err == nil || c.sended || c.status != 0 {
		return
	}
	// устанавливаем статус, в зависимости от ошибки
	switch err {
	case ErrBadRequest: // 400
		c.status = http.StatusBadRequest
	case ErrUnauthorized: // 401
		c.status = http.StatusUnauthorized
	case ErrForbidden: // 403
		c.status = http.StatusForbidden
	case ErrNotFound: // 404
		c.status = http.StatusNotFound
	case ErrLengthRequired: // 411
		c.status = http.StatusLengthRequired
	case ErrRequestEntityTooLarge: // 413
		c.status = http.StatusRequestEntityTooLarge
	case ErrUnsupportedMediaType: // 415
		c.status = http.StatusUnsupportedMediaType
	case ErrInternalServerError: // 500
		c.status = http.StatusInternalServerError
	case ErrNotImplemented: // 501
		c.status = http.StatusNotImplemented
	case ErrServiceUnavailable: // 503
		c.status = http.StatusServiceUnavailable
	default:
		if os.IsNotExist(err) {
			c.status = http.StatusNotFound
		} else if os.IsPermission(err) {
			c.status = http.StatusForbidden
		} else {
			c.status = http.StatusInternalServerError
		}
	}
}
