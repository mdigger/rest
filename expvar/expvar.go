// Package debugvar добавляет поддержку expvar в библиотеку rest.
package debugvar

import (
	"expvar"
	"fmt"

	"github.com/mdigger/rest"
)

// Path описывает путь обработчика для ExpVar. В отличии от стандартной
// библиотеки, вы можете изменить здесь этот путь.
var Path = "/debug/vars"

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
