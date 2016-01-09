package rest

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
)

// ParamFlag определяет символ, используемый в качестве определения параметра (должен идти
// в самом начале).
const paramFlag byte = ':'

// Param описывает именованный параметр и его значение. В качестве ключа используется имя параметра
// (без символа параметра), а в качестве значения — строка из пути, соответствующая данной позиции.
//
// Я не стал использовать для параметров словарь, т.к. данный метод позволяет сохранить порядок
// следования параметров и использовать параметры с одинаковым именем.
type Param struct {
	Key, Value string
}

// record описывает информацию о пути, в котором есть параметры.
//
// Значение priority задает приоритет для сортировки ключей с одинаковым количеством параметров и
// формируется следующим образом: в старших восьми байтах содержится количество элементов с
// параметрами, а в младших — со статическими путями. Таким образом получается, что элементы
// с меньшим количеством параметров имеют более высокий приоритет и после сортировки будут
// обрабатывается позже, чем элементы с меньшим количеством параметров.
//
// Текущая реализация имеет ограничение на максимальное количество элементов пути — 32767.
// Это связано с методом хранения этого значения в свойстве priority. В принципе, ничего не мешает
// просто увеличить размер priority до uint32 и поправить соответствующие места, где он определен,
// но мне показалось, что для моих задач этого более, чем достаточно.
type record struct {
	params uint16      // количество параметров в пути
	handle interface{} // обработчик запроса
	parts  []string    // путь, разобранный на составные части
}

// records описывает список путей с параметрами и поддерживает сортировку по флагу приоритета.
type records []*record

// поддержка методов для сортировки.
func (n records) Len() int           { return len(n) }
func (n records) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n records) Less(i, j int) bool { return n[i].params < n[j].params }

// router описывает структуру для быстрого выбора обработчиков по пути запроса.
// Поддерживает как статические пути, так и пути с параметрами.
//
// Текущая реализация не привязана к конкретным типам обработчиков и может хранить любые
// объекты в качестве таких обработчиков.
type router struct {
	// хранилище статических путей, без параметров
	// в качестве ключа используется полный путь
	static map[string]interface{}
	// хранит информацию о путях с параметрами
	// в качестве ключа используется общее количество элементов пути
	fields   map[uint16]records
	maxParts uint16 // максимальное количество частей пути в определениях
}

// add добавляет описание нового пути запроса и ассоциирует его с указанным обработчиком запроса.
// Возвращает ошибку, если количество частей пути больше 32767. В качестве флага для определения
// именованных параметров используется символ ':' и '*' для завершающего параметра, который
// "забирает" в себя весь оставшийся путь.
func (r *router) add(url string, handle interface{}) error {
	parts := split(url) // нормализуем путь и разбиваем его на части
	// проверяем, что количество получившихся частей не превышает поддерживаемое количество.
	length := len(parts)
	if length > (1<<15 - 1) {
		return fmt.Errorf("path parts overflow: %d", len(parts))
	}
	var dynamic uint16 // считаем количество параметров
	for i, value := range parts {
		// if len(value) > 0 { после нормализации пути не должно быть пустых элементов
		switch value[0] {
		case byte('*'):
			if i != length-1 {
				return errors.New("catch-all parameter must be last")
			}
			dynamic |= 1 << 15 // взводим флаг *-параметра
			fallthrough
		case byte(':'):
			dynamic++ // увеличиваем счетчик параметров
		}
		// }
	}
	if dynamic == 0 { // в пути нет параметров — добавляем в статические обработчики
		if r.static == nil {
			// инициализируем статику, если не сделали этого раньше
			r.static = make(map[string]interface{})
		}
		r.static[strings.Join(parts, "/")] = handle
		return nil
	}
	level := uint16(length) // всего элементов пути
	if r.maxParts < level { // запоминаем максимальное количество определенных параметров
		r.maxParts = level
	}
	if r.fields == nil {
		// инициализируем динамические пути, если не сделали этого раньше
		r.fields = make(map[uint16]records)
	}
	// в пути есть динамические параметры — добавляем в список с параметрами
	record := &record{
		params: dynamic,
		handle: handle, // обработчик запроса
		parts:  parts,  // части пути
	}
	// сохраняем в массиве обработчиков с таким же количеством параметров
	r.fields[level] = append(r.fields[level], record)
	sort.Stable(r.fields[level]) // сортируем по количеству параметров
	return nil
}

// lookup возвращает обработчик и список именованных параметров с их значениям. Символ параметра
// из имени при этом изымается. Если подходящего обработчика не найдено, то возвращается nil.
func (r *router) lookup(url string) (interface{}, []Param) {
	parts := split(url) // нормализуем путь и разбиваем его на части
	// сначала ищем среди статических путей
	if r.static != nil { // если статические пути не определены, то пропускаем проверку
		if handle, ok := r.static[strings.Join(parts, "/")]; ok {
			return handle, nil
		}
	}
	if r.fields == nil { // если пути с параметрами не определены, то пропускаем проверку
		return nil, nil
	}
	length := uint16(len(parts))
	var total uint16
	if length > r.maxParts {
		total = r.maxParts
	} else {
		total = length
	}
	// запрашиваем список обработчиков для такого же количества элементов пути
	for l := total; l > 0; l-- {
		records := r.fields[l]
		if len(records) == 0 {
			continue // обработчики для такого пути не зарегистрированы
		}
		catchOnlyAll := l < length // флаг, что ищем только пути со "звездочкой" на конце
	nextRecord:
		for _, record := range records { // перебираем все записи с обработчиками
			if catchOnlyAll && (record.params^(1<<15) != 1<<15) {
				continue // игнорируем, если последний параметр не со звездочкой
			}
			var params []Param // сбрасываем предыдущие значения, если они были
		params:
			for i, part := range record.parts { // перебираем все части пути, заданные в обработчике
				// if len(part) > 0 { // это параметр?
				switch part[0] {
				case byte('*'):
					params = append(params, Param{
						Key:   part[1:],
						Value: strings.Join(parts[i:], "/"),
					})
					break params //
				case byte(':'):
					params = append(params, Param{
						Key:   part[1:],
						Value: parts[i],
					})
					continue // переходим к следующему элементу пути
				}
				// }
				if part != parts[i] { // статическая часть пути не совпадает с запрашиваемой
					continue nextRecord // переходим к следующему обработчику
				}
			}
			return record.handle, params // возвращаем обработчик и параметры
		}
	}
	return nil, nil // ничего подходящего не нашли
}

// split нормализует путь и возвращает его в виде частей.
func split(url string) []string {
	return strings.Split(strings.Trim(path.Clean(url), "/"), "/")
}
