package models

import (
	"bytes"
	"github.com/nkokorev/crm-go/utils"
	"html/template"
	"net/mail"
	"strings"
)

// mail of email message with template
type Envelope struct {
	Body   bytes.Buffer
	Subject string
	From mail.Address	// "Name" <mail@example.com>
	To string // no address

	tpl *template.Template
}

// Main location of email files
var locationTemplate = "files/emails/tpl"
var publishUrl = "https://ratuscrm.com/templates/publish/" // https://ratuscrm.com/templates/publish/{accountName}/{rand(5)}

// load template of email
func (env *Envelope) LoadBodyTemplate(filename string) (err error) {

	// тут можно сделать какой-то поиск среди доступных шаблонов
	env.tpl, err = template.ParseFiles(locationTemplate + filename)
	if err != nil {
		return err
	}

	return nil
}

// execute template to body
func (env *Envelope) ExecuteTemplate(T interface{}) error {

	err := env.tpl.Execute(&env.Body, T)
	if err != nil {
		return err
	}

	return nil
}

// publish Email after execute template
func (account Account) PublishEmail(env Envelope) (e *EnvelopePublish, err error) {

	// 1. Проверка (?)
	// todo

	// 2. Результат в виде html сохраняем в аккаунт
	html := env.Body.String()


	// 3. Публикуем скомпилированный шаблон по адресу https://ratuscrm.com/templates/publish
	url := publishUrl + strings.ToLower(account.Name) + "/" + utils.RandStringBytes(5)

	e, err = account.CreateEnvelopePublishes( EnvelopePublish{Url: url, Body: html} )
	if err != nil {
		return nil, err
	}


	return e, nil
}