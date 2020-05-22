package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
)

//UserVerificationMethod - верификация пользователя
type UserVerificationMethod struct {
	ID uint	`json:"id" gorm:"primary_key"`
	Name string `json:"name" gorm:"type:varchar(255)"` // Регистрация по email, ...
	Tag string `json:"tag" gorm:"type:varchar(50);unique;not null;"`// email, phone, email-phone - одна из заранее определенных констант!
	Description string `json:"description" gorm:"type:varchar(255);default:null;"`// краткое описание
}

const (
	VerificationMethodEmail = "email"
	VerificationMethodPhone = "phone"
	VerificationMethodEmailAndPhone = "email+phone"
)

func (UserVerificationMethod) PgSqlCreate() {
	db.CreateTable(&UserVerificationMethod{})

	var verificationMethods = []UserVerificationMethod{
		{Name:"Email-верификация", Tag: VerificationMethodEmail, Description:"Пользователю будет необходимо перейти по ссылке в email."},
		{Name:"SMS-верификация", Tag: VerificationMethodPhone, Description:"Пользователю необходимо будет ввести код из SMS."},
		{Name:"Двойная Email+SMS верификация", Tag: VerificationMethodEmailAndPhone, Description:"Пользователю необходимо будет ввести код из SMS в специальной форме по ссылке в email."},
	}
	for _, v := range verificationMethods {
		_, err := v.Create()
		if err != nil {
			log.Fatalf("Не удалось создать тип верификации: %v", v)
		}
	}
}

// Пользователь проходит верификацию, когда поля, указанные в методе верификации пользователя в аккаунте, - надежно подтверждены самим пользователем.
func (uvt UserVerificationMethod) Create () (*UserVerificationMethod, error) {

	if len([]rune(uvt.Name)) < 1 {
		return nil, utils.Error{Message:"Не верно указаны данные", Errors: map[string]interface{}{"name":"Введите описание типа верификации"}}
	}

	if len([]rune(uvt.Tag)) < 2 {
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
	if err == gorm.ErrRecordNotFound {
		err = utils.Error{Message:"Метод верификации не найден"}
	}
	return &uvt, err
}

func GetUserVerificationTypeByCode(tag string) (*UserVerificationMethod, error) {
	uvt := UserVerificationMethod{}
	err := db.First(&uvt,"tag = ?", tag).Error
	return &uvt, err
}

func (uvt UserVerificationMethod) Delete() error {
	return db.Model(&UserVerificationMethod{}).Where("id = ?", uvt.ID).Delete(uvt).Error
}

