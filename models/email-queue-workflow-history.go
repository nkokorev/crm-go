package models

import (
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// email queue  workflow email templates 
type EmailQueueWorkflowHistory struct {

	Id     		uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Имя очереди (Label)
	Name	string	`json:"name" gorm:"type:varchar(128);not null;"` // Welcome, Onboarding, ...

	// В работе данное письмо в указанной серии
	Status 	bool 	`json:"status" gorm:"type:bool;default:false;"`
	Order 	uint 	`json:"order" gorm:"type:int;not null;"` // порядок

	EmailQueueId	uint	`json:"emailQueueId" gorm:"type:int;"`
	EmailQueue		EmailQueue `json:"emailQueue"`

	EmailTemplateId	uint	`json:"emailTemplateId" gorm:"type:int;"`
	EmailTemplate	EmailTemplate `json:"emailTemplate"`

	// График: Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday | weekends, workday
	// Schedule	string `json:"emailTemplate"`     `json:"switchProducts"`
	// 1- mondey, workday = 8, weekend = 89
	Schedule	postgres.Jsonb `json:"schedule" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`

	// В какое время следует отправлять электронные письма:
	// инста отправка GateStart = GateEnd = null
	// В указанное время: GateStart = <...>, GateEnd = null
	// В указанный промежуток: между GateStart <> GateEnd
	GateStart 	time.Time // << учитывается только время [0-24]
	GateEnd		time.Time // << учитывается только время [0-24]

	// Что делать, если Gate не подходит? перенести на 1-24 часа / пропустить письмо и перейти к следующему


	// Внутреннее время
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (EmailQueueWorkflowHistory) PgSqlCreate() {
	db.CreateTable(&EmailQueueWorkflowHistory{})
	db.Model(&EmailQueueWorkflowHistory{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&EmailQueueWorkflowHistory{}).AddForeignKey("email_queue_id", "email_queues(id)", "CASCADE", "CASCADE")
	db.Model(&EmailQueueWorkflowHistory{}).AddForeignKey("email_template_id", "email_templates(id)", "CASCADE", "CASCADE")
	// db.Model(&EmailQueueWorkflowHistory{}).AddForeignKey("refunded_amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
}
func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) BeforeCreate(scope *gorm.Scope) error {
	emailQueueWorkflowHistory.Id = 0

	// PublicId
	lastIdx := uint(0)
	var ord Order

	err := db.Where("account_id = ?", emailQueueWorkflowHistory.AccountId).Select("public_id").Last(&ord).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err}
	if err == gorm.ErrRecordNotFound {
		lastIdx = 0
	} else {
		lastIdx = ord.PublicId
	}
	emailQueueWorkflowHistory.PublicId = lastIdx + 1

	return nil
}

func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) AfterCreate(scope *gorm.Scope) (error) {
	// event.AsyncFire(Event{}.PaymentCreated(emailQueueWorkflowHistory.AccountId, emailQueueWorkflowHistory.Id))
	return nil
}
func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) AfterUpdate(tx *gorm.DB) (err error) {

	// event.AsyncFire(Event{}.PaymentUpdated(emailQueueWorkflowHistory.AccountId, emailQueueWorkflowHistory.Id))

	return nil
}
func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) AfterDelete(tx *gorm.DB) (err error) {
	// event.AsyncFire(Event{}.PaymentDeleted(emailQueueWorkflowHistory.AccountId, emailQueueWorkflowHistory.Id))
	return nil
}
func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) AfterFind() (err error) {
	return nil
}

// ############# Entity interface #############
func (emailQueueWorkflowHistory EmailQueueWorkflowHistory) GetId() uint { return emailQueueWorkflowHistory.Id }
func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) setId(id uint) { emailQueueWorkflowHistory.Id = id }
func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) setPublicId(publicId uint) { emailQueueWorkflowHistory.PublicId = publicId }
func (emailQueueWorkflowHistory EmailQueueWorkflowHistory) GetAccountId() uint { return emailQueueWorkflowHistory.AccountId }
func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) setAccountId(id uint) { emailQueueWorkflowHistory.AccountId = id }
func (EmailQueueWorkflowHistory) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (emailQueueWorkflowHistory EmailQueueWorkflowHistory) create() (Entity, error)  {
	
	wb := emailQueueWorkflowHistory
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	err := wb.GetPreloadDb(false,false, true).First(&wb, wb.Id).Error
	if err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (EmailQueueWorkflowHistory) get(id uint) (Entity, error) {

	var emailQueueWorkflowHistory EmailQueueWorkflowHistory

	err := emailQueueWorkflowHistory.GetPreloadDb(false,false, true).First(&emailQueueWorkflowHistory, id).Error
	if err != nil {
		return nil, err
	}
	return &emailQueueWorkflowHistory, nil
}
func (EmailQueueWorkflowHistory) getByExternalId(externalId string) (*EmailQueueWorkflowHistory, error) {
	emailQueueWorkflowHistory := EmailQueueWorkflowHistory{}

	err := emailQueueWorkflowHistory.GetPreloadDb(false,false,true).First(&emailQueueWorkflowHistory, "external_id = ?", externalId).Error
	if err != nil {
		return nil, err
	}
	return &emailQueueWorkflowHistory, nil
}
func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) load() error {
	if emailQueueWorkflowHistory.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailQueueWorkflowHistory - не указан  Id"}
	}

	err := emailQueueWorkflowHistory.GetPreloadDb(false,false, true).First(emailQueueWorkflowHistory,emailQueueWorkflowHistory.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) loadByPublicId() error {
	
	if emailQueueWorkflowHistory.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailQueueWorkflowHistory - не указан  Id"}
	}

	if err := emailQueueWorkflowHistory.GetPreloadDb(false,false, true).First(emailQueueWorkflowHistory, "public_id = ?", emailQueueWorkflowHistory.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (EmailQueueWorkflowHistory) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return EmailQueueWorkflowHistory{}.getPaginationList(accountId, 0, 25, sortBy, "")
}

func (EmailQueueWorkflowHistory) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	webHooks := make([]EmailQueueWorkflowHistory,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&EmailQueueWorkflowHistory{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailQueueWorkflowHistory{}).GetPreloadDb(true,false,true).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&EmailQueueWorkflowHistory{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailQueueWorkflowHistory{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i,_ := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}

func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) update(input map[string]interface{}) error {
	return emailQueueWorkflowHistory.GetPreloadDb(false,false,false).Where("if = ?", emailQueueWorkflowHistory.Id).Omit("id", "account_id").Updates(input).Error
}

func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) delete () error {

	return emailQueueWorkflowHistory.GetPreloadDb(true,false,false).Where("id = ?", emailQueueWorkflowHistory.Id).Delete(emailQueueWorkflowHistory).Error
}
// ######### END CRUD Functions ############

func (emailQueueWorkflowHistory *EmailQueueWorkflowHistory) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&emailQueueWorkflowHistory)
	} else {
		_db = _db.Model(&EmailQueueWorkflowHistory{})
	}

	if preload {
		// return _db.Preload("PaymentAmount")
		return _db
	} else {
		return _db
	}
}

