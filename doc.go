// rest является простенькой библиотекой для быстрого создания web API для проектов.
//
// Основные ее достоинства в том, что она достаточно компактная, облегчает некоторые достаточно
// часто используемые вещи, поддерживает параметры в пути, совместима со стандартной библиотекой
// http и позволяет описывать обработчики в таком виде, как удобно мне.
//
//  mux.Handles(Handlers{
//      "/user/:id": {
//          "GET": rest.HandlerFunc(func(c *rest.Context) {
//              c.Body(rest.JSON{"user": c.Get("id")})
//          }),
//          "POST": rest.HandlerFunc(func(c *rest.Context) {
//              var data = make(rest.JSON)
//              if err := c.Parse(&data); err != nil {
//                  c.Code(500).Body(err)
//                  return
//              }
//              c.Body(rest.JSON{"user": c.Get("id"), "data": data})
//          }),
//      },
//      "/message/:text": {
//          "GET": rest.HandlerFunc(func(c *rest.Context) {
//              c.Body(rest.JSON{"message": c.Get("text")})
//          }),
//      },
//  })
//
// Вообще, библиотека написана исключительно для внутреннего использования и нет никаких
// гарантий, что она не будет время от времени серьезно изменяться. Поэтому, если вы хотите
// использовать ее в своих проектах, то делайте fork и вперед.
package rest
