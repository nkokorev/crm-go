package models

import (
	"errors"
	"net/mail"
)

type EmailBox struct {
	ID     uint   `json:"-" gorm:"primary_key"`
	
	AccountID uint `json:"-" gorm:"type:int;index;not null;"`
	DomainID uint `json:"domainId" gorm:"type:int;index;not null;"` // обязательно!
	
	Default bool `json:"default" gorm:"type:bool;default:false"` // является ли дефолтным почтовым ящиком для домена
	Allowed bool `json:"allowed" gorm:"type:bool;default:true"` // прошел ли проверку домен на право отправлять с него почту
	
	Name string `json:"name" gorm:"type:varchar(255);not null;"` // RatusCRM, Магазин 357 грамм..
	
	// Domain string `json:"host" gorm:"type:varchar(255);not null;"` // ratuscrm.com, 357gr.ru
	Box string `json:"box" gorm:"type:varchar(255);not null;"` // info, news, mail ...

	Domain *Domain
}

func (EmailBox) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EmailBox{})
	db.Exec("ALTER TABLE email_boxes \n ADD CONSTRAINT email_boxes_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (ebox EmailBox) GetMailAddress() mail.Address {
	return mail.Address{Name: ebox.Name, Address: ebox.Box + "@" + ebox.Domain.Hostname}
}

// ########### CRUD FUNCTIONAL #########

func (ebox EmailBox) create() (*EmailBox, error)  {
	err := db.Create(&ebox).Error
	return &ebox, err
}

func (EmailBox) get(id uint) (*EmailBox, error)  {

	ebox := EmailBox{}

	err := db.First(ebox, id).Error
	if err != nil {
		return nil, err
	}
	return &ebox, nil
}

func (ebox *EmailBox) update(input interface{}) error {
	return db.Model(ebox).Omit("id", "created_at", "deleted_at", "updated_at").Updates(&input).Error
}

func (et EmailBox) delete () error {
	if et.ID < 1 {
		return errors.New("Не возможно удалить email box: не указан id")
	}
	return db.Model(EmailBox{}).Where("id = ?", et.ID).Delete(et).Error
}

// ########### END OF CRUD FUNCTIONAL #########

// ########### ACCOUNT FUNCTIONAL ###########

func (account Account) CreateEmailBox(ebox EmailBox) (*EmailBox, error) {
	ebox.AccountID = account.ID
	return ebox.create()
}

func (account Account) GetEmailBox(id uint) (*EmailBox, error) {
	var ebox EmailBox
	err := db.Preload("Domain").First(&ebox, "id = ? AND account_id = ?", id, account.ID).Error
	return &ebox, err
}

func (account Account) GetEmailBoxes() (*[]EmailBox, error) {
	eboxes := make([]EmailBox,0)
	err := db.Preload("Domain").Find(&eboxes, "account_id = ?", account.ID).Error
	return &eboxes, err
}

func (account Account) GetEmailBoxDefault() (*EmailBox, error) {
	var eboxes EmailBox
	err := db.Preload("Domain").First(&eboxes, "account_id = ? AND default = true", account.ID).Error

	return &eboxes, err
}

// ########### END OF ACCOUNT FUNCTIONAL ###########

// ########### EMAIL TEMPLATE FUNCTIONAL ###########
// ########### EMAIL TEMPLATE FUNCTIONAL ###########