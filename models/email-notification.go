package models

import (
	_ "text/template"
	"html/template"
	"net/mail"
)

/**
	1. Единая точка входа для отправки транзакционных уведомлений (notifications)
	2. Email может принадлежать пользователю <user@example.com>, компании (info@example.com), а также незарегистрированному пользователю <unknow@example.com>
	2.
*/

type Attachment struct {
	// Name must be set to a valid file name.
	Name      string
	Data      []byte
	ContentID string
}

type EmailNotification struct {
	ID uint `json:"id"`

	Template template.Template // body template
	Subject template.Template // subject (text tpl)
	Variables map[string]interface{} // переменные для шаблона

	From mail.Address
	To mail.Address // не массив! + хранит переменные <Name, Address>

	Attachment Attachment
}

func (en *EmailNotification) Send () error {

	return nil
}

func (en *EmailNotification) TemplateLoad (path string) (*template.Template, error) {
	return nil, nil
}