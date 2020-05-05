package models

import (
	"net/mail"
)

type EmailBox struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index;not_null;"`
	PurposeRecord string `json:"purposeRecord" gorm:"type:varchar(15);default:'sending';"` //Sending, Receiving, Tracking

	Name string `json:"host" gorm:"type:varchar(255);not_null;"` // RatusCRM, Магазин 357 грамм..
	Domain string `json:"host" gorm:"type:varchar(255);not_null;"` // ratuscrm.com, 357gr.ru
	Box string `json:"boxName" gorm:"type:varchar(255);not_null;"` // info, news, mail ...
}

func (e EmailBox) GetAddress() *mail.Address {
	return &mail.Address{Name: e.Name, Address: e.Box + "@" + e.Domain}
}