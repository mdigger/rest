package rest

import (
	"net/http"
	"strings"

	"github.com/mdigger/router"
)

// ServeMux описывает список обработчиков, ассоциированных с путями запроса и
// методами. По сути, это замена http.ServeMux с поддержкой параметров в пути
// запроса.
type ServeMux struct {
	// Описывает дополнительные заголовки HTTP-ответа, которые будут добавлены
	// ко всем ответам, возвращаемым данным обработчиком
	Headers map[string]string
	routers map[string]*router.Paths // обработчики запросов по методам
}

// ServeHTTP обеспечивает поддержку интерфейса http.Handler. Таким образом,
// данный ServeMux можно использовать как стандартный обработчик HTTP-запросов.
//
// В процессе обработки запроса отслеживаются возвращаемые ошибки и
// перехватываются возможные вызовы panic. Если ответ на запрос еще не
// отправлялся, то в этих случаях в ответ будет отправлена ошибка.
func (m *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r) // формируем контекст для ответа
	defer func() {
		// перехватываем panic, если она случилась
		if e := recover(); e != nil {
			// если еще ничего не отсылали, то отсылаем эту ошибку
			c.Send(e)
			// // выводим дамп с ошибкой
			// c.errorLog(e, 4) // записываем ошибку в лог
		}
		c.close() // освобождаем по окончании
	}()
	// вызываем обработку запроса
	if err := m.Handler(c); err != nil {
		// c.errorLog(err, 1) // записываем ошибку в лог
		// если ничего не отправляли, то отправляем эту ошибку
		if !c.sended {
			c.Send(err) // возвращаемые ошибки игнорируем
		}
	}
}

// Handler отвечает за подбор обработчика и его выполнение.
//
// Если обработчик для данного пути и метода не найден, но есть обработчики
// для других методов, то возвращается статус http.StatusMethodNotAllowed и в
// заголовке передается список методов, которые можно применить к данному пути.
// В противном случае возвращается статус http.StatusNotFound.
func (m *ServeMux) Handler(c *Context) error {
	// добавляем заголовки, если они определены
	header := c.response.Header()
	if len(m.Headers) > 0 {
		for key, value := range m.Headers {
			header.Set(key, value)
		}
	}
	// получаем список обработчиков для данного метода
	if routers := m.routers[c.Request.Method]; routers != nil {
		// запрашиваем подходящий обработчик
		if handler, params := routers.Lookup(c.path); handler != nil {
			// добавляем найденные параметры к контексту
			c.params = append(c.params, params...)
			// вызываем обработчик запроса
			return handler.(Handler)(c)
		}
	}
	// обработчик для данного пути не найден
	// собираем список методов, которые поддерживаются для данного пути
	methods := make([]string, 0, len(m.routers))
	for method, handlers := range m.routers {
		if handler, _ := handlers.Lookup(c.path); handler != nil {
			methods = append(methods, method)
		}
	}
	if len(methods) > 0 {
		// если есть обработчики для данного пути, но с другими методами,
		// то отдаем этот список методов
		header.Set("Allow", strings.Join(methods, ", "))
		return c.Status(http.StatusMethodNotAllowed).Send(nil)
	}
	// обработчики пути не определены ни для одного метода
	return c.Send(ErrNotFound)
}

// Handle регистрирует обработчик для указанного метода и пути.
// В описании пути можно использовать именованные параметры (начинаются с
// символа ':') и завершающий именованный параметр (начинается с '*'), который
// указывает, что path может быть длиннее. В последнем случае вся остальная
// часть пути будет включена в данный параметр. Параметр со звездочкой, если
// указан, должен быть самым последним параметром пути.
//
// Если количество элементов пути в path больше 32768 или параметр со звездочкой
// используется не в самом последнем элементе пути, то возникает panic.
func (m *ServeMux) Handle(method, path string, handlers ...Handler) {
	handler := Handlers(handlers...)    // собираем все в один обработчик
	if method == "" || handler == nil { // игнорируем пустые обработчики
		return
	}
	// если список обработчиков еще не инициализирован, то инициализируем его
	if m.routers == nil {
		// обычно используется не более 9 методов HTTP
		m.routers = make(map[string]*router.Paths, 9)
	}
	// получаем список обработчиков для данного метода
	method = strings.ToUpper(method)
	r := m.routers[method]
	if r == nil {
		r = new(router.Paths)
		m.routers[method] = r
	}
	// добавляем обработчик для заданного метода и пути
	if err := r.Add(path, handler); err != nil {
		panic(err) // обработчик нас не устраивает по каким-то причинам
	}
}

// Handle создает новый ServeMux, добавляет в него обработчики для указанного
// пути и метода, и возвращает его.
func Handle(method, path string, handlers ...Handler) *ServeMux {
	var mux ServeMux
	mux.Handle(method, path, handlers...)
	return &mux
}

// Handles добавляет сразу список обработчиков для нескольких путей и методов.
// Это, по сути, просто удобный способ сразу определить большое количество
// обработчиков, не вызывая каждый раз ServeMux.Handle.
//
// Дополнительно можно указать список обработчиков, который будет выполнен
// перед выполнение заданных.
func (m *ServeMux) Handles(paths Paths, handlers ...Handler) {
	for path, methods := range paths {
		for method, handler := range methods {
			// добавляем middleware ко всем обработчикам, если они определены
			if len(handlers) > 0 {
				handler = Handlers(append(handlers, handler)...)
			}
			m.Handle(method, path, handler)
		}
	}
}

// Handles возвращает новый инициализированный ServeMux c заданными
// обработчиками HTTP-запросов.
func Handles(paths Paths, handlers ...Handler) *ServeMux {
	var mux ServeMux
	mux.Handles(paths, handlers...)
	return &mux
}

type (
	// Paths позволяет описать сразу несколько обработчиков для разных путей
	// и методов: ключем для данного словаря как раз являются пути запросов.
	// Используется в качестве аргумента при вызове метода ServeMux.Handles.
	Paths map[string]Methods
	// Methods позволяет описать обработчики для методов.
	Methods map[string]Handler
)
