package rest

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Error returned by the Bind function.
var (
	ErrUnsupportedCharset     = &Error{400, "unsupported charset"}
	ErrEmptyContentType       = &Error{400, "empty content type"}
	ErrUnsupportedContentType = &Error{400, "unsupported content type"}
	ErrUnsupportedHTTPMethod  = &Error{400, "unsupported http method"}
)

// bind parses the request and populates the received data specified structure.
// Supported parsing of JSON, XML and HTTP form. For HTTP form in the structure,
// you can use the tag "form:" to specify the name.
func bind(r *http.Request, v interface{}) (err error) {
	switch r.Method {
	case "GET", "HEAD":
		err = bindForm(r.URL.Query(), v)
	case "POST", "PUT", "PATCH":
		mediatype, params, _ := mime.ParseMediaType(
			r.Header.Get("Content-Type"))
		charset, ok := params["charset"]
		if ok && strings.ToUpper(charset) != "UTF-8" {
			err = ErrUnsupportedCharset
			break
		}
		switch mediatype {
		case "application/json":
			err = json.NewDecoder(r.Body).Decode(v)
		case "application/xml":
			err = xml.NewDecoder(r.Body).Decode(v)
		case "application/x-www-form-urlencoded", "multipart/form-data":
			if err = r.ParseForm(); err == nil {
				err = bindForm(r.PostForm, v)
			}
		case "":
			err = ErrEmptyContentType
		default:
			err = ErrUnsupportedContentType
		}
	default:
		err = ErrUnsupportedHTTPMethod
	}
	return err
}

func bindForm(data url.Values, v interface{}) error {
	typ := reflect.TypeOf(v).Elem()
	val := reflect.ValueOf(v).Elem()
	if typ.Kind() != reflect.Struct {
		return errors.New("binding element must be a struct")
	}
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}
		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get("form")
		if inputFieldName == "" {
			r, n := utf8.DecodeRuneInString(typeField.Name)
			inputFieldName = string(unicode.ToLower(r)) + typeField.Name[n:]
			// If "form" tag is nil, we inspect if the field is a struct.
			if structFieldKind == reflect.Struct {
				err := bindForm(data, structField.Addr().Interface())
				if err != nil {
					return err
				}
				continue
			}
		}
		inputValue, exists := data[inputFieldName]
		if !exists {
			continue
		}
		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for j := 0; j < numElems; j++ {
				if err := setWithProperType(sliceOf, inputValue[j],
					slice.Index(j)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else {
			if err := setWithProperType(typeField.Type.Kind(), inputValue[0],
				structField); err != nil {
				return err
			}
		}
	}
	return nil
}

func setWithProperType(valueKind reflect.Kind,
	val string, structField reflect.Value) error {
	// log.WithFields(log.Fields{
	// 	"valueKind":   valueKind,
	// 	"val":         val,
	// 	"structField": structField,
	// }).Info("setWithProperType")

	switch valueKind {
	case reflect.Ptr:
		if structField.IsNil() {
			structField.Set(reflect.New(structField.Type().Elem()))
		}
		return setWithProperType(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
		return nil
	default:
		// log.WithFields(log.Fields{
		// 	"valueKind":   valueKind,
		// 	"val":         val,
		// 	"structField": structField,
		// }).Warning("unsupported field type")
		return errors.New("unsupported field type")
	}
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "true"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
