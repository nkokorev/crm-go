package models

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mailgun/mailgun-go"
	"html/template"
	"log"
	"net/mail"
	"time"
)

type Email struct {

	Header map[string]string
	Body   bytes.Buffer
	BodySignedDKIM []byte
	Subject string
	To mail.Address
	From mail.Address

	Tpl *template.Template // шаблон письма с которого читаются данные
}

func SendTestMessage() error {

	// создаем письмо и сразу подгружаем шаблончик
	e, err := NewEmail("example", nil)
	if err != nil { return err }

	// ввод общих данных
	e.Subject = "Тестируем отправку двум адресатам"
	e.To = mail.Address{Address: "nkokorev@rus-marketing.ru"}
	//addrs := []string{"nkokorev@rus-marketing.ru", "mail-test@ratus-dev.ru"}
	//email.SetTo("mail-test@ratus-dev.ru")
	//email.SetTo("mex388@mail.ru")
	//email.SetReturnPath("abuse@mta1@ratuscrm.com")
	//email.SetMessageId("jd7dhds73h3738")

	e.From = mail.Address{Name: "RatusCRM", Address: "info@ratuscrm.com"}

	// ++=+++++++++++++++++++++++

	yourDomain := "ratuscrm.com" // e.g. mg.yourcompany.com

	mg := mailgun.NewMailgun(yourDomain, "cd00e0c60b26be77e32a943bd5768a19-65b08458-9049e45c")

	// The message object allows you to add attachments and Bcc recipients
	message := mg.NewMessage(e.From.String(), e.Subject, "", e.To.Address)

	message.SetHtml( string( e.GetBody() ) )

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message	with a 10 second timeout
	resp, id, err := mg.Send(ctx, message)

	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Printf("Id: %s Resp: %s\n", id, resp)

	return nil
}


// возвращает новое письмо и загружает шаблон при наличии имени файла
func NewEmail(filename string, T interface{}) (*Email, error) {
	email := new(Email)

	if filename != "" {
		err := email.LoadBodyTemplate(filename + ".html", T)
		if err != nil {
			return nil, err
		}

	}

	return email, nil
}

// Загружает шаблон письма (html.template)
func (email *Email) LoadBodyTemplate(filename string, T interface{}) (err error) {

	// тут можно сделать какой-то поиск среди доступных шаблонов
	email.Tpl, err = template.ParseFiles("files/" + filename)
	if err != nil {
		return err
	}

	if err = email.ExecuteBodyTemplate(T); err != nil {
		return err
	}

	return nil
}

// подгружает данные в шаблон и результат в тело письма
func (email *Email) ExecuteBodyTemplate(T interface{}) error {
	var buf = new(bytes.Buffer)

	err := email.Tpl.Execute(buf, T)
	if err != nil {
		return err
	}

	email.Body = *buf

	return nil
}

func (email Email) GetBody() []byte {
	return email.Body.Bytes()
}

// Возвращает список заголовоков
func (email Email) GetHeaders() []string {

	var headers []string
	for k,_ := range email.Header {
		headers = append(headers, k)
	}

	return headers
}

// Возвращает все сообщение

