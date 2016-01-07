package rest_test

import (
	"fmt"
	"net/http"
	"os"

	"github.com/geotrace/rest"
)

var c = new(rest.Context) // test context

func Example() {
	var mux rest.ServeMux // инициализируем обработчик запросов
	// добавляем описание обработчиков, задавая пути, методы и функции их обработки
	mux.Handles(rest.Paths{
		// при задании путей можно использовать именованные параметры с ':'
		"/user/:id": {
			"GET": func(c *rest.Context) {
				// можно быстро сформировать ответ в JSON
				c.Send(rest.JSON{"user": c.Get("id")})
			},
			// для одного пути можно сразу задать все обработчики для разных методов
			"POST": func(c *rest.Context) {
				var data = make(rest.JSON)
				// можно быстро десериализовать JSON, переданный в запросе, в объект
				if err := c.Parse(&data); err != nil {
					// возвращать ошибки тоже удобно
					c.Status(500).Send(err)
					return
				}
				c.Send(rest.JSON{"user": c.Get("id"), "data": data})
			},
		},
		// можно одновременно описать сразу несколько путей в одном месте
		"/message/:text": {
			"GET": func(c *rest.Context) {
				// параметры пути получаются простым запросом
				c.Send(rest.JSON{"message": c.Get("text")})
			},
		},
		"/file/:name": {
			"GET": func(c *rest.Context) {
				// поддерживает отдачу разного типа данных, в том числе и файлов
				file, err := os.Open(c.Get("name") + ".html")
				if err != nil {
					c.Status(404).Send(nil)
					return
				}
				// можно получать не только именованные элементы пути, но
				// параметры, используемые в запросе
				if c.Get("format") == "raw" {
					c.ContentType = `text; charset="utf-8"`
				} else {
					c.ContentType = `text/html; charset="utf-8"`
				}
				c.Send(file) // отдаем содержимое файла
				// закрытие файла произойдет автоматически
			},
		},
	})
	// можно сразу задать базовый путь для всех URL, используемых в обработчиках
	mux.BasePath = "/api/v1"
	// т.к. поддерживается интерфейс http.Handler, то можно использовать
	// с любыми стандартными библиотеками http
	http.ListenAndServe(":8080", mux)
}

func ExampleContext_SetData() {
	type myKeyType byte     // определяем собственный тип данных
	var myKey myKeyType = 1 // генерируем уникальный ключ данных
	// сохраняем данные в контексте, используя уникальный ключ
	c.SetData(myKey, "Test data")
	// читаем данные с помощью ключа
	str := c.Data(myKey).(string)
	fmt.Println(str)
	// Output: Test data
}

func ExampleContext_Send_file() {
	// открываем файл
	file, err := os.Open("README.md")
	if err != nil {
		c.Status(500).Send(err)
		return
	}
	// устанавливаем тип отдаваемых данных
	c.ContentType = "text/markdown; charset=UTF-8"
	// отдаем содержимое файла в качестве ответа
	c.Send(file)
	// закрытие файла не обязательно, т.к. метод Send автоматически
	// закроет его, если поддерживается интрефейс io.ReadCloser
	file.Close()
}

func ExampleContext_Status() {
	// возвращаем 404 ошибку
	c.Status(404).Send(nil)
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
	mux.Handle("GET", "/message/:text", func(c *rest.Context) {
		c.Send(rest.JSON{"message": c.Get("text")})
	})
}

func ExampleServeMux_Handler() {
	var mux rest.ServeMux
	// в качестве обработчиков можно использовать стандартные обработчики http
	mux.Handler("GET", "/tmpfiles/",
		http.StripPrefix("/tmpfiles/", http.FileServer(http.Dir("/tmp"))))
}

func ExampleServeMux_Handles() {
	var mux rest.ServeMux
	mux.Handles(rest.Paths{
		"/user/:id": {
			"GET":  user.get,
			"POST": user.post,
		},
		"/message/:text": {"GET": getMessage},
		"/file/:name":    {"GET": secure(getFile)},
	})
}

type User struct{}

func (User) get(*rest.Context)           {}
func (User) post(*rest.Context)          {}
func secure(h rest.Handler) rest.Handler { return h }

var (
	user       User
	getMessage = func(*rest.Context) {}
	getFile    = getMessage
)

func ExampleServeMux_ServeHTTP() {
	var mux rest.ServeMux
	mux.Handle("GET", "/message/:text", func(c *rest.Context) {
		c.Send(rest.JSON{"message": c.Get("text")})
	})
	// т.к. поддерживается интерфейс http.Handler, то можно использовать
	// с любыми стандартными библиотеками
	http.ListenAndServe(":8080", mux)
}
