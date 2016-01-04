package rest

import (
	"net/http"
	"net/url"
	"strings"
)

// Handler является любая функция, которая принимает Context.
type Handler func(*Context)

// Middleware описывает вспомогательные обработчики, которые могут использоваться
// в качестве конвейера обработки запросов.
type Middleware func(Handler) Handler

// Methods описывает список Handler, ассоциированные с HTTP-методами.
type Methods map[string]Handler

// Paths позволяет описать сразу несколько обработчиков для разных путей и методов: ключем для
// данного словаря как раз являются пути запросов. Используется в качестве аргумента при вызове
// метода ServeMux.Handles.
type Paths map[string]Methods

// ServeMux описывает список обработчиков, ассоциированных с путями запроса и методами.
//
// Внутри используется достаточно простой и быстрый алгоритм выбора обработчиков, основанный
// на количестве элементов пути (между разделителями '/'). К сожалению, данный алгоритм не
// позволяет использовать catch-all ('*') параметры и использовать вложенные мультиплексоры:
// они просто не будут корректно работать.
//
// Еще одним ограничением является количество элементов пути: на текущий момент поддерживается
// не более 255 таких элементов. Если в пути встретится большее количество элементов, то при
// добавлении такого пути будет вызвана panic.
type ServeMux struct {
	// позволяет задать базовый путь для всех запросов
	// данный путь "отрезается" и не используется при вычислении обработчика
	BasePath string
	// Глобальный обработчик, вызываемый перед всеми заданными обработчиками,
	// если определен.
	Middleware
	router // обработчики запросов по путям, без учета метода запроса
}

// Handles добавляет определение обработчиков сразу для всех методов для указанного пути, что
// иногда является достаточно удобным вариантом.
func (m *ServeMux) Handles(paths Paths) {
	for path, methods := range paths {
		if len(methods) > 0 {
			if err := m.router.add(path, methods); err != nil {
				panic(err)
			}
		}
	}
}

// Handle добавляет новый обработчик для указанного пути и метода запроса. При задании пути
// можно использовать именованные параметры (начинаются с символа ':'). В дальнейшем, можно
// будет получить значения этих параметров, спросив их по имени через метод Context.Get("name").
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
	m.Handles(Paths{path: Methods{method: handler}})
}

// Handler позволяет привязать к нашему описанию стандартный обработчик http.Handler.
// Т.к. стандартные обработчики не имеют доступа к Context, то, соответственно, они не могут
// получить доступ и к именованным параметрам пути. Для того, чтобы хоть как-то облегчить
// работу, такие параметры будут добавлены к URL в виде именованных параметров, так что с ними
// можно будет работать через http.Request.URL.Query().Get("name").
func (m *ServeMux) Handler(method, path string, handler http.Handler) {
	m.Handle(method, path, func(c *Context) {
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
		handler.ServeHTTP(c.Response, c.Request)
	})
}

// ServeHTTP обеспечивает поддержку интерфейса http.Handler и обрабатывает основной запрос.
func (m ServeMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	context := newContext(w, req) // формируем контекст для ответа
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
	if m.Middleware != nil { // если промежуточный обработчик определен, то вызываем его
		handler = m.Middleware(handler)
	}
	handler(context) // вызываем обработчик запроса
}
