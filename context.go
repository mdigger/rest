package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// JSON является вспомогательным типом для быстрого создания JSON-структур.
type JSON map[string]interface{}

// CompressData разрешает поддержку сжатия данных, если это поддерживается браузером.
// Если сжатие данных уже поддерживается, например, на уровне вашего обработчика, то вы можете
// заблокировать двойное сжатие, установив значение false.
var CompressData = true

// Context содержит контекстную информацию HTTP-запроса и методы удобного формирования ответа
// на них.
type Context struct {
	// HTTP запрос в разобранном виде
	*http.Request
	// именованные параметры из пути запроса
	Params []Param
	// интерфейс для публикации ответа на запрос
	Response http.ResponseWriter
	// тип информации в ответе
	ContentType string
	// код HTTP-ответа
	status int
	// разобранные параметры запроса в URL (кеш)
	urlQuery url.Values
	// дополнительные данные, устанавливаемые пользователем
	// в качестве ключа рекомендуется использовать приватный тип
	// и какое-нибудь его значение, что позволит застраховаться от
	// случайной перезаписи этих данных
	data map[interface{}]interface{}
}

// newContext возвращает новый инициализированный контекст. В отличии от просто создания нового
// контекста, вызов данного метода использует пул контекстов.
func newContext(w http.ResponseWriter, r *http.Request, params []Param) *Context {
	context := contexts.Get().(*Context)
	context.Request = r
	context.Params = params
	context.urlQuery = nil
	context.data = nil
	context.Response = w
	context.ContentType = ""
	context.status = 0
	return context
}

// free возвращает контекст в пул используемых контекстов для дальнейшего использования.
// Вызывается автоматически после того, как контекст перестает использоваться.
func (c *Context) free() {
	contexts.Put(c)
}

// Get возвращает значение именованного параметра. Если параметр с таким именем не найден,
// то возвращается значение параметра из URL с тем же именем. Разбор параметров запроса сохраняется
// внутри Context и повторного его разбора уже не требует. Но это происходит только при первом
// к ним обращении.
func (c *Context) Get(key string) string {
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

// Set позволяет добавить новый параметр с заданным именем и значением. Добавление нового параметра
// с таким же именем не изменяет и не удаляет предыдущего значения, а именно добавляет его в список.
func (c *Context) Set(key, value string) {
	c.Params = append(c.Params, Param{key, value})
}

// DataGet возвращает пользовательские данные, сохраненные в контексте запроса с указанным ключем.
func (c *Context) DataGet(key interface{}) interface{} {
	if c.data == nil {
		return nil
	}
	return c.data[key]
}

// DataSet сохраняет пользовательские данные в контексте запроса с указанным ключем.
// Рекомендуется в качестве ключа использовать какой-нибудь приватный тип и его значение,
// чтобы избежать случайного затирания данных другими обработчиками. Это гарантированно обезопасит
// их от случайного доступа к ним.
func (c *Context) DataSet(key, value interface{}) {
	if c.data == nil {
		c.data = make(map[interface{}]interface{})
	}
	c.data[key] = value
}

// Code устанавливает код HTTP-ответа, который будет отправлен сервером. Данный метод возвращает
// ссылку на основной контекст, чтобы можно было использовать его в последовательности выполнения
// команд. Например, можно сразу установить код ответа и тут же опубликовать данные.
func (c *Context) Code(code int) *Context {
	if code >= 200 && code < 600 {
		c.status = code
	}
	return c
}

// SetHeader устанавливает новое значение для указанного HTTP-заголовка. Если передаваемое
// значение заголовка пустое, то данный заголовок будет удален.
func (c *Context) SetHeader(key, value string) {
	if value == "" {
		c.Response.Header().Del(key)
	} else {
		c.Response.Header().Set(key, value)
	}
}

// SetLocation устанавливает в заголовке HTTP-ответа значение для Lacation.
func (c *Context) SetLocation(value string) {
	c.SetHeader("Location", value)
}

// ParseBody декодирует содержимое запроса в объект. После чтения из запроса
// http.Request.Body автоматически закрывается и дополнительного закрытия не требуется.
//
// На данный момент поддерживается только разбор объектов в формате JSON.
func (c *Context) ParseBody(obj interface{}) error {
	defer c.Request.Body.Close()
	return json.NewDecoder(c.Request.Body).Decode(obj)
}

// Body публикует данные, переданные в параметре, в качестве ответа. Если ContentType не указан,
// то используется "application/json".
//
// В зависимости от типа передаваемых данных, ответ формируется по разному.
// Если данные являются бинарными ([]byte) или поддерживают интерфейс io.Reader, то данные отдаются
// как есть, без какого-либо изменения. Если io.Reader поддерживает io.Close, то он будет
// автоматически закрыт. Строки и ошибки преобразуются в простое JSON-сообщение, состоящие из кода
// статуса и текста сообщения. Остальные типы приводятся к формату JSON.
//
// Если клиент поддерживает сжатие при передаче данных, то автоматически включается поддержка
// сжатия ответа. Чтобы отключить данное поведение, необходимо установить флаг CompressData в false.
func (c *Context) Body(data interface{}) {
	var headers = c.Response.Header() // быстрый доступ к заголовкам ответа
	if c.ContentType == "" {
		c.ContentType = "application/json; charset=utf-8"
	}
	headers.Set("Content-Type", c.ContentType)
	// поддерживаем компрессию, если она поддерживается клиентом и не запрещена в библиотеке
	var writer io.Writer = c.Response
	if CompressData {
		switch accept := c.Request.Header.Get("Accept-Encoding"); {
		case strings.Contains(accept, "gzip"): // Поддерживается gzip-сжатие
			headers.Set("Content-Encoding", "gzip")
			headers.Add("Vary", "Accept-Encoding")
			writer = gzipGet(writer)
			defer gzipPut(writer.(io.Closer))
		case strings.Contains(accept, "deflate"): // Поддерживается deflate-сжатие
			headers.Set("Content-Encoding", "deflate")
			headers.Add("Vary", "Accept-Encoding")
			writer = deflateGet(writer)
			defer deflatePut(writer.(io.Closer))
		}
	}
	// обрабатываем статус выполнения запроса
	if c.status == 0 {
		c.status = http.StatusOK
	}
	c.Response.WriteHeader(c.status) // отдаем статус ответа
	enc := json.NewEncoder(writer)   // инициализируем JSON-encoder
	// в зависимости от типа данных поддерживаются разные методы вывода
	var err error
	switch data := data.(type) {
	case nil: // нечего отдавать
		if c.status >= 400 { // если статус соответствует ошибке, то формируем текст с ее описанием
			err = enc.Encode(JSON{"code": c.status, "error": http.StatusText(c.status)})
		}
	case io.Reader: // поток данных отдаем как есть
		_, err = io.Copy(writer, data)
		if data, ok := data.(io.Closer); ok {
			data.Close() // закрываем по окончании, раз поддерживается
		}
	case []byte: // уже готовый к отдаче набор данных
		_, err = writer.Write(data) // тоже отдаем как есть
	case error: // ошибки возвращаем в виде специального JSON
		err = enc.Encode(JSON{"code": c.status, "error": data.Error()})
	case string: // строки тоже возвращаем в виде специального JSON
		m := JSON{"code": c.status}
		if c.status >= 400 { // в случае ошибок это будет error
			m["error"] = data
		} else { // с случае просто текстовых сообщений — message
			m["message"] = data
		}
		err = enc.Encode(m)
	default: // во всех остальных случаях отдаем JSON-представление
		err = enc.Encode(data)
	}
	// если возникла ошибка, то пытаемся ее вернуть
	if err != nil {
		// почти http.Error
		headers.Set("Content-Type", "text/plain; charset=utf-8")
		headers.Set("X-Content-Type-Options", "nosniff")
		c.Response.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(writer, err)
	}
}

// contexts содержит пул контекстов
var contexts = sync.Pool{New: func() interface{} { return new(Context) }}
