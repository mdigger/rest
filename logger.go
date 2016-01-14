package rest

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	isTTY     bool      // флаг, что поддерживается вывод в цвете
	accessLog io.Writer // вывод в лог
)

func init() {
	// инициализируем вывод в лог
	SetLogger(os.Stderr)
}

// SetLogger позволяет определить вывод лога обработки запросов и ошибок.
// Если установлен флаг Debug, то в лог так же пишутся все запросы, которые
// вызвали ошибку и дамп вызовов функций, приведших к panic.
func SetLogger(out io.Writer) {
	if out, ok := out.(*os.File); ok {
		fi, err := out.Stat()
		if err == nil {
			m := os.ModeDevice | os.ModeCharDevice
			isTTY = fi.Mode()&m == m
		}
	} else {
		isTTY = false
	}
	accessLog = out
}

// log выводит информацию в лог, если он определен.
func (c *Context) log() {
	if accessLog == nil {
		return // не выводим в лог, если он не определен
	}
	stop := time.Now()                   // время окончания обработки запроса
	buf := buffers.Get().(*bytes.Buffer) // формируем буфер для генерации лога
	buf.Reset()
	// время окончания обработки и вывода в лог
	buf.WriteString(stop.Format("2006/01/02 15:04:05"))
	// адрес пользователя
	remoteAddr := c.Request.Header.Get("X-Real-IP")
	// Если IP-адрес путой, то смотрим на адрес proxy
	if remoteAddr == "" {
		remoteAddr = c.Request.Header.Get("X-Forwarded-For")
	}
	// Если и этот адрес не указан, то читаем адрес socket
	if remoteAddr == "" {
		remoteAddr, _, _ = net.SplitHostPort(c.Request.RemoteAddr)
	}
	// адрес пользователя, продолжительность обработки и размер переданных данных
	fmt.Fprintf(buf, " %-15s %12v %9.3fKb ",
		remoteAddr, stop.Sub(c.started), float32(c.size)/1024)
	// Определяем цвет для вывода статуса, в зависимости от кода.
	if isTTY {
		buf.WriteString("\x1b[3")
		switch {
		case c.status < 200:
			buf.WriteRune('4') // blue
		case c.status < 300:
			buf.WriteRune('2') // green
		case c.status < 400:
			buf.WriteRune('3') // yellow
		case c.status < 500:
			buf.WriteRune('5') // magenta
		default:
			buf.WriteRune('1') // red
		}
		buf.WriteString(";2m")
	}
	fmt.Fprintf(buf, "%3d", c.status) // код ответа сервера
	if isTTY {
		buf.WriteString("\x1b[0m")
	}
	fmt.Fprintf(buf, " %7s ", c.Request.Method) // метод HTTP-запроса
	buf.WriteString(c.URL.RequestURI())         // URL
	if Debug {
		hn := c.Data(dataSet(137))
		if name, ok := hn.(string); ok {
			fmt.Fprintf(buf, " {%s}", name) // имя функции обработчика
		}
	}
	buf.WriteRune('\n')
	buf.WriteTo(accessLog)
	buffers.Put(buf)
}

// errorLog выводит в лог информацию об ошибке
func (c *Context) errorLog(err interface{}, trace bool) {
	if accessLog == nil {
		return // не выводим в лог, если он не определен
	}
	buf := buffers.Get().(*bytes.Buffer) // формируем буфер для генерации лога
	buf.Reset()
	// время окончания обработки и вывода в лог
	buf.WriteString(time.Now().Format("2006/01/02 15:04:05"))
	// публикуем саму ошибку
	if isTTY {
		buf.WriteString(" \x1b[31mError:\x1b[0m ")
	} else {
		buf.WriteString(" Error: ")
	}
	fmt.Fprintf(buf, "%v\n", err)
	// добавляем информацию о файле и строке, где произошла ошибка
	callLevel := 2
	if trace {
		callLevel = 4
	}
	if _, file, line, ok := runtime.Caller(callLevel); ok && file != "<autogenerated>" {
		if isTTY {
			buf.WriteString("## \x1b[31mFile:\x1b[0m ")
		} else {
			buf.WriteString("## File: ")
		}
		fmt.Fprintf(buf, "%s:%d\n", file, line)
	}
	if Debug {
		hn := c.Data(dataSet(137))
		if name, ok := hn.(string); ok {
			if isTTY {
				buf.WriteString("## \x1b[31mHandler:\x1b[0m ")
			} else {
				buf.WriteString("## Handler: ")
			}
			buf.WriteString(name + "\n")
		}
		if trace {
			traceBuffer := make([]byte, 1<<16)
			n := runtime.Stack(traceBuffer, false)
			if isTTY {
				buf.WriteString("## \x1b[31mStack dump\x1b[0m\n")
			} else {
				buf.WriteString("## Stack dump:\n")
			}
			buf.Write(traceBuffer[:n])
			buf.WriteString(strings.Repeat("-", 80) + "\n")
		}
	}
	if dump, err := httputil.DumpRequest(c.Request, true); err == nil {
		if isTTY {
			buf.WriteString("## \x1b[31mRequest dump\x1b[0m\n")
		} else {
			buf.WriteString("## Request dump:\n")
		}
		buf.Write(dump)
		buf.WriteRune('\n')
	}
	buf.WriteTo(accessLog)
	buffers.Put(buf)
}
