package models

// Настройки домена, принадлежащего аккаунту. В основном для отправки email'ов
type Domain struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index"`

	DomainName string `json:"domainName" gorm:"type:varchar(255);default:'example.com';index;"` // ratuscrm.com

	MailBoxes []MailBox `json:"mailBoxes"` // доступные почтовые ящики

	DKIMRSAPublicKey string `json:"dkimRsaPublicKey" gorm:"type:varchar(255);"`
	DKIMRSAPrivateKey string `json:"dkimRsaPrivateKey" gorm:"type:varchar(255);"`
	DKIMSelector string `json:"dkimSelector" gorm:"type:varchar(255);"` // {dk1}._domainkey.ratuscrm.com
}

func (Domain) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Domain{})
	db.Exec("ALTER TABLE domains \n--     ADD CONSTRAINT uix_email_account_id_parent_id unique (email,account_id,parent_id),\n    ADD CONSTRAINT mta_settings_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (domain Domain) create() (*Domain, error)  {
	err := db.Create(&domain).Error
	return &domain, err
}

func (domain Domain) delete () error {
	return db.Model(Domain{}).Where("id = ?", domain.ID).Delete(domain).Error
}

func (domain *Domain) update(input interface{}) error {
	return db.Model(domain).Omit("id", "created_at", "deleted_at", "updated_at").Updates(&input).Error
}

// обязательно в контексте аккаунта
func (account Account) GetDomain(id uint) (*Domain, error) {
	var domain Domain
	err := db.Preload("MailBoxes").First(&domain, "id = ? AND account_id = ?", id, account.ID).Error
	return &domain, err
}

// возвращает все доступные домены с предзагрузкой mailboxes обязательно в контексте аккаунта
func (account Account) GetDomains() (*[]Domain, error) {
	var domains []Domain
	err := db.Preload("MailBoxes").Find(&domains, "account_id = ?", account.ID).Error
	return &domains, err
}