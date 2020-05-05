package models

// Allowed Sender mail of domain (Account context)
// Senders info, news, ...
type EmailSender struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index"` // принадлежность к аккаунту
	DomainID uint `json:"domainId" gorm:"index"` // к какому домену принадлежит

	Default bool `json:"default" gorm:"type:bool;default:false"` // является ли дефолтным почтовым ящиком для домена
	Allowed bool `json:"allowed" gorm:"type:bool;default:false"` // прошел ли проверку домен на право отправлять с него почту

	FromName string `json:"fromName" gorm:"type:varchar(255);"` // RatusCRM, Admin, ООО ПК-ВТВИнженеринг
	BoxName string `json:"boxName" gorm:"type:varchar(255);default:null"` // info / mail / mailbox ... + domain name. AllowedList
	
	Domain Domain `json:"domain"`
}

func (EmailSender) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EmailSender{})
	db.Exec("ALTER TABLE email_senders \n--     ADD CONSTRAINT uix_email_account_id_parent_id unique (email,account_id,parent_id),\n    ADD CONSTRAINT mail_boxes_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT mail_boxes_domain_id_fkey FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (es EmailSender) create() (*EmailSender, error)  {
	err := db.Create(&es).Error
	return &es, err
}

func (es EmailSender) delete () error {
	return db.Model(EmailSender{}).Where("id = ?", es.ID).Delete(es).Error
}

func (es *EmailSender) update(input EmailSender) error {
	return db.Model(es).Omit("id", "created_at", "deleted_at", "updated_at").Updates(&input).Error
}

// обязательно в контексте аккаунта
