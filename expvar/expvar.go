// Package expvar добавляет поддержку стандартной библиотеки expvar в rest.
//
// При подключении библиотеки к проекту автоматически добавляются стандартные
// обработчики, определенные в expvar, такие как "cmdline" и "memstats". Кроме
// этого, данная библиотека так же регистрирует "goroutines", содержащее число
// запущенных горутин. И "uptime" с количеством наносекунд, прошедших с момента
// старта сервиса.
package expvar

import (
	"expvar"
	"fmt"
	"runtime"
	"time"

	"github.com/mdigger/rest"
)

// Path описывает путь обработчика для ExpVar. В отличии от стандартной
// библиотеки, вы можете изменить здесь этот путь.
var Path = "/debug/vars"

func init() {
	startTime := time.Now().UTC() // время запуска сервиса
	expvar.Publish("uptime", expvar.Func(func() interface{} {
		return int64(time.Since(startTime))
	}))
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
}

// Register регистрирует обработчик ExpVar среди обработчиков данного
// сервера. Используется для совместимости со стандартной библиотекой expvar.
func Register(m *rest.ServeMux) {
	// дублирует обработчик ExpVar.
	m.Handle("GET", Path, func(c *rest.Context) error {
		c.ContentType = "application/json; charset=utf-8"
		fmt.Fprintf(c, "{\n")
		first := true
		expvar.Do(func(kv expvar.KeyValue) {
			if !first {
				fmt.Fprintf(c, ",\n")
			}
			first = false
			fmt.Fprintf(c, "%q: %s", kv.Key, kv.Value)
		})
		fmt.Fprintf(c, "\n}\n")
		return nil
	})
}
