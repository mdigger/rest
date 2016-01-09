package rest

import "net/http"

// Взведенный флаг Debug указывает, что описания ошибок возвращаются как есть.
// В противном случае всегда возвращается только стандартное описание статуса
// HTTP, сформированное на базе этой ошибки.
var Debug = false

// HTTPError описывает возвращаемую по HTTP ошибку, для которой определен статус.
type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewHTTPError возвращает ошибку с указанным кодом окончания запроса. Это ошибку можно вернуть
// в качестве ошибки выполнения обработки в Context.
func NewHTTPError(code int) *HTTPError {
	if code < 200 || code >= 600 {
		code = http.StatusInternalServerError
	}
	return &HTTPError{Code: code, Message: http.StatusText(code)}
}

// Error возвращает строковое представление описания ошибки. Если текст ошибки не задан, то
// возвращается текст статуса для данного кода.
//
// Если флаг Debug не взведен, то всегда, вместо описания ошибки, возвращается описание кода
// окончания запроса, а текст ошибки скрывается.
func (e HTTPError) Error() string {
	if e.Message == "" || !Debug {
		if e.Code == 0 {
			return http.StatusText(http.StatusInternalServerError)
		}
		return http.StatusText(e.Code)
	}
	return e.Message
}
