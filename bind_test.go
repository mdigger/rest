package rest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type testStruct struct {
	Test string `json:"test" xml:"test,attr" form:"struct.test"`
}
type BindData struct {
	String  string
	Int     int      `json:"int" xml:"int,attr" form:"int"`
	Int8    int8     `json:"int8" xml:"int8,attr" form:"int8"`
	Int16   int16    `json:"int16" xml:"int16,attr" form:"int16"`
	Int32   int32    `json:"int32" xml:"int32,attr" form:"int32"`
	Int64   int64    `json:"int64" xml:"int64,attr" form:"int64"`
	UInt    uint     `json:"uint" xml:"uint,attr" form:"uint"`
	UInt8   uint8    `json:"uint8" xml:"uint8,attr" form:"uint8"`
	UInt16  uint16   `json:"uint16" xml:"uint16,attr" form:"uint16"`
	UInt32  uint32   `json:"uint32" xml:"uint32,attr" form:"uint32"`
	UInt64  uint64   `json:"uint64" xml:"uint64,attr" form:"uint64"`
	Float32 float32  `json:"float32" xml:"float32,attr" form:"float32"`
	Float64 float64  `json:"float64" xml:"float64,attr" form:"float64"`
	Bool    bool     `json:"bool" xml:"bool,attr" form:"bool"`
	Array   []string `json:"array" xml:"array" form:"array"`
	_       string
	Struct  testStruct
}

func (bd *BindData) Check(source *BindData) error {
	if bd == nil {
		return errors.New("nil data")
	}
	if bd.String != source.String {
		return fmt.Errorf("bad string value: %v", bd.String)
	}
	if bd.Int != source.Int {
		return fmt.Errorf("bad int value: %v", bd.Int)
	}
	if bd.Int8 != source.Int8 {
		return fmt.Errorf("bad int8 value: %v", bd.Int8)
	}
	if bd.Int16 != source.Int16 {
		return fmt.Errorf("bad int16 value: %v", bd.Int16)
	}
	if bd.Int32 != source.Int32 {
		return fmt.Errorf("bad int32 value: %v", bd.Int32)
	}
	if bd.Int64 != source.Int64 {
		return fmt.Errorf("bad int64 value: %v", bd.Int64)
	}
	if bd.UInt != source.UInt {
		return fmt.Errorf("bad uint value: %v", bd.UInt)
	}
	if bd.UInt8 != source.UInt8 {
		return fmt.Errorf("bad uint8 value: %v", bd.UInt8)
	}
	if bd.UInt16 != source.UInt16 {
		return fmt.Errorf("bad uint16 value: %v", bd.UInt16)
	}
	if bd.UInt32 != source.UInt32 {
		return fmt.Errorf("bad uint32 value: %v", bd.UInt32)
	}
	if bd.UInt64 != source.UInt64 {
		return fmt.Errorf("bad uint64 value: %v", bd.UInt64)
	}
	if bd.Float32 != source.Float32 {
		return fmt.Errorf("bad float32 value: %v", bd.Float32)
	}
	if bd.Float64 != source.Float64 {
		return fmt.Errorf("bad float64 value: %v", bd.Float64)
	}
	if bd.Bool != source.Bool {
		return fmt.Errorf("bad bool value: %v", bd.Bool)
	}
	if len(bd.Array) != len(source.Array) {
		return fmt.Errorf("bad array value: %v", bd.Array)
	}
	if bd.Struct.Test != source.Struct.Test {
		return fmt.Errorf("bad struct value: %v", bd.Struct.Test)
	}
	return nil
}

func (bd *BindData) Query() string {
	values := url.Values{}
	values.Set("string", bd.String)
	if bd.Int != 0 {
		values.Set("int", strconv.Itoa(bd.Int))
	} else {
		values.Set("int", "")
	}
	if bd.Int8 != 0 {
		values.Set("int8", strconv.FormatInt(int64(bd.Int8), 10))
	} else {
		values.Set("int8", "")
	}
	if bd.Int16 != 0 {
		values.Set("int16", strconv.FormatInt(int64(bd.Int16), 10))
	} else {
		values.Set("int16", "")
	}
	if bd.Int32 != 0 {
		values.Set("int32", strconv.FormatInt(int64(bd.Int32), 10))
	} else {
		values.Set("int32", "")
	}
	if bd.Int64 != 0 {
		values.Set("int64", strconv.FormatInt(int64(bd.Int64), 10))
	} else {
		values.Set("int64", "")
	}
	if bd.UInt != 0 {
		values.Set("uint", strconv.FormatUint(uint64(bd.UInt), 10))
	} else {
		values.Set("uint", "")
	}
	if bd.UInt8 != 0 {
		values.Set("uint8", strconv.FormatUint(uint64(bd.UInt8), 10))
	} else {
		values.Set("uint8", "")
	}
	if bd.UInt16 != 0 {
		values.Set("uint16", strconv.FormatUint(uint64(bd.UInt16), 10))
	} else {
		values.Set("uint16", "")
	}
	if bd.UInt32 != 0 {
		values.Set("uint32", strconv.FormatUint(uint64(bd.UInt32), 10))
	} else {
		values.Set("uint32", "")
	}
	if bd.UInt64 != 0 {
		values.Set("uint64", strconv.FormatUint(uint64(bd.UInt64), 10))
	} else {
		values.Set("uint64", "")
	}
	if bd.Float32 != 0 {
		values.Set("float32", strconv.FormatFloat(float64(bd.Float32), 'f', 4, 32))
	} else {
		values.Set("float32", "")
	}
	if bd.Float64 != 0 {
		values.Set("float64", strconv.FormatFloat(float64(bd.Float64), 'f', 4, 64))
	} else {
		values.Set("float64", "")
	}
	if bd.Bool {
		values.Set("bool", strconv.FormatBool(bd.Bool))
	} else {
		values.Set("bool", "")
	}
	if len(bd.Array) > 0 {
		for _, str := range bd.Array {
			values.Add("array", str)
		}
	}
	values.Add("struct.test", bd.Struct.Test)
	return values.Encode()
}

func TestBind(t *testing.T) {
	testData := &BindData{
		String:  "string",
		Int:     10,
		Int8:    8,
		Int16:   16,
		Int32:   32,
		Int64:   64,
		UInt:    10,
		UInt8:   8,
		UInt16:  16,
		UInt32:  32,
		UInt64:  64,
		Float32: 32.32,
		Float64: 64.64,
		Bool:    true,
		Array:   []string{"s1", "s2", "s3", "s4", "s5"},
		Struct:  testStruct{"test"},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json")
	var v = new(BindData)
	if err = bind(r, v); err != nil {
		t.Error("bind error:", err)
	}
	if err = testData.Check(v); err != nil {
		t.Error("check error:", err)
	}

	data, err = xml.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}
	r = httptest.NewRequest("PUT", "/", bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/xml")
	v = new(BindData)
	if err = bind(r, v); err != nil {
		t.Error("bind error:", err)
	}
	if err = testData.Check(v); err != nil {
		t.Error("check error:", err)
	}

	r = httptest.NewRequest("PATCH", "/", bytes.NewReader([]byte(testData.Query())))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	v = new(BindData)
	if err = bind(r, v); err != nil {
		t.Error("bind error:", err)
	}
	if err = testData.Check(v); err != nil {
		t.Error("check error:", err)
	}

	r = httptest.NewRequest("GET", "/?"+testData.Query(), nil)
	v = new(BindData)
	if err = bind(r, v); err != nil {
		t.Error("bind error:", err)
	}
	if err = testData.Check(v); err != nil {
		t.Error("check error:", err)
	}

	// pretty.Println(v)

	r = httptest.NewRequest("TEST", "/", bytes.NewReader(data))
	if err = bind(r, v); err != ErrUnsupportedHTTPMethod {
		t.Error("bind error: unsupported method")
	}

	r = httptest.NewRequest("POST", "/", bytes.NewReader(data))
	r.Header.Set("Content-Type", "text/json")
	if err = bind(r, v); err != ErrUnsupportedContentType {
		t.Error("bind error: unsupported media type")
	}

	r = httptest.NewRequest("POST", "/", bytes.NewReader(data))
	r.Header.Set("Content-Type", "application/json; charset=windows-1251")
	if err = bind(r, v); err != ErrUnsupportedCharset {
		t.Error("bind error: unsupported charset")
	}

	r = httptest.NewRequest("POST", "/", bytes.NewReader(data))
	if err = bind(r, v); err != ErrEmptyContentType {
		t.Error("bind error: empty media type")
	}

	testData = new(BindData)
	r = httptest.NewRequest("GET", "/?"+testData.Query(), nil)
	v = new(BindData)
	if err = bind(r, v); err != nil {
		t.Error("bind error:", err)
	}
	if err = testData.Check(v); err != nil {
		t.Error("check error:", err)
	}
}

func TestBindForm(t *testing.T) {
	datamap := make(map[string]string)
	err := bindForm(nil, &datamap)
	if err.Error() != "binding element must be a struct" {
		t.Error("bad type")
	}

	err = setWithProperType(reflect.Complex64, "val", reflect.ValueOf(nil))
	if err.Error() != "unsupported field type" {
		t.Error("unsupported field type")
	}
}

func TestBindDebug(t *testing.T) {
	var v = new(struct{})
	data := `{adfasdf}`
	r := httptest.NewRequest("POST", "/", strings.NewReader(data))
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	bind(r, v)
}

func TestFormWithPtr(t *testing.T) {
	var data = url.Values{
		"readed": {"true"},
		"note":   {"text note"},
		"b":      {"123"},
	}
	var obj = new(struct {
		Readed *bool
		Note   *string
		B      *int
	})
	req := httptest.NewRequest("PATCH", "/", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := bind(req, obj); err != nil {
		t.Error(err)
	}
	// pretty.Println(obj)
}
