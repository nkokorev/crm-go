package models

import "time"

/**
Объект платежа - кто-то, что-то вам заплатил. Или хочет заплатить. Или должен...
 */

type Payment struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Paid 	bool 		`json:"paid" gorm:"type:bool;default:false;"` // признак оплаты платежа

	// Платежи бывают разные
	OwnerType	string
	OwnerID	uint	`json:"ownerId" gorm:"type:int"` // ID в

	AmountValue float64 // объем платежа
	AmountCurrency string `json:"amountCurrency"  gorm:"type:varchar(4);default:'RUB'"` // сумма валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html


	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-" sql:"index"`

}
