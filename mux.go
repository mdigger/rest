package rest

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

// ServeMux описывает список обработчиков, ассоциированных с путями запроса и
// методами.
type ServeMux struct {
	// Позволяет задать базовый путь для всех запросов.
	BasePath string
	// Описывает дополнительные заголовки HTTP-ответа, которые будут добавлены
	// ко всем ответам, возвращаемым данным обработчиком
	Headers map[string]string

	routers map[string]router // обработчики запросов по методам
}

// ServeHTTP обеспечивает поддержку интерфейса http.Handler. Таким образом,
// данный ServeMux можно использовать как стандартный обработчик HTTP-запросов.
//
// Если обработчик для данного пути и метода не найден, но есть обработчики
// для других методов, то возвращается статус http.StatusMethodNotAllowed и в
// заголовке передается список методов, которые можно применить к данному пути.
// В противном случае возвращается статус http.StatusNotFound.
//
// В процессе обработки запроса отслеживаются возвращаемые ошибки и
// перехватываются возможные вызовы panic. Если ответ на запрос еще не
// отправлялся, то в этих случаях в ответ будет отправлена ошибка.
func (m ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r) // формируем контекст для ответа
	defer func() {
		// перехватываем panic, если она случилась
		if e := recover(); e != nil {
			// если еще ничего не отсылали, то отсылаем эту ошибку
			c.Send(e)
			// выводим дамп с ошибкой
			if accessLog != nil {
				c.errorLog(e, true) // записываем ошибку в лог
			}
		}
		c.close() // освобождаем по окончании
	}()
	// добавляем заголовки, если они определены
	header := c.response.Header()
	if len(m.Headers) > 0 {
		for key, value := range m.Headers {
			header.Set(key, value)
		}
	}
	path := r.URL.Path // путь запроса
	// если задан базовый путь, то удаляем его из пути обработки
	if m.BasePath != "" {
		// проверяем, что путь начинается с базового пути
		if !strings.HasPrefix(path, m.BasePath) {
			c.Send(ErrNotFound)
			return
		}
		path = strings.TrimPrefix(path, m.BasePath)
	}
	// получаем список обработчиков для данного метода
	routers := m.routers[r.Method]
	// запрашиваем подходящий обработчик
	if handler, params := routers.lookup(path); handler != nil {
		c.params = params // добавляем найденные параметры к контексту
		// если включен режим отладки, то добавляем в контекст имя функции
		// с обработчиком запроса, которое будет использоваться при выводе
		// лога
		if Debug && accessLog != nil {
			c.SetData(dataSet(137), // магическое число 137 с приватным типом
				runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name())
		}
		// вызываем обработчик запроса
		if err := handler(c); err != nil {
			if accessLog != nil {
				c.errorLog(err, false) // записываем ошибку в лог
			}
			// если ничего не отправляли, то отправляем эту ошибку
			if !c.sended {
				c.Send(err) // возвращаемые ошибки игнорируем
			}
		}
		return
	}
	// обработчик для данного пути не найден
	// собираем список методов, которые поддерживаются для данного пути
	methods := make([]string, 0, len(m.routers))
	for method, handlers := range m.routers {
		if handler, _ := handlers.lookup(path); handler != nil {
			methods = append(methods, method)
		}
	}
	if len(methods) > 0 {
		// если есть обработчики для данного пути, но с другими методами,
		// то отдаем этот список методов
		header.Set("Allow", strings.Join(methods, ", "))
		c.Status(http.StatusMethodNotAllowed).Send(nil)
	} else {
		// обработчики пути не определены ни для одного метода
		c.Send(ErrNotFound)
	}
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
func (m *ServeMux) Handle(method, path string, handler Handler) {
	if method == "" || handler == nil { // игнорируем пустые обработчики
		return
	}
	// если список обработчиков еще не инициализирован, то инициализируем его
	if m.routers == nil {
		// обычно используется не более 9 методов HTTP
		m.routers = make(map[string]router, 9)
	}
	// получаем список обработчиков для данного метода
	router, ok := m.routers[strings.ToUpper(method)]
	// добавляем обработчик для заданного метода и пути
	if err := router.add(path, handler); err != nil {
		panic(err) // обработчик нас не устраивает по каким-то причинам
	}
	if !ok {
		m.routers[strings.ToUpper(method)] = router
	}
}

// Handles добавляет сразу список обработчиков для нескольких путей и методов.
// Это, по сути, просто удобный способ сразу определить большое количество
// обработчиков, не вызывая каждый раз ServeMux.Handle.
func (m *ServeMux) Handles(paths Paths) {
	for path, methods := range paths {
		for method, handler := range methods {
			m.Handle(method, path, handler)
		}
	}
}

type (
	// Paths позволяет описать сразу несколько обработчиков для разных путей
	// и методов: ключем для данного словаря как раз являются пути запросов.
	// Используется в качестве аргумента при вызове метода ServeMux.Handles.
	Paths map[string]Methods
	// Methods позволяет описать обработчики для методов.
	Methods map[string]Handler
)

type dataSet byte // для внутреннего использования с установкой данных
