package models

// Настройки домена, принадлежащего аккаунту (DKIM & SPF ... )
type Domain struct {

	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"type:int;index;not_null;"`

	Hostname string `json:"sendingHostname" gorm:"type:varchar(255);not_null;"` // ratuscrm.com, pic._domainkey.ratuscrm.com

	// DKIM values...
	DKIMPublicRSAKey string `json:"dkimPublicRsaKey" gorm:"type:text;"` // публичный ключ
	DKIMPrivateRSAKey string `json:"dkimPublicRsaKey" gorm:"type:text;"` // приватный ключ
	DKIMSelector string `json:"dkimSelector" gorm:"type:varchar(255);default:'dk1'"` // dk1

	EmailBoxes []EmailBox `json:"emailBoxes"` // доступные почтовые ящики с которых можно отправлять
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
func (account Account) GetDomains() ([]Domain, error) {
	var domains []Domain
	err := db.Preload("MailBoxes").Find(&domains, "account_id = ?", account.ID).Error
	return domains, err
}

func (account Account) CreateDomain(domain Domain) (*Domain, error) {

	// 1. Set this account an owner
	domain.AccountID = account.ID

	// 2. Checking some rules
	// todo: чо-то там
	domainNew, err :=  domain.create()
	if err != nil {
		return nil, err
	}

	return domainNew, nil
}

func (domain *Domain) AddMailBox (ebox EmailBox) (*EmailBox,error) {
	// todo: проверка на существующий контекст
	// info@ и т.д. один вариант

	// устанавливаем владельца такого же
	ebox.AccountID = domain.AccountID
	ebox.DomainID = domain.ID

	return ebox.create()
}

