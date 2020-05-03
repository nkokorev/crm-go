package models

// Настройки домена, принадлежащего аккаунту (DKIM & SPF ... )
type Domain struct {

	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index;not_null;"`
	PurposeRecord string `json:"purposeRecord" gorm:"type:varchar(15);not_null;"` //Sending, Receiving, Tracking

	Type string `json:"sendingType" gorm:"type:varchar(20);default:'TXT';"` // TXT, MX, CNAME
	Hostname string `json:"sendingHostname" gorm:"type:varchar(255);not_null;"` // ratuscrm.com, pic._domainkey.ratuscrm.com
	Priority int `json:"priority" gorm:"type:int;default:10;"` // for MX - 10
	Value string `json:"sendingValue" gorm:"type:varchar(255);not_null"` // v=spf1 include:mailgun.org ~all OR k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AM... <...>

	Senders []EmailSender `json:"senders"` // доступные почтовые ящики ???
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

func (domain *Domain) AddMailBox (es EmailSender) (*EmailSender,error) {
	// todo: проверка на существующий контекст
	// info@ и т.д. один вариант

	// устанавливаем владельца такого же
	es.DomainID = domain.ID
	es.AccountID = domain.AccountID

	return es.create()
}

