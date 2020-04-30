package models

// Allowed Mail Box list of domain in Account
type MailBox struct {
	ID     uint   `json:"id" gorm:"primary_key"`

	AccountID uint `json:"accountId" gorm:"index"` // принадлежность к аккаунту
	DomainID uint `json:"domainId" gorm:"index"` // к какому домену принадлежит

	Default bool `json:"default" gorm:"type:bool;default:false"` // является ли дефолтным почтовым ящиком для домена
	Allowed bool `json:"allowed" gorm:"type:bool;default:false"` // прошел ли проверку домен на право отправлять с него почту

	FromName string `json:"fromName" gorm:"type:varchar(255);"` // RatusCRM, Admin, ООО ПК-ВТВИнженеринг
	BoxName string `json:"boxName" gorm:"type:varchar(255);default:null"` // info / mail / mailbox ... + domain name. AllowedList
	
	Domain Domain `json:"domain"`
}

func (MailBox) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&MailBox{})
	db.Exec("ALTER TABLE mail_boxes \n--     ADD CONSTRAINT uix_email_account_id_parent_id unique (email,account_id,parent_id),\n    ADD CONSTRAINT mail_boxes_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT mail_boxes_domain_id_fkey FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (mailBox MailBox) create() (*MailBox, error)  {
	err := db.Create(&mailBox).Error
	return &mailBox, err
}

func (mailBox MailBox) delete () error {
	return db.Model(MailBox{}).Where("id = ?", mailBox.ID).Delete(mailBox).Error
}

func (mailBox *MailBox) update(input MailBox) error {
	return db.Model(mailBox).Omit("id", "created_at", "deleted_at", "updated_at").Updates(&input).Error
}

// обязательно в контексте аккаунта
func (account Account) GetMailBox(id uint) (*MailBox, error) {
	var mailBox MailBox
	err := db.Preload("Domain").First(&mailBox, "id = ? AND account_id = ?", id, account.ID).Error
	return &mailBox, err
}

// возвращает все доступные адреса
func (account Account) GetMailBoxes() (*[]MailBox, error) {
	var mailBoxes []MailBox
	err := db.Preload("Domain").Find(&mailBoxes, "account_id = ?", account.ID).Error
	return &mailBoxes, err
}

// Возвращает дефолтный почтовый ящик
func (account Account) GetMailBoxDefault() (*MailBox, error) {
	var mailBox MailBox
	err := db.Preload("Domain").First(&mailBox, "account_id = ? AND default = true", account.ID).Error

	return &mailBox, err
}