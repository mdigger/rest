package rest

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"sync"
)

var (
	deflatePool sync.Pool // Кеш deflate-компрессоров
	gzipPool    sync.Pool // Кеш gzip-компрессоров
)

// gzipGet возвращает из кеша или создает новый Writer.
func gzipGet(dst io.Writer) (writer *gzip.Writer) {
	// Запрашиваем новый Writer из кеша
	if w := gzipPool.Get(); w != nil {
		writer = w.(*gzip.Writer) // Приводим его к формату
		writer.Reset(dst)         // Сбрасываем состояние
	} else {
		writer = gzip.NewWriter(dst) // Создаем новый
	}
	return writer
}

// gzipPut закрывает и возвращает Writer обратно в кеш.
func gzipPut(writer io.Closer) {
	writer.Close()       // Закрываем
	gzipPool.Put(writer) // Возвращаем в кеш
}

// deflateGet возвращает из кеша или создает новый Writer.
func deflateGet(dst io.Writer) (writer *flate.Writer) {
	// Запрашиваем новый Writer из кеша
	if w := deflatePool.Get(); w != nil {
		writer = w.(*flate.Writer) // Приводим его к формату
		writer.Reset(dst)          // Сбрасываем состояние
	} else {
		writer, _ = flate.NewWriter(dst, flate.DefaultCompression) // Создаем новый
	}
	return writer
}

// deflatePut закрывает и возвращает Writer обратно в кеш.
func deflatePut(writer io.Closer) {
	writer.Close()          // Закрываем
	deflatePool.Put(writer) // Возвращаем в кеш
}
