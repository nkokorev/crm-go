package models

import (
	"bytes"
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"html/template"
	"time"
)

// Template of email body message
type EmailTemplate struct {

	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index;not_null;"`

	Name string `json:"name" gorm:"type:varchar(255);not_null"` // inside name of mail
	Body string `json:"file" gorm:"type:text;"` // сам шаблон письма

	// GORM vars
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt" sql:"index"`
}

func (EmailTemplate) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EmailTemplate{})
	db.Exec("ALTER TABLE email_templates \n ADD CONSTRAINT email_templates_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

// ########### CRUD FUNCTIONAL #########

func (et EmailTemplate) create() (*EmailTemplate, error)  {
	err := db.Create(&et).Error
	return &et, err
}

func (EmailTemplate) get(id uint) (*EmailTemplate, error)  {

	et := EmailTemplate{}

	err := db.First(et, id).Error
	if err != nil {
		return nil, err
	}
	return &et, nil
}

func (et *EmailTemplate) update(input interface{}) error {
	return db.Model(et).Omit("id", "created_at", "deleted_at", "updated_at").Updates(&input).Error
}

func (et EmailTemplate) delete () error {
	return db.Model(EmailTemplate{}).Where("id = ?", et.ID).Delete(et).Error
}

// ########### ACCOUNT FUNCTIONAL ###########

func (account Account) CreateEmailTemplate(et EmailTemplate) (*EmailTemplate, error) {
	et.AccountID = account.ID
	return et.create()
}

func (account Account) DeleteEmailTemplate(et EmailTemplate) (error) {
	if et.AccountID != account.ID {
		return errors.New("Шаблон принадлежит другому аккаунту")
	}
	return et.delete()
}

func (account Account) GetEmailTemplate(id uint) (*EmailTemplate, error) {

	et, err := (EmailTemplate{}).get(id)
	if err != nil {
		return nil, err
	}

	if et.AccountID != account.ID {
		return nil, errors.New("Шаблон принадлежит другому аккаунту")
	}

	return et, nil

}

// ########### END OF ACCOUNT FUNCTIONAL ###########

// Возвращает тело письма в формате string в кодировке HTML, учитывая переменные в T[map]
func (et EmailTemplate) GetHTML(T map[string](string)) (html string, err error) {
	body := new(bytes.Buffer)

	// Parse template from Database
	tpl, err := template.ParseGlob(et.Body)
	if err != nil {
		return "", err
	}

	// Компиляция шаблона с переменными
	err = tpl.Execute(body, T)
	if err != nil {
		return "", err
	}

	return body.String(), nil
}

// publish Email after execute template
func (account Account) PublishEmail(et EmailTemplate, T map[string](string)) (e *EnvelopePublished, err error) {

	// 1. Проверка (?)
	// todo

	// 2. Результат в виде html сохраняем в аккаунт
	html, err := et.GetHTML(T)
	if err != nil {
		return nil, err
	}
	
	e, err = account.CreateEnvelopePublishes( EnvelopePublished{HashID: utils.RandStringBytes(5), Body: html} )
	if err != nil {
		return nil, err
	}


	return e, nil
}