package rest

import (
	"net/http"
	"strings"
)

type (
	// Paths позволяет описать сразу несколько обработчиков для разных путей
	// и методов: ключем для данного словаря как раз являются пути запросов.
	// Используется в качестве аргумента при вызове метода ServeMux.Handles.
	Paths map[string]Methods
	// Methods позволяет описать обработчики для методов: ключем является как
	// раз HTTP-метод, а значение — любой поддерживаемый тип обработчика. На
	// сегодняшний момент поддерживаются следующие типы обработчиков:
	// Handler, http.Handler и http.HandlerFunc.
	Methods map[string]interface{}
)

// ServeMux описывает список обработчиков, ассоциированных с путями запроса и
// методами.
type ServeMux struct {
	// Позволяет задать базовый путь для всех запросов. Данный путь "отрезается"
	// и не используется при вычислении обработчика.
	BasePath string
	// Описывает дополнительные заголовки HTTP-ответа, которые будут добавлены
	// ко всем ответам, возвращаемым данным обработчиком
	Headers map[string]string
	// Глобальный обработчик, вызываемый перед всеми заданными обработчиками,
	// если определен.
	Middleware func(Handler) Handler
	// Вы можете определить свою функцию, которая будет, в зависимости от
	// ошибки, возвращать разные коды завершения запроса и текст, который будет
	// возвращаться. Например, для всех ошибок mgo.ErrNotFound устанавливать код
	// 404.
	Errors func(err error) (status int, msg string)
	// Кодировщик, используемый для декодирования запросов и кодирования ответов.
	Coder  Coder
	router // обработчики запросов по путям, без учета метода запроса
}

// Handles добавляет обработчики для указанных путей и методов. При указании
// путей можно использовать параметры, которые задаются символом ':' перед его
// названием. Так же можно использовать "завершающий" параметр, который
// "заберет"" в себя всю оставшуюся часть пути. Такой параметр задается символом
// '*' и должен обязательно идти последним в пути, в противном случае случится
// panic.
//
// В случае, если указанный обработчик не может быть приведен к типу
// поддерживаемых обработчиков, количество элементов пути больше 32767 или
// параметр со звездочкой используется не в конце пути, то случается panic.
func (m *ServeMux) Handles(paths Paths) {
	for path, methods := range paths { // перебираем все пути
		if len(methods) == 0 {
			continue // игнорируем, если методы не определены
		}
		var handlers map[string]Handler // результирующий список обработчиков
		// получаем список уже заданных обработчиков
		mh, _ := m.router.lookup(path)
		// если обработчики уже определены, то используем их
		// иначе формируем новый список обработчиков
		if h, ok := mh.(map[string]Handler); ok && h != nil {
			handlers = h
		} else {
			handlers = make(map[string]Handler, len(methods))
		}
		// добавляем в список новые обработчики по методам
		for method, h := range methods { // перебираем все методы
			if method == "" || h == nil {
				continue // игнорируем пустые обработчики
			}
			// приводим типы к обработчику
			handlers[strings.ToUpper(method)] = handler(h)
		}
		// добавляем обработчики
		if err := m.router.add(path, handlers); err != nil {
			panic(err)
		}
	}
}

// Handle добавляет новый обработчик для указанного метода и пути. Если
// обработчик не может быть приведен к поддерживаемому типу, то возникает panic.
//
// На сегодняшний момент поддерживаются следующие типы обработчиков: Handler,
// http.Handler и http.HandlerFunc.
func (m *ServeMux) Handle(method, path string, handler interface{}) {
	// сводим задачу к первоначальной
	m.Handles(Paths{path: Methods{method: handler}})
}

// ServeHTTP обеспечивает поддержку интерфейса http.Handler и обрабатывает
// основной запрос.
func (m ServeMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	context := newContext(w, req) // формируем контекст для ответа
	defer context.close()         // освобождаем по окончании
	// задаем кодировщик для разбора запросов и публикации ответов
	if m.Coder != nil {
		context.coder = m.Coder
	} else {
		context.coder = defaultCoder
	}
	// добавляем заголовки, если они определены
	if len(m.Headers) > 0 {
		for key, value := range m.Headers {
			context.HeaderSet(key, value)
		}
	}
	// если установлен базовый путь, то отрезаем его
	if m.BasePath != "" {
		p := strings.TrimPrefix(context.Request.URL.Path, m.BasePath)
		if len(p) == len(context.Request.URL.Path) {
			context.Status(http.StatusNotFound).Send(nil)
			return
		}
		context.Request.URL.Path = p
	}
	// получаем обработчик для указанного пути
	route, params := m.router.lookup(context.Request.URL.Path)
	if route == nil { // обработчик для пути не определен
		// при статусе больше 399 пустой body формирует JSON с описанием
		// ошибки автоматически
		context.Status(http.StatusNotFound).Send(nil)
		return
	}
	// добавляем параметры пути в контекст, если они есть
	if len(params) > 0 {
		context.Params = append(context.Params, params...)
	}
	methods, ok := route.(map[string]Handler) // приводим список методов
	if !ok || len(methods) == 0 {
		// если методы не определены, то лучше вернем, что путь не найден
		context.Status(http.StatusNotFound).Send(nil)
		return
	}
	// запрашиваем обработчик для метода
	handler := methods[strings.ToUpper(context.Request.Method)]
	if handler == nil { // обработчик для данного метода не определен
		// формируем список поддерживаемых методов
		allows := make([]string, 0, len(methods))
		for method := range methods {
			allows = append(allows, method)
		}
		context.HeaderSet("Allow", strings.Join(allows, ", ")).
			Status(http.StatusMethodNotAllowed).Send(nil)
		return
	}
	// если промежуточный обработчик определен, то вызываем его
	if m.Middleware != nil {
		handler = m.Middleware(handler)
	}
	// вызываем обработчик запроса
	if err := handler(context); err != nil && !context.sended {
		// преобразуем ошибку, если задан обработчик
		// игнорируем ошибки уже в нашем формате со статусом
		if _, ok := err.(Error); !ok && m.Errors != nil {
			status, msg := m.Errors(err)
			err = NewError(status, msg)
		}
		if err != nil { // если ошибка все еще есть, то отправляем ее в ответ
			context.Send(err) // отдаем ошибку
		}
	}
}
