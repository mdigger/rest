package rest

import (
	"net/http"
	"strings"
)

// Handler является любая функция, которая принимает Context.
type Handler func(*Context)

// Method описывает список Handle, ассоциированные с HTTP-методами.
type Method map[string]Handler

// ServeMux описывает список обработчиков, ассоциированных с путями запроса и методами.
type ServeMux struct {
	router // обработчики запросов по путям, без учета метода запроса
	// Глобальный обработчик, вызываемый перед всеми заданными обработчиками,
	// если определен.
	CustomHandler func(Handler) Handler
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
		if methods, ok := route.(Method); ok {
			methods[method] = handler // добавляем новый обработчик пути для данного метода
			return
		}
	}
	// обработчик для данного пути не определен
	if err := m.router.add(path, Method{method: handler}); err != nil {
		panic(err)
	}
}

// Handles добавляет определение обработчиков сразу для всех методов для указанного пути.
func (m *ServeMux) Handles(path string, handlers Method) {
	if len(handlers) == 0 {
		return
	}
	if err := m.router.add(path, handlers); err != nil {
		panic(err)
	}
}

// Handler позволяет привязать к нашему описанию стандартный обработчик http.Handler.
func (m *ServeMux) Handler(method, path string, handler http.Handler) {
	m.Handle(method, path, func(c *Context) {
		handler.ServeHTTP(c.Response, c.Request)
	})
}

// HandlerFunc позволяет привязать к нашему описанию стандартный обработчик http.Handler.
func (m *ServeMux) HandlerFunc(method, path string, handler http.HandlerFunc) {
	m.Handle(method, path, func(c *Context) {
		handler(c.Response, c.Request)
	})
}

// Lookup возвращает обработчик для указанного пути и метода, а так же заполненный список
// параметров пути. Если обработчик для данных параметров не определен, то возвращается nil.
func (m *ServeMux) Lookup(method, path string) (h Handler, params Params) {
	route, params := m.router.lookup(path)
	if route == nil {
		return
	}
	// в роутере хранятся обработчики с привязкой к методам
	if methods, ok := route.(Method); ok {
		h = methods[strings.ToUpper(method)]
	}
	return
}

// ServeHTTP обеспечивает поддержку интерфейса http.Handler и обрабатывает основной запрос.
func (m *ServeMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route, params := m.router.lookup(req.URL.Path) // получаем обработчик для указанного пути
	context := newContext(w, req, params)          // формируем контекст для ответа
	if route == nil {
		// при статусе больше 399 пустой body формирует JSON с описанием ошибки автоматически
		context.Code(http.StatusNotFound).Body(nil)
		return
	}
	methods, ok := route.(Method) // приводим список методов
	if !ok || len(methods) == 0 { // если методы не определены, то лучше вернем, что путь не найден
		context.Code(http.StatusNotFound).Body(nil)
		return
	}
	handler := methods[strings.ToUpper(req.Method)] // запрашиваем обработчик для метода
	if handler == nil {                             // обработчик для данного метода не определен
		allows := make([]string, 0, len(methods)) // формируем список поддерживаемых методов
		for method := range methods {
			allows = append(allows, method)
		}
		context.SetHeader("Allow", strings.Join(allows, ", "))
		context.Code(http.StatusMethodNotAllowed).Body(nil)
		return
	}
	if m.CustomHandler != nil { // если кастомный обработчик определен, то вызываем его
		m.CustomHandler(handler)(context)
	} else {
		handler(context) // вызываем обработчик запроса
	}
}
