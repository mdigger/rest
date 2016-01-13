package rest

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// JSON позволяет быстро описать данные в одноименном формате.
type JSON map[string]interface{}

var (
	// Взведенный флаг Debug указывает, что описания ошибок возвращаются как
	// есть. В противном случае всегда возвращается только стандартное описание
	// статуса HTTP, сформированное на базе этой ошибки.
	Debug bool = false
	// MaxBody описывает максимальный размер поддерживаемого запроса.
	// Если размер превышает указанный, то возвращается ошибка. Если не хочется
	// ограничений, то можно установить значение 0, тогда проверка производиться
	// не будет.
	MaxBody int64 = 1 << 15 // 32 мегабайта
	// Флаг Compress разрешает сжатие данных. Чтобы запретить сжимать данные,
	// установите значение данного флага в false. При инициализации сжатия
	// проверяется, что оно уже не включено, например, на уровне глобального
	// обработчика запросов и, в этом случае, сжатие не будет включено, даже
	// если флаг установлен.
	Compress bool = true
	// JSONError управляет форматом вывода ошибок: если флаг не взведен, то
	// ошибки отдаются как текст. В противном случае описание ошибок
	// возвращается в виде JSON, в котором содержится статус и описание ошибки.
	JSONError bool = true
)

// Context содержит контекстную информацию HTTP-запроса и методы формирования
// ответа на них. Т.к. http.Request импортируется в Context напрямую, то можно
// использовать все его свойства и методы, как родные свойства и методы самого
// контекста.
//
// Context скрывает http.ResponseWriter от прямого использования и, вместо
// этого, предоставляет свои методы для формирования ответа. Это позволяет
// обойти некоторые скользкие моменты и, иногда, несколько упростить код.
//
// При отдаче ответа анализируются первые байты данных и на основании них
// устанавливается тип ответа. Если вы хотите определить тип ответа
// самостоятельно, то проще всего установить значение ContentType строкой с
// описанием нужного типа.
type Context struct {
	*http.Request        // HTTP запрос в разобранном виде
	ContentType   string // тип информации в ответе

	response http.ResponseWriter         // ответ на запрос
	params   []param                     // именованные параметры из пути запроса
	status   int                         // код HTTP-ответа
	sended   bool                        // флаг отосланного ответа
	query    url.Values                  // параметры запроса в URL (кеш)
	data     map[interface{}]interface{} // дополнительные данные пользователя
	size     int                         // размер переданных данных
	started  time.Time                   // время начала обработки запроса
	writer   io.Writer                   // интерфейс для записи ответов
	compress bool                        // флаг, что мы включили сжатие
}

// newContext возвращает новый инициализированный контекст. В отличии от просто
// создания нового контекста, вызов данного метода использует пул контекстов.
func newContext(w http.ResponseWriter, r *http.Request) *Context {
	c := contexts.Get().(*Context)
	// очищаем его от возможных старых данных
	c.Request = r
	c.ContentType = ""
	c.response = w
	c.params = nil
	c.status = 0
	c.sended = false
	c.query = nil
	c.data = nil
	c.size = 0
	c.started = time.Now()
	// если сжатие еще не установлено, но поддерживается клиентом, то включаем его
	if Compress && w.Header().Get("Content-Encoding") == "" &&
		strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		gzw := gzips.Get().(*gzip.Writer)
		gzw.Reset(w)
		c.writer = gzw
		c.compress = true
	} else {
		c.writer = w
		c.compress = false
	}
	return c
}

// close возвращает контекст в пул используемых контекстов для дальнейшего
// использования. Вызывается автоматически после того, как контекст перестает
// использоваться.
func (c *Context) close() {
	// если ответ не был послан, то шлем ошибку
	if !c.sended {
		c.Status(http.StatusInternalServerError).Send(
			http.StatusText(http.StatusInternalServerError))
	}
	// если инициализировано сжатие, то закрываем и освобождаем компрессор
	if c.compress {
		if gzw, ok := c.writer.(*gzip.Writer); ok {
			gzw.Close()
			gzips.Put(gzw)
		}
	}
	c.log()         // выводим лог, если поддерживается
	contexts.Put(c) // помещаем контекст обратно в пул
}

// Header возвращает HTTP-заголовки ответа. Используется для поддержки
// интерфейса http.ResponseWriter.
func (c *Context) Header() http.Header {
	return c.response.Header()
}

// WriteHeader записывает заголовок ответа. Вызов метода автоматически взводит
// внутренний флаг, что отправка ответа начата. После его вызова отсылка
// каких-либо данных другим способом, кроме Write, уже не поддерживается.
// Используется для поддержки интерфейса http.ResponseWriter.
func (c *Context) WriteHeader(code int) {
	if c.sended {
		return
	}
	c.status = code
	if c.status == 0 {
		c.status = http.StatusOK
	} else if c.status < 100 || c.status >= 600 {
		c.status = http.StatusInternalServerError
	}
	c.sended = true
	c.response.WriteHeader(c.status)
}

// Write записывает данные в качестве ответа сервера. Может вызываться несколько
// раз. Используется для поддержки интерфейса http.ResponseWriter.
//
// При первом вызове (может быть не явный) автоматически устанавливается статус
// ответа. Если статус ответа был не задан, то будет использован статус 200
// (ОК). Так же, если не был задан ContentType, то он будет определен
// автоматически на основании анализа первых байт данных.
func (c *Context) Write(data []byte) (int, error) {
	if !c.sended {
		// выполняем только при первой отдаче данных
		header := c.response.Header()
		if header.Get("Content-Type") == "" {
			if c.ContentType == "" {
				// если тип не установлен, то анализируем его на основании
				// содержимого ответа
				c.ContentType = http.DetectContentType(data)
			}
			header.Set("Content-Type", c.ContentType)
		}
		// перед первой отдачей данных отдаем статус ответа
		c.WriteHeader(c.status)
	}
	// записываем данные в качестве ответа
	n, err := c.writer.Write(data)
	c.size += n
	return n, err
}

// Flush отдает накопленный буфер с ответом. Используется для поддержки
// интерфейса http.Flusher.
func (c *Context) Flush() {
	if flusher, ok := c.writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Status устанавливает код HTTP-ответа, который будет отправлен сервером. Вызов
// данного метода не приводит к немедленной отправке ответа, а только
// устанавливает внутренний статус. Статус должен быть в диапазоне от 200 до
// 599, в противном случае статус не изменяется.
//
// Метод возвращает ссылку на основной контекст, чтобы можно было использовать
// его в последовательности выполнения команд. Например, можно сразу установить
// код ответа и тут же опубликовать данные.
func (c *Context) Status(code int) *Context {
	if !c.sended && code >= 200 && code < 600 {
		c.status = code
	}
	return c
}

// Param возвращает значение именованного параметра. Если параметр с таким
// именем не найден, то возвращается значение параметра из URL с тем же именем.
//
// Разобранные параметры запроса пути сохраняются внутри Context и повторного
// его разбора уже не требует. Но это происходит только при первом к ним
// обращении.
func (c *Context) Param(key string) string {
	for _, param := range c.params {
		if param.Key == key {
			return param.Value
		}
	}
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query.Get(key)
}

// Data возвращает пользовательские данные, сохраненные в контексте запроса с
// указанным ключем. Обычно эти данные используются, когда необходимо передать
// их между несколькими обработчиками.
func (c *Context) Data(key interface{}) interface{} {
	if c.data == nil {
		return nil
	}
	return c.data[key]
}

// SetData сохраняет пользовательские данные в контексте запроса с указанным
// ключем.
//
// Рекомендуется в качестве ключа использовать какой-нибудь приватный тип и его
// значение, чтобы избежать случайного затирания данных другими обработчиками:
// это гарантированно обезопасит от случайного доступа к ним. Но строки тоже
// поддерживаются. :)
func (c *Context) SetData(key, value interface{}) {
	if c.data == nil {
		c.data = make(map[interface{}]interface{})
	}
	c.data[key] = value
}

// Bind разбирает данные запроса в формате JSON и заполняет ими указанный в
// параметре объект.
//
// Если Content-Type запроса не соответствует "application/json", то
// возвращается ошибка ErrUnsupportedMediaType. Так же может возвращать ошибку
// ErrLengthRequired, если не указана длина запроса, ErrRequestEntityTooLarge —
// если запрос превышает значение MaxBody, и ErrBadRequest — если не смогли
// разобрать запрос и поместить результат разбора в объект obj. Все эти ошибки
// поддерживаются методом Send и отдают соответствующий статус ответа на запрос.
func (c *Context) Bind(obj interface{}) error {
	// разбираем заголовок с типом информации в запросе
	mediatype, params, _ := mime.ParseMediaType(
		c.Request.Header.Get("Content-Type"))
	charset, ok := params["charset"]
	if !ok {
		charset = "UTF-8"
	}
	// если запрос не является JSON, то возвращаем ошибку
	if mediatype != "application/json" || strings.ToUpper(charset) != "UTF-8" {
		return ErrUnsupportedMediaType
	}
	// если запрос превышает допустимый объем, то возвращаем ошибку
	if MaxBody > 0 {
		if c.Request.ContentLength == 0 {
			return ErrLengthRequired
		} else if c.Request.ContentLength > MaxBody {
			return ErrRequestEntityTooLarge
		}
	}
	// разбираем данные из запроса
	if err := json.NewDecoder(c.Request.Body).Decode(obj); err != nil {
		return ErrBadRequest
	}
	return nil
}

// Эти ошибки обрабатываются при передаче их в метод Context.Send и
// устанавливают соответствующий статус ответа.
//
// Кроме указанных здесь ошибок, так же проверяется, что ошибка отвечает на
// os.IsNotExist (в этом случае статус станет 404) или os.IsPermission (статус
// 403). Все остальные ошибки устанавливают статус 500.
var (
	ErrDataAlreadySent       = errors.New("data already sent")
	ErrBadRequest            = errors.New("bad request")              // 400
	ErrUnauthorized          = errors.New("unauthorized")             // 401
	ErrForbidden             = errors.New("forbidden")                // 403
	ErrNotFound              = errors.New("not found")                // 404
	ErrLengthRequired        = errors.New("length required")          // 411
	ErrRequestEntityTooLarge = errors.New("request entity too large") // 413
	ErrUnsupportedMediaType  = errors.New("unsupported media type")   // 415
	ErrInternalServerError   = errors.New("internal server error")    // 500
	ErrNotImplemented        = errors.New("not implemented")          // 501
	ErrServiceUnavailable    = errors.New("service unavailable")      // 503
)

// Send отсылает переданные данные как ответ на запрос. В зависимости от типа
// данных, используются разные форматы ответов. Поддерживаются данные в формате
// string, error, []byte, io.Reader и nil. Все остальные типы данных приводятся
// к формату JSON.
//
// Данный метод можно использовать только один раз: после того, как ответ
// отправлен, повторный вызов данного метода сразу возвращает ошибку.
func (c *Context) Send(data interface{}) (err error) {
	// не можем отправить ответ, если он уже отправлен
	if c.sended {
		return ErrDataAlreadySent
	}
	// в зависимости от типа данных, отдаем их разными способами
	switch data := data.(type) {
	case nil:
		if c.status == 0 {
			c.status = http.StatusNoContent
		}
		_, err = c.Write(nil)
	case string:
		if c.ContentType == "" {
			c.ContentType = "text/plain; charset=utf-8"
		}
		_, err = fmt.Fprint(c, data)
	case error:
		// если статус не установлен, то ориентируемся на тип ошибки
		if c.status == 0 {
			switch data {
			case ErrBadRequest: // 400
				c.status = http.StatusBadRequest
			case ErrUnauthorized: // 401
				c.status = http.StatusUnauthorized
			case ErrForbidden: // 403
				c.status = http.StatusForbidden
			case ErrNotFound: // 404
				c.status = http.StatusNotFound
			case ErrLengthRequired: // 411
				c.status = http.StatusLengthRequired
			case ErrRequestEntityTooLarge: // 413
				c.status = http.StatusRequestEntityTooLarge
			case ErrUnsupportedMediaType: // 415
				c.status = http.StatusUnsupportedMediaType
			case ErrInternalServerError: // 500
				c.status = http.StatusInternalServerError
			case ErrNotImplemented: // 501
				c.status = http.StatusNotImplemented
			case ErrServiceUnavailable: // 503
				c.status = http.StatusServiceUnavailable
			default:
				if os.IsNotExist(data) {
					c.status = http.StatusNotFound
				} else if os.IsPermission(data) {
					c.status = http.StatusForbidden
				} else {
					c.status = http.StatusInternalServerError
				}
			}
		}
		// В отладочном режиме возвращаем текст ошибки, иначе — тест статуса
		var msg string
		if Debug {
			msg = data.Error()
		} else {
			msg = http.StatusText(c.status)
		}
		// В зависимости от флага, ошибку выводим как JSON или как текст
		if JSONError {
			c.ContentType = "application/json; charset=utf-8"
			err = json.NewEncoder(c).Encode(JSON{"code": c.status, "error": msg})
		} else {
			c.ContentType = "text/plain; charset=utf-8"
			_, err = fmt.Fprint(c, msg)
		}
	case []byte:
		_, err = c.Write(data)
	case io.Reader:
		_, err = io.Copy(c, data)
	default: // кодируем как JSON
		if c.ContentType == "" {
			c.ContentType = "application/json; charset=utf-8"
		}
		err = json.NewEncoder(c).Encode(data)
	}
	// если в процессе отправки произошла ошибка, но мы еще ничего не
	// отправили, то отдаем ошибку
	if err != nil && !c.sended {
		c.status = http.StatusInternalServerError
		var msg string
		if Debug {
			msg = err.Error()
		} else {
			msg = http.StatusText(c.status)
		}
		// В зависимости от флага, ошибку выводим как JSON или как текст
		if JSONError {
			c.ContentType = "application/json; charset=utf-8"
			json.NewEncoder(c).Encode(JSON{"code": c.status, "error": msg})
		} else {
			c.ContentType = "text/plain; charset=utf-8"
			fmt.Fprint(c, msg)
		}
	}
	return
}

// Redirect отсылает ответ с требованием временного перехода по указанному URL.
// Ошибка никогда не возвращается.
func (c *Context) Redirect(url string) error {
	http.Redirect(c, c.Request, url, http.StatusFound)
	return nil
}

// пулы
var (
	contexts = sync.Pool{New: func() interface{} { return new(Context) }}
	gzips    = sync.Pool{New: func() interface{} { return new(gzip.Writer) }}
	buffers  = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
)
