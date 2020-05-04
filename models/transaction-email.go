package models

import "net/mail"

// История отправки
type TransactionalEmail struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index;not_null;"`
	EnvelopeID uint `json:"envelopeId" gorm:"index;not_null;"` // собственно какое письмо отправляем

	// Удобно для транзакционных сообщений
	Subject string
	From mail.Address	// "Name" <mail@example.com>
	To string // "info@ratus.media"
}

// Отправляем письмо trEmail.To
func (trEmail TransactionalEmail) Send()  {

}
