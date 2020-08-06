package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// активная база с реальными задачами по отправке
type EmailQueueWorkflow struct {

	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"` // << нужен ли? т.к. поиск будет по сериям писем скорее всего...

	// К какой серии писем относится задача
	EmailQueueId	uint	`json:"emailQueueId" gorm:"type:int;not null;"` // может быть надо добавить index, т.к. по этой колонке будет поиск
	EmailQueue		EmailQueue `json:"emailQueue"`

	// Номер необходимого шага в серии EmailQueue. Шаг определяется ситуационно в момент Expected Time. Если шага нет - серия завершается за пользователя.
	// После выполнения - № шага увеличивается на 1
	ExpectedStepId	uint	`json:"expectedStepId" gorm:"type:int;not null;"`

	// Запланированное время отправки (time for next step)
	ExpectedTimeStart 	time.Time `json:"expectedTimeStart"`

	// Id пользователя, которому отправляется серия. Обязательный параметр.
	UserId 	uint `json:"userId" gorm:"type:int;not null;"`
	User	User `json:"user"`

	// Число попыток отправки. Если почему-то не удалось отправить письмо есть возможность перенести отправку и повторить попытку. При 2-3х попытках, завершается неудачей.
	NumberOfAttempts uint `json:"numberOfAttempts" gorm:"type:smallint;default:0;"`

	// Когда была последняя попытка отправки (обычно = CreatedAt time)
	LastTriedAt time.Time  `json:"lastTriedAt"`

	CreatedAt time.Time  `json:"createdAt"`
	// UpdatedAt time.Time  `json:"updatedAt"` << не очень нужно знать, когда произошло обновление
}

func (EmailQueueWorkflow) PgSqlCreate() {
	db.CreateTable(&EmailQueueWorkflow{})
	db.Model(&EmailQueueWorkflow{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&EmailQueueWorkflow{}).AddForeignKey("email_queue_id", "email_queues(id)", "CASCADE", "CASCADE")
	db.Model(&EmailQueueWorkflow{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
}

func (emailQueueWorkflow *EmailQueueWorkflow) BeforeCreate(scope *gorm.Scope) error {
	emailQueueWorkflow.Id = 0
	return nil
}

func (emailQueueWorkflow *EmailQueueWorkflow) AfterCreate(scope *gorm.Scope) (error) {
	// event.AsyncFire(Event{}.PaymentCreated(emailQueueWorkflow.AccountId, emailQueueWorkflow.Id))
	return nil
}
func (emailQueueWorkflow *EmailQueueWorkflow) AfterUpdate(tx *gorm.DB) (err error) {

	// event.AsyncFire(Event{}.PaymentUpdated(emailQueueWorkflow.AccountId, emailQueueWorkflow.Id))

	return nil
}
func (emailQueueWorkflow *EmailQueueWorkflow) AfterDelete(tx *gorm.DB) (err error) {
	// event.AsyncFire(Event{}.PaymentDeleted(emailQueueWorkflow.AccountId, emailQueueWorkflow.Id))
	return nil
}
func (emailQueueWorkflow *EmailQueueWorkflow) AfterFind() (err error) {
	return nil
}

// ############# Entity interface #############
func (emailQueueWorkflow EmailQueueWorkflow) GetId() uint { return emailQueueWorkflow.Id }
func (emailQueueWorkflow *EmailQueueWorkflow) setId(id uint) { emailQueueWorkflow.Id = id }
func (emailQueueWorkflow *EmailQueueWorkflow) setPublicId(publicId uint) { }
func (emailQueueWorkflow EmailQueueWorkflow) GetAccountId() uint { return emailQueueWorkflow.AccountId }
func (emailQueueWorkflow *EmailQueueWorkflow) setAccountId(id uint) { emailQueueWorkflow.AccountId = id }
func (EmailQueueWorkflow) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (emailQueueWorkflow EmailQueueWorkflow) create() (Entity, error)  {
	
	wb := emailQueueWorkflow
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

func (EmailQueueWorkflow) get(id uint) (Entity, error) {

	var emailQueueWorkflow EmailQueueWorkflow

	err := emailQueueWorkflow.GetPreloadDb(false,false, true).First(&emailQueueWorkflow, id).Error
	if err != nil {
		return nil, err
	}
	return &emailQueueWorkflow, nil
}
func (EmailQueueWorkflow) getByExternalId(externalId string) (*EmailQueueWorkflow, error) {
	emailQueueWorkflow := EmailQueueWorkflow{}

	err := emailQueueWorkflow.GetPreloadDb(false,false,true).First(&emailQueueWorkflow, "external_id = ?", externalId).Error
	if err != nil {
		return nil, err
	}
	return &emailQueueWorkflow, nil
}
func (emailQueueWorkflow *EmailQueueWorkflow) load() error {
	if emailQueueWorkflow.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailQueueWorkflow - не указан  Id"}
	}

	err := emailQueueWorkflow.GetPreloadDb(false,false, true).First(emailQueueWorkflow,emailQueueWorkflow.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailQueueWorkflow *EmailQueueWorkflow) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (EmailQueueWorkflow) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return EmailQueueWorkflow{}.getPaginationList(accountId, 0, 25, sortBy, "")
}

func (EmailQueueWorkflow) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	webHooks := make([]EmailQueueWorkflow,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&EmailQueueWorkflow{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailQueueWorkflow{}).GetPreloadDb(true,false,true).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&EmailQueueWorkflow{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailQueueWorkflow{}).Where("account_id = ?", accountId).Count(&total).Error
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

func (emailQueueWorkflow *EmailQueueWorkflow) update(input map[string]interface{}) error {
	return emailQueueWorkflow.GetPreloadDb(false,false,false).Where("id = ?", emailQueueWorkflow.Id).Omit("id", "account_id").Updates(input).Error
}

func (emailQueueWorkflow *EmailQueueWorkflow) delete () error {

	return emailQueueWorkflow.GetPreloadDb(true,false,false).Where("id = ?", emailQueueWorkflow.Id).Delete(emailQueueWorkflow).Error
}
// ######### END CRUD Functions ############

func (emailQueueWorkflow *EmailQueueWorkflow) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&emailQueueWorkflow)
	} else {
		_db = _db.Model(&EmailQueueWorkflow{})
	}

	if preload {
		// return _db.Preload("PaymentAmount")
		return _db
	} else {
		return _db
	}
}

