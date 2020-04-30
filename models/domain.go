package models

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
)

// Настройки домена, принадлежащего аккаунту. В основном для отправки email'ов
type Domain struct {

	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index"`

	Host string `json:"host" gorm:"type:varchar(255);default:'example.com';index;unique;"` // ratuscrm.com

	MailBoxes []MailBox `json:"mailBoxes"` // доступные почтовые ящики

	DKIMRSAPublicKey string `json:"dkimRsaPublicKey" gorm:"type:varchar(1024);"`
	DKIMRSAPrivateKey string `json:"dkimRsaPrivateKey" gorm:"type:varchar(1024);"`
	DKIMSelector string `json:"dkimSelector" gorm:"type:varchar(255);"` // {dk1}._domainkey.ratuscrm.com
}

func (Domain) PgSqlCreate() {
	fmt.Println("Создаем таблицу")
	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Domain{})
	db.Exec("ALTER TABLE domains \n--     ADD CONSTRAINT uix_email_account_id_parent_id unique (email,account_id,parent_id),\n    ADD CONSTRAINT mta_settings_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (domain Domain) create() (*Domain, error)  {

	// Normalize DKIM RSA keys
	//domain.DKIMRSAPrivateKey = strings.ReplaceAll(domain.DKIMRSAPrivateKey, "\n", "")
	//domain.DKIMRSAPublicKey = strings.ReplaceAll(domain.DKIMRSAPublicKey, "\n", "")

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

func (domain *Domain) AddMailBox (mailbox MailBox) (*MailBox,error) {
	// todo: проверка на существующий контекст
	// info@ и т.д. один вариант

	// устанавливаем владельца такого же
	mailbox.DomainID = domain.ID
	mailbox.AccountID = domain.AccountID

	return mailbox.create()
}

func (domain Domain) GetPrivateKey () *rsa.PrivateKey {

	block, _ := pem.Decode([]byte(string(domain.DKIMRSAPrivateKey)))

	enc := x509.IsEncryptedPEMBlock(block)
	b := block.Bytes

	var err error
	if enc {
		log.Println("is encrypted pem block")
		b, err = x509.DecryptPEMBlock(block, nil)
		if err != nil {
			log.Fatal(err)
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(b)
	if err != nil {
		log.Fatal(err)
	}
	return key
}

// надо бы проверить эту функцию..
func (domain Domain) GetPrivateKeyByte () []byte {
	//return x509.MarshalPKCS1PrivateKey(domain.GetPrivateKey())
	return []byte(string(domain.DKIMRSAPrivateKey))
}