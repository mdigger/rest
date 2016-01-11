package rest

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// JSON является вспомогательным типом для быстрого создания JSON-структур.
type JSON map[string]interface{}

// Coder описывает интерфейс кодировщика форматов для разбора и отдачи ответов.
// По умолчанию используется кодировщик в формате JSON с ограничением на размер
// запроса в 16 мегабайт. Вы можете задать свой кодировщик при инициализации
// ServeMux.
type Coder interface {
	Decode(c *Context, data interface{}) error
	Encode(c *Context, data interface{}) error
}

// defaultCoder содержит Coder по умолчанию: JSON с максимальным размером данных
// в запросе — 16 мегабайт.
var defaultCoder Coder = JSONCoder{MaxBodyBytes: 1 << 24}

// JSONCoder описывает Coder, поддерживающий кодирование и декодирование данных
// запроса в формат JSON.
//
// MaxBodyBytes ограничивает размер данных, принимаемых в запросе. Если данных
// будет больше указанного размера, то они просто обрежутся. Это позволяет
// защитить сервер от слишком больших запросов.
// По умолчанию ограничение установлено в 32 мегабайта.
// Если не хотите использовать ограничений, то установить данное значение в 0.
type JSONCoder struct {
	MaxBodyBytes int64 // максимально допустимый размер данных в запросе
}

// Encode отдает data в формате JSON. Для кодирования используется внутренний
// буфер, выбираемый из пула буферов. Если `Content-Type` не установлен, то
// устанавливает его как `application/json; charset=utf-8`. Так же устанавливает
// размер данных.
func (JSONCoder) Encode(c *Context, data interface{}) error {
	buf := buffers.Get().(*bytes.Buffer) // получаем буфер из пула
	defer buffers.Put(buf)               // возвращаем в пул по окончании
	buf.Reset()                          // сбрасываем предыдущие состояния
	// декодируем объект в формат JSON используя буфер
	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return err
	}
	if c.Header().Get("ContentType") == "" {
		c.HeaderSet("Content-Type", "application/json; charset=utf-8")
	}
	c.HeaderSet("Content-Length", strconv.Itoa(buf.Len()))
	if _, err := buf.WriteTo(c); err != nil { // отдаем сформированный ответ
		return err
	}
	return nil
}

// Decoder декодирует содержимое запроса в формате JSON в объект. Перед
// декодированием проверяется, что данные представлены в формате JSON и не
// превышают максимально допустимый размер, задаваемый в MaxBodyBytes.
func (j JSONCoder) Decode(c *Context, data interface{}) error {
	// разбираем заголовок с типом информации в запросе
	mediatype, params, _ := mime.ParseMediaType(
		c.Request.Header.Get("Content-Type"))
	charset, ok := params["charset"]
	if !ok {
		charset = "UTF-8"
	}
	// если запрос не является JSON, то возвращаем ошибку
	if mediatype != "application/json" || strings.ToUpper(charset) != "UTF-8" {
		return NewError(http.StatusUnsupportedMediaType, "")
	}
	// если запрос превышает допустимый объем, то возвращаем ошибку
	if j.MaxBodyBytes > 0 {
		if c.Request.ContentLength == 0 {
			return NewError(http.StatusLengthRequired, "")
		} else if c.Request.ContentLength > j.MaxBodyBytes {
			return NewError(http.StatusRequestEntityTooLarge, "")
		}
	}
	// разбираем содержимое запроса
	if err := json.NewDecoder(c.Request.Body).Decode(data); err != nil {
		return NewError(http.StatusBadRequest, err.Error())
	}
	return nil
}

var buffers = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
