package models

// Publish email message for share / store
type EnvelopePublished struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountId" gorm:"index;not_null;"`
	EnvelopeID uint `json:"envelopeId" gorm:"not_null;"`
	HashID string `json:"hashId" gorm:"type:varchar(255);index;not_null"` // {hashId}: search for him .: https://ratuscrm.com/templates/publish/{accountName}/{rand(5)}

	Body string `json:"body" gorm:"type:text;"` // compiled body (= html)

	Envelope Envelope // связанное сообщение, которое подлежит публикации
}


func (EnvelopePublished) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EnvelopePublished{})
	db.Exec("ALTER TABLE envelope_published \n    ADD CONSTRAINT envelope_published_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT envelope_published_envelope_id_fkey FOREIGN KEY (envelope_id) REFERENCES envelopes(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (EnvelopePublished) TableName() string {
	return "envelope_published"
}

func (env EnvelopePublished) create() (*EnvelopePublished, error)  {

	err := db.Create(&env).Error
	return &env, err
}

func (env EnvelopePublished) delete () error {
	return db.Model(Domain{}).Where("id = ?", env.ID).Delete(env).Error
}

func (env *EnvelopePublished) update(input interface{}) error {
	return db.Model(env).Omit("id", "created_at", "deleted_at", "updated_at").Updates(&input).Error
}

// обязательно в контексте аккаунта
func (account Account) GetEnvelopePublish(id uint) (*EnvelopePublished, error) {
	var env EnvelopePublished
	err := db.First(&env, "id = ? AND account_id = ?", id, account.ID).Error
	return &env, err
}

// возвращает все доступные домены с предзагрузкой mailboxes обязательно в контексте аккаунта
func (account Account) GetEnvelopePublishes() ([]EnvelopePublished, error) {
	var envelopes []EnvelopePublished
	err := db.Find(&envelopes, "account_id = ?", account.ID).Error
	return envelopes, err
}

func (account Account) CreateEnvelopePublishes(env EnvelopePublished) (*EnvelopePublished, error) {

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

