package rest

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
)

// Context содержит контекстную информацию HTTP-запроса и методы формирования ответа на них.
// Т.к. http.Request импортируется в Context напрямую, то можно использовать все его свойства и
// методы, как родные свойства и методы самого контекста.
//
// Context скрывает http.ResponseWriter от прямого использования и, вместо этого, предоставляет
// свои методы для формирования ответа. Это позволяет обойти некоторые скользкие моменты и, иногда,
// несколько упростить код.
//
// При отдаче ответа анализируются первые байты данных и на основании них устанавливается тип
// ответа. Если вы хотите определить тип ответа самостоятельно, то проще всего установить значение
// ContentType строкой с описанием нужного типа.
//
// Для кодирования строк, ошибок и объектов используется кодировщик по умолчанию (JSON), если не
// задано другого при инициализации ServeMux.
type Context struct {
	*http.Request         // HTTP запрос в разобранном виде
	Params        []Param // именованные параметры из пути запроса
	ContentType   string  // тип информации в ответе

	response http.ResponseWriter         // интерфейс для публикации ответа на запрос
	status   int                         // код HTTP-ответа
	sended   bool                        // флаг, что ответ уже отослан (заголовок ответа отдан)
	urlQuery url.Values                  // разобранные параметры запроса в URL (кеш)
	data     map[interface{}]interface{} // дополнительные данные, устанавливаемые пользователем.
	coder    Coder                       // кодировщик ответов и разборщик запросов
}

// newContext возвращает новый инициализированный контекст. В отличии от просто создания нового
// контекста, вызов данного метода использует пул контекстов.
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	context := contexts.Get().(*Context) // получаем контекст из пула контекстов
	// очищаем его от возможных старых данных
	context.Request = r
	context.Params = nil
	context.ContentType = ""
	context.response = w
	context.status = 0
	context.sended = false
	context.urlQuery = nil
	context.data = nil
	context.coder = defaultCoder
	return context
}

// close возвращает контекст в пул используемых контекстов для дальнейшего использования.
// Вызывается автоматически после того, как контекст перестает использоваться.
func (c *Context) close() {
	contexts.Put(c) // возвращаем контекст в пул контекстов для повторного использования
}

// Header возвращает HTTP-заголовки ответа.
func (c *Context) Header() http.Header {
	return c.response.Header()
}

// Write возвращает данные из параметра в качестве ответа сервера. Автоматически устанавливает
// статус ответа в http.StatusOK, если не было указано другого статуса, а так же взводит внутренний
// флаг, что отсылка ответа начата. Если не был установлен заголовок `Content-Type`, то определяет
// тип информации по первым байтам данных и автоматически устанавливает его при записи первой
// порции информации. При этом свойство Context.ContentType не используется.
func (c *Context) Write(data []byte) (int, error) {
	if !c.sended {
		if c.Header().Get("Content-Type") == "" {
			c.SetHeader("Content-Type", http.DetectContentType(data))
		}
		c.WriteHeader(c.status)
	}
	return c.response.Write(data)
}

// WriteHeader записывает заголовок ответа. Вызов метода автоматически взводит внутренний флаг,
// что отправка ответа начата. После его вызова отсылка каких-либо данных другим способом, кроме
// Write уже не поддерживается.
func (c *Context) WriteHeader(code int) {
	c.status = code
	if c.status == 0 {
		c.status = http.StatusOK
	} else if c.status < 100 || c.status >= 600 {
		c.status = http.StatusInternalServerError
	}
	c.sended = true
	c.response.WriteHeader(c.status)
}

// Flush отдает накопленный буфер с ответом, если поддерживается. Метод срабатывает только, если
// хоть какая-то часть данных уже передана.
func (c *Context) Flush() {
	if flusher, ok := c.response.(http.Flusher); ok && c.sended {
		flusher.Flush()
	}
}

// Status устанавливает код HTTP-ответа, который будет отправлен сервером. Вызов данного метода
// не приводит к немедленной отправке ответа, а только устанавливает внутренний статус.
//
// Метод возвращает ссылку на основной контекст, чтобы можно было использовать его в
// последовательности выполнения команд. Например, можно сразу установить код ответа и тут же
// опубликовать данные.
func (c *Context) Status(code int) *Context {
	if !c.sended && code >= 200 && code < 600 {
		c.status = code
	}
	return c
}

// SetHeader устанавливает новое значение для указанного HTTP-заголовка ответа. Все записи с таким
// же именем заголовка будут перезаписаны. Если передаваемое значение заголовка пустое, то данный
// заголовок будет удален.
//
// Если устанавливается заголовок `Content-Type`, то соответствующее свойство контекста тоже
// принимает это же значение. Заголовок `Content-Length` будет установлен только в том случае,
// если ответ не сжимается (проверка проводится по заголовку ответа).
func (c *Context) SetHeader(key, value string) *Context {
	if !c.sended { // нельзя изменить заголовок после ответа
		switch key {
		case "Content-Type":
			c.ContentType = value
		case "Content-Length":
			if c.Header().Get("Content-Encoding") != "" {
				value = "" // не устанавливаем длину, если поддерживается сжатие ответа
			}
		}
		if value == "" {
			c.response.Header().Del(key) // удаляем ключ, если пустое значение
		} else {
			c.response.Header().Set(key, value) // перезаписываем или создаем ключ заголовка
		}
	}
	return c // возвращаем контекст, чтобы поддержать конвейер
}

// Parse декодирует содержимое запроса в объект. Максимальный размер содержимого запроса ограничен
// размером MaxBytes, если установлен. Возвращает ошибку HTTPError, если данные не соответствуют
// формату JSON или не получается их разобрать.
func (c *Context) Parse(data interface{}) *HTTPError {
	return c.coder.Decode(c, data) // декодируем запрос и возвращаем ошибку, если случилась
}

// Param возвращает значение именованного параметра. Если параметр с таким именем не найден,
// то возвращается значение параметра из URL с тем же именем.
//
// Разобранные параметры запроса пути сохраняются внутри Context и повторного его разбора уже не
// требует. Но это происходит только при первом к ним обращении.
func (c *Context) Param(key string) string {
	for _, param := range c.Params {
		if param.Key == key {
			return param.Value
		}
	}
	if c.urlQuery == nil {
		c.urlQuery = c.Request.URL.Query()
	}
	return c.urlQuery.Get(key)
}

// Data возвращает пользовательские данные, сохраненные в контексте запроса с указанным ключем.
// Обычно такие данные сохраняются в контексте запроса, если их нужно передать между несколькими
// обработчиками.
func (c *Context) Data(key interface{}) interface{} {
	if c.data == nil {
		return nil
	}
	return c.data[key]
}

// SetData сохраняет пользовательские данные в контексте запроса с указанным ключем.
//
// Рекомендуется в качестве ключа использовать какой-нибудь приватный тип и его значение,
// чтобы избежать случайного затирания данных другими обработчиками: это гарантированно обезопасит
// от случайного доступа к ним. Но строки тоже поддерживаются. :)
func (c *Context) SetData(key, value interface{}) {
	if c.data == nil {
		c.data = make(map[interface{}]interface{})
	}
	c.data[key] = value
}

// Error возвращает ошибку с указанным кодом окончания запроса. Является просто удобным способом
// сформировать ошибку HTTPError. Эту ошибку можно вернуть в качестве ошибки выполнения обработки.
func (c *Context) Error(code int) *HTTPError {
	return NewHTTPError(code)
}

// ErrDoubleSend возвращается Context.Send в случае повторной попытки послать данные, когда
// ответ уже был отправлен.
var ErrDoubleSend = errors.New("double send")

// Send публикует переданные в параметре данные в качестве ответа. Если context.ContentType
// не указан, то используется тип данных будет определен по первым отдаваемым байтам.
//
// В зависимости от типа передаваемых данных, ответ формируется по разному.
// Если данные являются бинарными ([]byte) или поддерживают интерфейс io.Reader, то отдаются
// как есть, без какого-либо изменения. Если io.Reader поддерживает io.ReadCloser, то он будет
// автоматически закрыт. Строки и ошибки преобразуются в простое JSON-сообщение, состоящие из кода
// статуса и текста сообщения. Остальные типы приводятся к формату JSON.
//
// Вызов данного метода сразу инициализирует отдачу содержимого в качестве ответа. Поэтому нет
// смысла вызывать его несколько раз, т.к. нельзя второй раз записать разные коды ответа. В случае
// повторного вызова этого метода, когда данные уже были отданы, будет возвращена ошибка
// ErrDoubleSend.
//
// Если клиент поддерживает сжатие данных, то автоматически включается поддержка сжатия ответа.
// Чтобы отключить данное поведение, установите флаг Compress в false.
func (c *Context) Send(data interface{}) error {
	if c.sended {
		return ErrDoubleSend // уже отправлено сообщение — ничего больше изменить не получится
	}
	if c.ContentType != "" {
		c.SetHeader("Content-Type", c.ContentType)
	}
	// в зависимости от типа данных поддерживаются разные методы вывода
	// для []byte и io.Reader отдаем все как есть, а для остальных типов данных формируем ответ
	// в формате JSON
	switch d := data.(type) {
	case nil: // нечего отдавать
		if c.status == 0 {
			c.status = http.StatusNoContent
		} else if c.status >= 400 {
			// если статус соответствует ошибке, то формируем текст с ее описанием
			return c.encode(&HTTPError{c.status, http.StatusText(c.status)})
		}
		return nil
	case error:
		if he, ok := d.(*HTTPError); ok {
			if he.Code >= 200 && he.Code < 600 {
				c.status = he.Code // если в ошибке есть статус, то устанавливаем именно его
			} else {
				c.status = http.StatusInternalServerError
			}
		} else if c.status == 0 { // если статус не установлен, то ориентируемся на тип ошибки
			switch {
			case os.IsNotExist(d):
				c.status = http.StatusNotFound
			case os.IsPermission(d):
				c.status = http.StatusForbidden
			default:
				c.status = http.StatusInternalServerError
			}
		}
		return c.encode(&HTTPError{c.status, d.Error()})
	case string: // строки тоже возвращаем в виде специального JSON
		return c.encode(&HTTPError{c.status, d})
	case []byte: // уже готовый к отдаче набор данных
		c.SetHeader("Content-Length", strconv.Itoa(len(d)))
		_, err := c.Write(d) // тоже отдаем как есть
		return err
	case io.Reader: // поток данных отдаем как есть
		// вычисляем размер данных и записываем их в заголовок
		if seeker, ok := d.(io.Seeker); ok {
			// переходим к концу потока и смотрим размер
			size, err := seeker.Seek(0, os.SEEK_END)
			if err != nil {
				return err
			}
			// возвращаемся к началу потока
			if _, err = seeker.Seek(0, os.SEEK_SET); err != nil {
				return err
			}
			// устанавливаем размер ответа
			c.SetHeader("Content-Length", strconv.FormatInt(size, 10))
		}
		_, err := io.Copy(c, d) // копируем данные в ответ
		if closer, ok := d.(io.Closer); ok {
			closer.Close() // закрываем по окончании, раз поддерживается
		}
		return err
	default: // во всех остальных случаях отдаем JSON-представление
		return c.encode(data)
	}
}

// encode декодируем ответ в формат JSON и отдает его. Для кодирования используется пул с
// внутренними буферами.
func (c *Context) encode(data interface{}) error {
	if err := c.coder.Encode(c, data); err != nil {
		return err
	}
	return nil
}

// пулы контекстов
var contexts = sync.Pool{New: func() interface{} { return new(Context) }}
