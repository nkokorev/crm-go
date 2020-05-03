package models

// Сохраненные email'ы
type EnvelopePublish struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index;not_null;"`

	Url string `json:"url" gorm:"type:varchar(255);not_null"` // https://ratuscrm.com/templates/publish/{accountName}/{rand(5)}
	Body string `json:"body" gorm:"type:text;"`
}


func (EnvelopePublish) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EnvelopePublish{})
	db.Exec("ALTER TABLE envelope_publishes \n    ADD CONSTRAINT envelope_publish_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (env EnvelopePublish) create() (*EnvelopePublish, error)  {

	err := db.Create(&env).Error
	return &env, err
}

func (env EnvelopePublish) delete () error {
	return db.Model(Domain{}).Where("id = ?", env.ID).Delete(env).Error
}

func (env *EnvelopePublish) update(input interface{}) error {
	return db.Model(env).Omit("id", "created_at", "deleted_at", "updated_at").Updates(&input).Error
}

// обязательно в контексте аккаунта
func (account Account) GetEnvelopePublish(id uint) (*EnvelopePublish, error) {
	var env EnvelopePublish
	err := db.First(&env, "id = ? AND account_id = ?", id, account.ID).Error
	return &env, err
}

// возвращает все доступные домены с предзагрузкой mailboxes обязательно в контексте аккаунта
func (account Account) GetEnvelopePublishes() ([]EnvelopePublish, error) {
	var envelopes []EnvelopePublish
	err := db.Find(&envelopes, "account_id = ?", account.ID).Error
	return envelopes, err
}

func (account Account) CreateEnvelopePublishes(env EnvelopePublish) (*EnvelopePublish, error) {

	// 1. Set this account an owner
	env.AccountID = account.ID

	// 2. Checking some rules
	// todo: чо-то там
	domainNew, err :=  env.create()
	if err != nil {
		return nil, err
	}

	return domainNew, nil
}

