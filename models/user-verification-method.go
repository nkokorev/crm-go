package models

import (
	"github.com/nkokorev/crm-go/utils"
)

type UserVerificationMethod struct {
	ID uint	`json:"id" gorm:"primary_key"`
	Name string `json:"name" gorm:"type:varchar(255)"` // Регистрация по email, ...
	Code string `json:"code" gorm:"type:varchar(50);unique;not null;"`// email, phone, email-phone
	Description string `json:"description" gorm:"type:varchar(255);default:null;"`// краткое описание
}

const (
	VerificationMethodEmail = "email"
	VerificationMethodPhone = "phone"
	VerificationMethodEmailAndPhone = "email+Phone"
)

// Пользователь проходит верификацию, когда поля, указанные в методе верификации пользователя в аккаунте, - надежно подтверждены самим пользователем.

func (uvt UserVerificationMethod) Create () (*UserVerificationMethod, error) {

	if len([]rune(uvt.Name)) < 1 {
		return nil, utils.Error{Message:"Не верно указаны данные", Errors: map[string]interface{}{"name":"Введите описание типа верификации"}}
	}

	if len([]rune(uvt.Code)) < 2 {
		return nil, utils.Error{Message:"Не верно указаны данные", Errors: map[string]interface{}{"code":"код должен быть не менее 2х символов"}}
	}

	if err := db.Create(&uvt).Error; err != nil {
		return nil, err
	}

	return &uvt, nil
}

func GetUserVerificationTypeById(id uint) (*UserVerificationMethod, error) {
	uvt := UserVerificationMethod{}
	err := db.First(&uvt,id).Error
	return &uvt, err
}

func GetUserVerificationTypeByCode(code string) (*UserVerificationMethod, error) {
	uvt := UserVerificationMethod{}
	err := db.First(&uvt,"code = ?", code).Error
	return &uvt, err
}

func (uvt UserVerificationMethod) Delete() error {
	return db.Model(&UserVerificationMethod{}).Where("id = ?", uvt.ID).Delete(uvt).Error
}

