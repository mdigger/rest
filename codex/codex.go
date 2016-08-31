// Package codex соответствует интерфейсу rest.Coder и добавляет одновременную
// поддержку форматов MsgPack, Cbor, Binc и JSON.
package codex

import (
	"mime"
	"strings"

	"github.com/mdigger/rest"
	"github.com/mdigger/rest/httpaccept"
	"github.com/ugorji/go/codec"
)

var (
	hmsgpack = new(codec.MsgpackHandle)
	hcbor    = new(codec.CborHandle)
	hbinc    = new(codec.BincHandle)
	hjson    = new(codec.JsonHandle)
)

func init() {
	hjson.Canonical = true        // сортировать ключи в словаре
	hjson.Indent = -1             // отступ с табуляцией
	rest.Encoder = Coder{1 << 15} // регистрируем при экспорте
}

// Coder поддерживает декодирование запроса и отсылку ответа в форматах JSON,
// CBOR, MsgPack и Binc.
type Coder struct {
	MaxBody int64 // максимально допустимый размер запроса
}

// Bind разбирает содержимое запроса в формате MsgPack, Cbor, Binc или JSON
// и сериализует его содержимое в объект.
func (cdx Coder) Bind(c *rest.Context, obj interface{}) error {
	r := c.Request // запрос
	// если запрос превышает допустимый объем, то возвращаем ошибку
	if cdx.MaxBody > 0 {
		if r.ContentLength == 0 {
			return rest.ErrLengthRequired
		} else if r.ContentLength > cdx.MaxBody {
			return rest.ErrRequestEntityTooLarge
		}
	}
	// разбираем заголовок с типом информации в запросе
	mediatype, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	charset, ok := params["charset"]
	if !ok {
		charset = "UTF-8"
	}
	var h codec.Handle
	switch mediatype {
	case "application/json":
		if strings.ToUpper(charset) != "UTF-8" {
			return rest.ErrUnsupportedMediaType
		}
		h = hjson
	case "application/msgpack", "application/x-msgpack":
		h = hmsgpack
	case "application/cbor":
		h = hcbor
	case "application/binc", "application/x-binc":
		h = hbinc
	default:
		return rest.ErrUnsupportedMediaType
	}
	// разбираем данные из запроса
	if err := codec.NewDecoder(r.Body, h).Decode(obj); err != nil {
		return rest.ErrBadRequest
	}
	return nil
}

// Encode кодирует и отправляет ответ с содержимым obj в формате MsgPack,
// Cbor, Binc или JSON, в зависимости от предпочтений, определяемых на основании
// заголовка запроса.
func (Coder) Encode(c *rest.Context, obj interface{}) error {
	mediatype := httpaccept.Negotiate(c.Request.Header.Get("Accept"), []string{
		"application/cbor",
		"application/msgpack",
		"application/x-msgpack",
		"application/binc",
		"application/x-binc",
		"application/json",
	})
	var h codec.Handle
	switch mediatype {
	case "application/msgpack", "application/x-msgpack":
		h = hmsgpack
	case "application/cbor":
		h = hcbor
	case "application/binc", "application/x-binc":
		h = hbinc
	default:
		h = hjson
		mediatype = "application/json; charset=utf-8"
	}
	c.ContentType = mediatype
	return codec.NewEncoder(c, h).Encode(obj)
}
