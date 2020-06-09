package models

import "net/mail"

// История отправки
type TransactionalEmail struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index;not null;"`
	EnvelopeID uint `json:"envelopeId" gorm:"index;not null;"` // собственно какое письмо отправляем

	// Удобно для транзакционных сообщений
	Subject string
	From mail.Address	// "Name" <mail@example.com>
	To string // "info@ratus.media"
}

// Отправляем письмо trEmail.To
func (trEmail TransactionalEmail) Send()  {

}
