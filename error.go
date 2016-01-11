package rest

import "net/http"

// Взведенный флаг Debug указывает, что описания ошибок возвращаются как есть.
// В противном случае всегда возвращается только стандартное описание статуса
// HTTP, сформированное на базе этой ошибки.
var Debug = false

// Error описывает возвращаемую по HTTP ошибку, для которой определен статус
// возврата HTTP.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewError возвращает ошибку с указанным кодом окончания запроса. Эту
// ошибку можно вернуть в качестве ошибки выполнения обработки в Context. Но
// если вы хотите сохранить и сам текст ошибки, то лучше использовать
// инициализацию объекта через &HTTPError.
func NewError(code int, err string) error {
	if code < 200 || code >= 600 {
		code = http.StatusInternalServerError
	}
	if err == "" {
		err = http.StatusText(code)
	}
	return Error{Code: code, Message: err}
}

// Error возвращает строковое представление описания ошибки. Если текст ошибки
// не задан, то возвращается текст статуса для данного кода.
//
// Если флаг Debug не взведен, то всегда, вместо описания ошибки, возвращается
// описание кода окончания запроса, а текст ошибки скрывается.
func (e Error) Error() string {
	if e.Message == "" || !Debug {
		if e.Code == 0 {
			return http.StatusText(http.StatusInternalServerError)
		}
		return http.StatusText(e.Code)
	}
	return e.Message
}
