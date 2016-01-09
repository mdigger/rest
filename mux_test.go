package rest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"testing"
)

func TestMux(t *testing.T) {
	var mux ServeMux // инициализируем обработчик запросов
	// добавляем описание обработчиков, задавая пути, методы и функции их обработки
	mux.Handles(Paths{
		// при задании путей можно использовать именованные параметры с ':'
		"/user/:id": {
			"GET": func(c *Context) error {
				// можно быстро сформировать ответ в JSON
				return c.Send(JSON{"user": c.Param("id")})
			},
			// для одного пути можно сразу задать все обработчики для разных методов
			"POST": func(c *Context) error {
				var data = make(JSON)
				// можно быстро десериализовать JSON, переданный в запросе, в объект
				if err := c.Parse(&data); err != nil {
					// возвращать ошибки тоже удобно
					return err
				}
				return c.Send(JSON{"user": c.Param("id"), "data": data})
			},
		},
		// можно одновременно описать сразу несколько путей в одном месте
		"/message/:text": {
			"GET": func(c *Context) error {
				// параметры пути получаются простым запросом
				return c.Send(JSON{"message": c.Param("text")})
			},
		},
		"/file/:name": {
			"GET": func(c *Context) error {
				// поддерживает отдачу разного типа данных, в том числе и файлов
				file, err := os.Open(c.Param("name") + ".html")
				if err != nil {
					return err
				}
				// можно получать не только именованные элементы пути, но
				// параметры, используемые в запросе
				if c.Param("format") == "raw" {
					c.ContentType = `text/plain; charset="utf-8"`
				} else {
					c.ContentType = `text/html; charset="utf-8"`
				}
				return c.Send(file) // отдаем содержимое файла
				// закрытие файла произойдет автоматически
			},
		},
		"/favicon.ico": {
			// для работы со статическими файлами определена специальная функция
			"GET": ServeFile("./favicon.ico"),
		},
	})
	// можно сразу задать базовый путь для всех URL, используемых в обработчиках
	mux.BasePath = "/api/v1"
	// можно задать глобальные заголовки для всех ответов
	mux.Headers = map[string]string{
		"X-Powered-By": "My Server",
	}
	mux.Handle("GET", "/aaa", http.NotFoundHandler)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	r, err := http.NewRequest("GET", ts.URL+mux.BasePath+"/message/test?param=name", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(dump))
}
