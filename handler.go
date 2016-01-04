package rest

import (
	"net/http"
	"strings"
)

// Handler описывает интерфейс обработчиков, которые могут принимать Context.
type Handler interface {
	ServeHTTPC(*Context) // вызывается для обработки запросов
}

// HandlerFunc является любая функция, которая принимает Context.
type HandlerFunc func(*Context)

// ServeHTTPC поддерживает интерфейс Handler для упрощенных функций обработки.
func (f HandlerFunc) ServeHTTPC(c *Context) { f(c) }

// Methods описывает список Handler, ассоциированные с HTTP-методами.
type Methods map[string]Handler

// Handlers позволяет описать сразу несколько обработчиков для разных путей и методов: ключем для
// данного словаря как раз являются пути запросов. Используется в качестве аргумента при вызове
// метода ServeMux.Handles.
type Handlers map[string]Methods

// ServeMux описывает список обработчиков, ассоциированных с путями запроса и методами.
type ServeMux struct {
	// Глобальный обработчик, вызываемый перед всеми заданными обработчиками,
	// если определен.
	CustomHandler func(Handler) Handler
	router        // обработчики запросов по путям, без учета метода запроса
}

// Handles добавляет определение обработчиков сразу для всех методов для указанного пути, что
// иногда является достаточно удобным вариантом.
func (m *ServeMux) Handles(handlers Handlers) {
	for path, methods := range handlers {
		if len(methods) > 0 {
			if err := m.router.add(path, methods); err != nil {
				panic(err)
			}
		}
	}
}

// Handle добавляет новый обработчик для указанного пути и метода запроса.
func (m *ServeMux) Handle(method, path string, handler Handler) {
	if method == "" || handler == nil {
		return
	}
	method = strings.ToUpper(method) // хочу быть уверенным
	// предполагаем, что обработчик для данного пути уже есть
	if route, _ := m.router.lookup(path); route != nil {
		// в роутере хранятся обработчики с привязкой к методам
		if methods, ok := route.(Methods); ok {
			methods[method] = handler // добавляем новый обработчик пути для данного метода
			return
		}
	}
	// обработчик для данного пути не определен
	m.Handles(Handlers{path: Methods{method: handler}})
}

// Handler позволяет привязать к нашему описанию стандартный обработчик http.Handler.
func (m *ServeMux) Handler(method, path string, handler http.Handler) {
	m.Handle(method, path, HandlerFunc(func(c *Context) {
		handler.ServeHTTP(c.Response, c.Request)
	}))
}

// ServeHTTPC поддерживает интерфейс Handler и отвечает за основную обработку запроса.
func (m ServeMux) ServeHTTPC(context *Context) {
	// получаем обработчик для указанного пути
	route, params := m.router.lookup(context.Request.URL.Path)
	context.Params = append(context.Params, params...)
	if route == nil {
		// при статусе больше 399 пустой body формирует JSON с описанием ошибки автоматически
		context.Code(http.StatusNotFound).Body(nil)
		return
	}
	methods, ok := route.(Methods) // приводим список методов
	if !ok || len(methods) == 0 {  // если методы не определены, то лучше вернем, что путь не найден
		context.Code(http.StatusNotFound).Body(nil)
		return
	}
	handler := methods[strings.ToUpper(context.Request.Method)] // запрашиваем обработчик для метода
	if handler == nil {                                         // обработчик для данного метода не определен
		allows := make([]string, 0, len(methods)) // формируем список поддерживаемых методов
		for method := range methods {
			allows = append(allows, method)
		}
		context.SetHeader("Allow", strings.Join(allows, ", "))
		context.Code(http.StatusMethodNotAllowed).Body(nil)
		return
	}
	if m.CustomHandler != nil { // если кастомный обработчик определен, то вызываем его
		m.CustomHandler(handler).ServeHTTPC(context)
	} else {
		handler.ServeHTTPC(context) // вызываем обработчик запроса
	}
}

// ServeHTTP обеспечивает поддержку интерфейса http.Handler и обрабатывает основной запрос.
func (m ServeMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.ServeHTTPC(newContext(w, req, nil)) // формируем контекст для ответа
}
