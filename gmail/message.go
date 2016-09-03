package gmail

import (
	"bytes"
	"encoding/base64"
	"errors"
	"mime"
	"text/template"

	gm "google.golang.org/api/gmail/v1"
)

// messageTeplate содержит разобранный и готовый для использования шаблон
// для отправки сообщения по почте.
var messageTeplate = template.Must(
	template.New("").Funcs(template.FuncMap{"qenc": func(name string) string {
		return mime.QEncoding.Encode("utf-8", name)
	}}).Parse(`From: {{qenc .Name}}
Reply-To: {{qenc .Name}} <{{.From}}>
To: {{.To}}
Subject: {{qenc .Subject}}

{{ .Body }}`))

// Message описывает формат сообщения для отправки через GMail.
type Message struct {
	Name    string // имя отправителя
	From    string // e-mail отправителя
	To      string // e-mail получателя
	Subject string // тема сообщения
	Body    string // основное содержимое письма
}

// GetText преобразует сообщения по шаблону в текст и возвращает его.
func (m Message) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := messageTeplate.Execute(&buf, m); err != nil {
		return nil, err // возвращаем ошибку
	}
	return buf.Bytes(), nil // возвращаем получившийся текст
}

// ErrGMailServiceNotInitialized описывает ошибку не инициализированного
// сервиса GMail.
var ErrGMailServiceNotInitialized = errors.New("gmail service not initialized")

// Send отправляет сообщение через GMail.
//
// Перед отправкой необходимо инициализировать сервис, вызвав функцию
// gmail.Init(), которая должна выполняться до старта сервера, потому что может
// потребовать ввода кода ответа при первой инициализации сервиса.
func (m Message) Send() error {
	if gmailService == nil || gmailService.Users == nil {
		return ErrGMailServiceNotInitialized
	}

	text, err := m.Bytes() // получаем текст сформированного сообщения
	if err != nil {
		return err
	}
	// кодируем содержимое сообщения в формат Base64
	body := base64.RawURLEncoding.EncodeToString(text)
	// формируем сообщение в формате GMail
	var gmailMessage = &gm.Message{Raw: body}
	// отправляем сообщение на сервер GMail
	_, err = gmailService.Users.Messages.Send("me", gmailMessage).Do()
	return err // возвращаем статус отправки сообщения

}
