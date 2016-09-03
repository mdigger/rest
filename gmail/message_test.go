package gmail

import (
	"fmt"
	"testing"
)

// тестовое сообщение
var msg = &Message{
	Name:    "Тестовый пользователь",
	From:    "dmitrys@xyzrd.com",
	To:      "sedykh@gmail.com",
	Subject: "Тема сообщения",
	Body:    "Краткий комментарий к сообщению.\nСодержит несколько строк.",
}

// TestMessageText проверяем формирование текст сообщений на основании шаблона
func TestMessageText(t *testing.T) {
	// преобразуем в текст сообщения на основании шаблона
	text, err := msg.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	// выводим в консоль, чтобы проверить правильность текста
	fmt.Println(string(text))
}

// // TestMessageSend отправляет сообщение через GMail.
// func TestMessageSend(t *testing.T) {
// 	if err := Init("config.json", "token.json"); err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := msg.Send(); err != nil {
// 		t.Fatal(err)
// 	}
// }
