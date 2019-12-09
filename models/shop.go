package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
)

type Shop struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`

	Name string `json:"name"`
	Address string `json:"address"`
}

func (shop *Shop) Create () error {

	// чекаем на всякий случай ID аккаунта
	if shop.AccountID < 1 {
		return errors.New("Необходимо указать Account ID")
	}

	if shop.Name == "" {
		return utils.Error{Message:"Ошибки при создании склада", Errors: map[string]interface{} {"name":"Имя склада обязательно к заполнению"} }
	}

	return db.Omit("id").Create(shop).Error
}