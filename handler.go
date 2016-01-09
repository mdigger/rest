package rest

import (
	"fmt"
	"net/http"
	"net/url"
)

// Handler является любая функция, которая принимает Context. Если в результате возвращается
// ошибка и ответа от сервера еще не передавалось, то эта ошибка будет возвращена в качестве
// ошибки сервера.
//
// Функция автоматически отслеживает типы ошибок и для некоторых из них может изменять статус
// возврата ответа. В частности, если ошибка соответствует os.IsNotExist, то вернется 404 ошибка.
// А если соответствует os.IsPermission, то — 403.
//
// Если Handler используется совместно с ServeMux, то вы можете самостоятельно переопределить
// коды возврата для ошибок, задав функцию ServeMux.Errors.
type Handler func(*Context) error

// ServeHTTP поддерживает интерфейс http.Handler для Handler, что позволяет использовать его
// с любыми совместимыми с http.Handler библиотеками.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context := newContext(w, r)        // инициализируем новый контекст запроса
	context.coder = defaultCoder       // используем кодировщик по умолчанию
	if err := h(context); err != nil { // выполняем обработчик
		context.Send(err) // пытаемся отослать ошибку, если еще ничего не отдавали
	}
	context.close() // освобождаем контекст запроса и помещаем его обратно в пул
}

// handler приводит разные виды обработчиков к интерфейсу Handler.
func handler(handlr interface{}) Handler {
	switch h := handlr.(type) {
	case Handler: // приводить тип не требуется
		return h
	case http.HandlerFunc: // стандартную http-функцию оборачиваем в простой обработчик
		return func(c *Context) error {
			c.sended = true // мы не управляем отдачей содержимого ответа
			// т.к. стандартные обработчики не могут получить доступ к именованным параметрам пути,
			// то добавляем их как именованные параметры в запрос
			if len(c.Params) > 0 {
				urlQuery := make(url.Values, len(c.Params))
				for _, param := range c.Params {
					urlQuery.Add(param.Key, param.Value)
				}
				p := urlQuery.Encode()
				if c.Request.URL.RawQuery != "" {
					p += "&" + c.Request.URL.RawQuery
				}
				c.Request.URL.RawQuery = p
			}
			h(c, c.Request) // вызываем функцию, передав ей запрос и ответ
			return nil      // возвращаем, что ошибок нет
		}
	case http.Handler: // стандартный обработчик используем как простую http-функцию
		return handler(h.ServeHTTP) // сводим задачу к вышеуказанной
	default: // не поддерживаемый тип обработчика
		panic(fmt.Errorf("unsupported Handler type %T", handlr))
	}
}

// Handlers объединяет несколько обработчиков Handler в очередь.
func Handlers(handlers ...interface{}) Handler {
	return func(c *Context) error {
		for _, h := range handlers {
			if err := handler(h)(c); err != nil {
				return err
			}
		}
		return nil
	}
}
