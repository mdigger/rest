# REST API

[![GoDoc](https://godoc.org/github.com/mdigger/rest?status.svg)](https://godoc.org/github.com/mdigger/rest)
[![Build Status](https://travis-ci.org/mdigger/rest.svg)](https://travis-ci.org/mdigger/rest)
[![Coverage Status](https://coveralls.io/repos/mdigger/rest/badge.svg?branch=master&service=github)](https://coveralls.io/github/mdigger/rest?branch=master)

Библиотека для быстрого описания REST API.

```go
package main

import (
	"net/http"
	"os"

	"github.com/mdigger/rest"
)

func main() {
	var mux rest.ServeMux // инициализируем обработчик запросов
	// добавляем описание обработчиков, задавая пути, методы и функции их обработки
	mux.Handles(rest.Paths{
		// при задании путей можно использовать именованные параметры с ':'
		"/user/:id": {
			"GET": func(c *rest.Context) error {
				// можно быстро сформировать ответ в JSON
				return c.Send(rest.JSON{"user": c.Param("id")})
			},
			// для одного пути можно сразу задать все обработчики для разных методов
			"POST": func(c *rest.Context) error {
				var data = make(rest.JSON)
				// можно быстро десериализовать JSON, переданный в запросе, в объект
				if err := c.Bind(&data); err != nil {
					// возвращать ошибки тоже удобно
					return err
				}
				return c.Send(rest.JSON{"user": c.Param("id"), "data": data})
			},
		},
		// можно одновременно описать сразу несколько путей в одном месте
		"/message/:text": {
			"GET": func(c *rest.Context) error {
				// параметры пути получаются простым запросом
				return c.Send(rest.JSON{"message": c.Param("text")})
			},
		},
		"/file/:name": {
			"GET": func(c *rest.Context) error {
				// поддерживает отдачу разного типа данных, в том числе и файлов
				file, err := os.Open(c.Param("name") + ".html")
				if err != nil {
					return err
				}
				defer file.Close()
				// можно получать не только именованные элементы пути, но
				// параметры, используемые в запросе
				if c.Param("format") == "raw" {
					c.ContentType = `text/plain; charset="utf-8"`
				} else {
					c.ContentType = `text/html; charset="utf-8"`
				}
				return c.Send(file) // отдаем содержимое файла
			},
		},
		"/favicon.ico": {
			// для работы со статическими файлами определена специальная функция
			"GET": rest.File("./favicon.ico"),
		},
	}, func(c *rest.Context) error {
		// проверяем авторизацию для всех запросов, определенных выше
		login, password, ok := c.BasicAuth()
		if !ok {
			c.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.Send(rest.ErrUnauthorized)
		}
		if login != "login" || password != "password" {
			return c.Send(rest.ErrForbidden)
		}
		return nil
	})
	// можно задать глобальные заголовки для всех ответов
	mux.Headers = map[string]string{
		"X-Powered-By": "My Server",
	}
	// т.к. поддерживается интерфейс http.Handler, то можно использовать
	// с любыми стандартными библиотеками http
	http.ListenAndServe(":8080", mux)
}
```