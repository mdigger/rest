package rest

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"runtime"
	"sort"
	"strings"
)

// param описывает именованный параметр и его значение. В качестве ключа
// используется имя параметра (без символа параметра), а в качестве значения —
// строка из пути, соответствующая данной позиции.
//
// Я не стал использовать для параметров словарь, т.к. данный метод позволяет
// сохранить порядок следования и использовать параметры с одинаковым именем.
type param struct {
	Key, Value string
}

// record описывает информацию о пути, в котором есть параметры.
type record struct {
	params  uint16   // количество параметров; старший бит — параметр динамический
	parts   []string // путь, разобранный на составные части
	handler Handler  // обработчик запроса
}

// records описывает список путей с параметрами и поддерживает сортировку по
// флагу приоритета.
type records []*record

// поддержка методов для сортировки.
func (n records) Len() int           { return len(n) }
func (n records) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n records) Less(i, j int) bool { return n[i].params < n[j].params }

// router описывает структуру для быстрого выбора обработчиков по пути запроса.
// Поддерживает как статические пути, так и пути с параметрами.
type router struct {
	// хранилище статических путей, без параметров;
	// в качестве ключа используется полный путь
	static map[string]Handler
	// хранит информацию о путях с параметрами;
	// в качестве ключа используется общее количество элементов пути
	fields map[uint16]records
	// максимальное количество частей пути во всех определениях
	maxParts uint16
	// позиция, в которой встречается самый ранний динамический параметр
	catchAll uint16
}

// split нормализует путь и возвращает его в виде частей.
func split(url string) []string {
	return strings.SplitAfter(strings.TrimPrefix(path.Clean(url), "/"), "/")
}

// add добавляет новый обработчик для указанного пути. В описании пути можно
// использовать именованные параметры (начинаются с символа ':') и завершающий
// именованный параметр (начинается с '*'), который указывает, что url может
// быть длиннее. В последнем случае вся остальная часть пути будет включена
// в данный параметр. Параметр со звездочкой, если указан, должен быть самым
// последним параметром пути.
//
// Возвращает ошибку, если обработчик не определен (nil), если количество
// элементов пути в url больше 32768 или параметр со звездочкой используется не
// в самом последнем элементе пути.
func (r *router) add(url string, handler Handler) error {
	if handler == nil {
		return errors.New("nil handler")
	}
	parts := split(url) // нормализуем путь и разбиваем его на части
	// проверяем, что количество получившихся частей не превышает поддерживаемое
	// количество
	length := len(parts)
	if length > (1<<15 - 1) {
		return fmt.Errorf("path parts overflow: %d", len(parts))
	}
	level := uint16(length) // всего элементов пути
	// считаем количество параметров в определении пути
	var params uint16
	for i, value := range parts {
		if value == "" {
			continue
		}
		switch value[0] {
		case byte(':'):
			params++ // увеличиваем счетчик параметров
		case byte('*'):
			// такой параметр должен быть самым последним в определении путей
			if i != length-1 {
				return errors.New("catch-all parameter must be last")
			}
			params |= 1 << 15 // взводим флаг catchAll параметра
			if r.catchAll == 0 || r.catchAll > level {
				// это самый ранний catchAll параметр, который нам встретился —
				// сохраняем его позицию
				r.catchAll = level
			}
		}
	}
	// в пути нет параметров — добавляем в статические обработчики
	if params == 0 {
		if r.static == nil {
			r.static = make(map[string]Handler)
		}
		r.static[strings.Join(parts, "")] = handler
		return nil
	}
	// запоминаем максимальное количество элементов пути во всех определениях
	if r.maxParts < level {
		r.maxParts = level
	}
	// инициализируем динамические пути, если не сделали этого раньше
	if r.fields == nil {
		r.fields = make(map[uint16]records)
	}
	// добавляем в массив обработчиков с таким же количеством параметров
	r.fields[level] = append(r.fields[level], &record{params, parts, handler})
	sort.Stable(r.fields[level]) // сортируем по количеству параметров
	return nil
}

// lookup возвращает обработчик и список именованных параметров с их значениям.
// Если подходящего обработчика не найдено, то возвращается nil.
func (r *router) lookup(url string) (Handler, []param) {
	parts := split(url) // нормализуем путь и разбиваем его на части
	// сначала ищем среди статических путей; если статические пути не
	// определены, то пропускаем проверку
	if r.static != nil {
		if handler, ok := r.static[strings.Join(parts, "")]; ok {
			return handler, nil
		}
	}
	// если пути с параметрами не определены, то на этом заканчиваем проверку
	if r.fields == nil {
		return nil, nil
	}
	length := uint16(len(parts)) // вычисляем количество элементов пути
	// наши определения могут быть короче, если используются catchAll параметры,
	// поэтому вычисляем с какой длины начинать
	var total uint16
	// если длина запроса больше максимальной длины определений, то нужно
	// замахиваться на меньшее...
	if length > r.maxParts {
		// если нет динамических параметров, то ничего и не подойдет,
		// потому что наш запрос явно длиннее
		if r.catchAll == 0 {
			return nil, nil
		}
		total = r.maxParts // начнем с максимального определения пути
	} else {
		total = length // наш запрос короче самого длинного определения
	}
	// запрашиваем список обработчиков для такого же количества элементов пути
	for l := total; l > 0; l-- {
		// проверяем, что на этом уровне динамические пути еще встречаются
		// if l < r.catchAll {
		// 	log.Println(111)
		// 	break // больше нет динамических параметров дальше
		// }
		records := r.fields[l] // получаем определения путей для данной длины
		if len(records) == 0 {
			// обработчики для такой длины пути не зарегистрированы —
			// переходим к более короткому пути
			continue
		}
	nextRecord:
		// обработчики есть — перебираем все записи с ними
		for _, record := range records {
			// если наш путь длиннее обработчика, а он не содержит catchAll
			// параметра, то он точно нам не подойдет
			if l < length && record.params>>15 != 1 {
				continue
			}
			// здесь мы будем собирать значения параметров к данному запросу
			// если ранее они были не пустые от другого обработчика, то
			// сбрасываем их
			var params []param
		params:
			// перебираем все части пути, заданные в обработчике
			for i, part := range record.parts {
				switch part[0] {
				case byte(':'): // это одиночный параметр
					params = append(params, param{
						// имя будет без ':' в начале и без возможного '/' в конце
						Key: strings.TrimSuffix(part[1:], "/"),
						// значением берем элемент пути без возможного '/' в конце
						Value: strings.TrimSuffix(parts[i], "/"),
					})
					continue // переходим к следующему элементу пути
				case byte('*'): // это параметр, который заберет все
					params = append(params, param{
						Key: part[1:], // исключаем '*' из имени
						// добавляем весь оставшийся путь
						Value: strings.Join(parts[i:], ""),
					})
					break params // больше ловить нечего — нашли
				}
				// статическая часть пути не совпадает с запрашиваемой
				if part != parts[i] {
					// переходим к следующему обработчику
					continue nextRecord
				}
			}
			// возвращаем найденный обработчик и заполненные параметры
			return record.handler, params
		}
	}
	// сюда мы попадаем, если так ничего подходящего и не нашли
	return nil, nil
}

// path возвращает список элементов пути, связанных с данным обработчиком.
// Если обработчик связан с несколькими путями, то вернется самый первый.
func (r *router) path(handler Handler) []string {
	// получаем адрес обработчика
	he := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Entry()
	// перебираем статические пути
	for url, handler := range r.static {
		// сравниваем адреса методов
		if he == runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Entry() {
			return split(url) // нашли нужный адрес - возвращаем элементы пути
		}
	}
	// перебираем все пути с параметрами
	for _, records := range r.fields {
		for _, record := range records {
			// сравниваем адреса методов
			if he == runtime.FuncForPC(reflect.ValueOf(record.handler).Pointer()).Entry() {
				return record.parts // возвращаем элементы пути
			}
		}
	}
	return nil // данный обработчик не зарегистрирован
}
