package rest_test

import (
	"fmt"
	"net/http"
	"os"

	"github.com/geotrace/rest"
)

var c = new(rest.Context) // test context

func Example() {
	var mux rest.ServeMux
	mux.Handles(rest.Paths{
		"/user/:id": {
			"GET": rest.Func(func(c *rest.Context) {
				c.Body(rest.JSON{"user": c.Get("id")})
			}),
			"POST": rest.Func(func(c *rest.Context) {
				var data = make(rest.JSON)
				if err := c.Parse(&data); err != nil {
					c.Code(500).Body(err)
					return
				}
				c.Body(rest.JSON{"user": c.Get("id"), "data": data})
			}),
		},
		"/message/:text": {
			"GET": rest.Func(func(c *rest.Context) {
				c.Body(rest.JSON{"message": c.Get("text")})
			}),
		},
	})
	// т.к. поддерживается интерфейс http.Handler, то можно использовать
	// с любыми стандартными библиотеками
	http.ListenAndServe(":8080", mux)
}

func ExampleContext_DataSet() {
	type myKeyType byte     // определяем собственный тип данных
	var myKey myKeyType = 1 // генерируем уникальный ключ данных
	// сохраняем данные в контексте, используя уникальный ключ
	c.DataSet(myKey, "Test data")
	// читаем данные с помощью ключа
	str := c.Data(myKey).(string)
	fmt.Println(str)
	// Output: Test data
}

func ExampleContext_Body() {
	// открываем файл
	file, err := os.Open("README.md")
	if err != nil {
		panic(err)
	}
	// устанавливаем тип отдаваемых данных
	c.ContentType = "text/markdown; charset=UTF-8"
	// отдаем содержимое файла в качестве ответа
	c.Body(file)
	// закрытие файла не обязательно, т.к. метод Body автоматически
	// закроет его, если поддерживается интрефейс io.ReadCloser
	file.Close()
}

func ExampleContext_Code() {
	// возвращаем 404 ошибку
	c.Code(404).Body(nil)
}

func ExampleContext_Parse() {
	// инициализируем формат данных для разбора
	obj := make(map[string]interface{})
	// читаем запрос и получаем данные в разобранном виде
	if err := c.Parse(&obj); err != nil {
		panic(err)
	}
}

func ExampleContext_SetHeader() {
	c.SetHeader("ETag", "ab0138")
	c.SetHeader("Location", "/user/43952945")
}

func ExampleServeMux_Handle() {
	var mux rest.ServeMux
	mux.Handle("GET", "/message/:text", rest.Func(func(c *rest.Context) {
		c.Body(rest.JSON{"message": c.Get("text")})
	}))
}

func ExampleServeMux_Handler() {
	var mux rest.ServeMux
	// в качестве обработчиков можно использовать стандартные обработчики http
	mux.Handler("GET", "/tmpfiles/",
		http.StripPrefix("/tmpfiles/", http.FileServer(http.Dir("/tmp"))))
}

func ExampleServeMux_ServeHTTP() {
	var mux rest.ServeMux
	mux.Handle("GET", "/message/:text", rest.Func(func(c *rest.Context) {
		c.Body(rest.JSON{"message": c.Get("text")})
	}))
	// т.к. поддерживается интерфейс http.Handler, то можно использовать
	// с любыми стандартными библиотеками
	http.ListenAndServe(":8080", mux)
}
