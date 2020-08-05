package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// M <> M : email queue  <> email templates
type EmailQueueEmailTemplate struct {

	Id     		uint   	`json:"id" gorm:"primary_key"`
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

func (EmailQueueEmailTemplate) PgSqlCreate() {
	db.CreateTable(&EmailQueueEmailTemplate{})
	db.Model(&EmailQueueEmailTemplate{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&EmailQueueEmailTemplate{}).AddForeignKey("email_queue_id", "email_queues(id)", "CASCADE", "CASCADE")
	db.Model(&EmailQueueEmailTemplate{}).AddForeignKey("email_template_id", "email_templates(id)", "CASCADE", "CASCADE")
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) BeforeCreate(scope *gorm.Scope) error {
	emailQueueEmailTemplate.Id = 0
	return nil
}

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) AfterCreate(scope *gorm.Scope) (error) {
	// event.AsyncFire(Event{}.PaymentCreated(emailQueueEmailTemplate.AccountId, emailQueueEmailTemplate.Id))
	return nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) AfterUpdate(tx *gorm.DB) (err error) {

	// event.AsyncFire(Event{}.PaymentUpdated(emailQueueEmailTemplate.AccountId, emailQueueEmailTemplate.Id))

	return nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) AfterDelete(tx *gorm.DB) (err error) {
	// event.AsyncFire(Event{}.PaymentDeleted(emailQueueEmailTemplate.AccountId, emailQueueEmailTemplate.Id))
	return nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) AfterFind() (err error) {
	return nil
}

// ############# Entity interface #############
func (emailQueueEmailTemplate EmailQueueEmailTemplate) GetId() uint { return emailQueueEmailTemplate.Id }
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) setId(id uint) { emailQueueEmailTemplate.Id = id }
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) setPublicId(publicId uint) { }
func (emailQueueEmailTemplate EmailQueueEmailTemplate) GetAccountId() uint { return emailQueueEmailTemplate.AccountId }
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) setAccountId(id uint) { emailQueueEmailTemplate.AccountId = id }
func (EmailQueueEmailTemplate) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (emailQueueEmailTemplate EmailQueueEmailTemplate) create() (Entity, error)  {
	
	wb := emailQueueEmailTemplate
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

func (EmailQueueEmailTemplate) get(id uint) (Entity, error) {

	var emailQueueEmailTemplate EmailQueueEmailTemplate

	err := emailQueueEmailTemplate.GetPreloadDb(false,false, true).First(&emailQueueEmailTemplate, id).Error
	if err != nil {
		return nil, err
	}
	return &emailQueueEmailTemplate, nil
}
func (EmailQueueEmailTemplate) getByExternalId(externalId string) (*EmailQueueEmailTemplate, error) {
	emailQueueEmailTemplate := EmailQueueEmailTemplate{}

	err := emailQueueEmailTemplate.GetPreloadDb(false,false,true).First(&emailQueueEmailTemplate, "external_id = ?", externalId).Error
	if err != nil {
		return nil, err
	}
	return &emailQueueEmailTemplate, nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) load() error {
	if emailQueueEmailTemplate.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailQueueEmailTemplate - не указан  Id"}
	}

	err := emailQueueEmailTemplate.GetPreloadDb(false,false, true).First(emailQueueEmailTemplate,emailQueueEmailTemplate.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailQueueEmailTemplate *EmailQueueEmailTemplate) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (EmailQueueEmailTemplate) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return EmailQueueEmailTemplate{}.getPaginationList(accountId, 0, 25, sortBy, "")
}

func (EmailQueueEmailTemplate) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	emailQueueEmailTemplates := make([]EmailQueueEmailTemplate,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&EmailQueueEmailTemplate{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueueEmailTemplates, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailQueueEmailTemplate{}).GetPreloadDb(true,false,true).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&EmailQueueEmailTemplate{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueueEmailTemplates).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailQueueEmailTemplate{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailQueueEmailTemplates))
	for i,_ := range emailQueueEmailTemplates {
		entities[i] = &emailQueueEmailTemplates[i]
	}

	return entities, total, nil
}

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) update(input map[string]interface{}) error {
	return emailQueueEmailTemplate.GetPreloadDb(false,false,false).Where("if = ?", emailQueueEmailTemplate.Id).Omit("id", "account_id").Updates(input).Error
}

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) delete () error {

	return emailQueueEmailTemplate.GetPreloadDb(true,false,false).Where("id = ?", emailQueueEmailTemplate.Id).Delete(emailQueueEmailTemplate).Error
}
// ######### END CRUD Functions ############

func (emailQueueEmailTemplate *EmailQueueEmailTemplate) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&emailQueueEmailTemplate)
	} else {
		_db = _db.Model(&EmailQueueEmailTemplate{})
	}

	if preload {
		// return _db.Preload("PaymentAmount")
		return _db
	} else {
		return _db
	}
}

