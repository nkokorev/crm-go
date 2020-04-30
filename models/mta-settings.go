package models

type MTASettings struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index"`

	DomainName string `json:"domainName" gorm:"type:varchar(255);default:'example.com'"` // ratuscrm.com
	FromName string `json:"mailFrom" gorm:"type:varchar(255);"` // RatusCRM
	FromAddress string `json:"mailFrom" gorm:"type:varchar(255);"` // info@ratuscrm.com
	
	RSAPublicKey string `json:"rsaPublicKey" gorm:"type:varchar(255);"`
	RSAPrivateKey string `json:"rsaPrivateKey" gorm:"type:varchar(255);"`
	DKIMSelector string `json:"dkimSelector" gorm:"type:varchar(255);"` // dk1._domainkey.ratuscrm.com
}


func (MTASettings) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&MTASettings{})
	db.Exec("ALTER TABLE mta_settings \n--     ADD CONSTRAINT uix_email_account_id_parent_id unique (email,account_id,parent_id),\n    ADD CONSTRAINT mta_settings_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (mta MTASettings) create() (*MTASettings, error)  {
	err := db.Create(&mta).Error
	return &mta, err
}

func (mta MTASettings) delete () error {
	return db.Model(MTASettings{}).Where("id = ?", mta.ID).Delete(mta).Error
}

func (mta *MTASettings) update(input MTASettings) error {
	return db.Model(mta).Omit("ID", "created_at", "deleted_at", "updated_at").Updates(&input).Error
}

// обязательно в контексте аккаунта
func (account Account) GetMTASettings(id uint) (*MTASettings, error) {
	var mta MTASettings
	err := db.First(&mta, "id = ? AND account_id = ?", id, account.ID).Error
	return &mta, err
}