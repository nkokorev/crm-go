package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"net"
	"time"
)

// История отправок писем в автоматических серия. История хранится какое-то время.
// Храним: факт отправки: чего, кому, когда, число попыток, статус (успех/нет) - для быстрой выборки, открытия / отписки / ip-адрес пользователя (?).
type MTAHistory struct {

	Id     		uint   	`json:"id" gorm:"primary_key"` // очень большой индекс может быть
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Id пользователя, которому было отправлено письмо серия.
	UserId 	uint `json:"userId" gorm:"type:int;not null;default:1;"`
	User	User `json:"user"`

	// email_queues, email_campaigns, email_notifications
	OwnerType	string	`json:"ownerType" gorm:"varchar(32);default:'email_queues';not null;"` // << тип события: кампания, серия, уведомление
	OwnerId		uint	`json:"ownerId" gorm:"type:smallint;default:1;not null;"` // ID типа события: какая серия, компания или уведомление

	// Queues: номер шага в очереди, который был совершен. Для статистики серий писем.
	QueueStepId	uint	`json:"queueStepId" gorm:"type:smallint;default:null;"`

	// Queues: последний ли шаг или промежуточный шаг в цепочке. По нему выборка завершивших серию писем за указанный период времени.
	QueueCompleted 	bool 	`json:"queueCompleted" gorm:"type:bool;default:false;"`

	// ID конкретной связи <Queue>&<EmailTemplate>. Для сбора статистики по конкретному шаблону.
	// EmailQueueEmailTemplateId		uint	`json:"emailQueueEmailTemplateId" gorm:"type:smallint;default:1;not null;"`

	// К какой серии писем относится задача
	// EmailQueueId	uint	`json:"emailQueueId" gorm:"type:int;index;not null;"` // index, т.к. выборка будет идти по этой колонке
	// EmailQueue		EmailQueue `json:"emailQueue"`


	// Какой конкретно шаблон был отправлен. Может пригодиться для истории, кто что кому отправлял.
	EmailTemplateId	uint	`json:"emailTemplateId" gorm:"type:int;index;not null;"` // << index для выборки по конкретному письму

	// Была ли успешна ли отправка. По этому показателю делаем выборку для кампаний и статистики уведомлений, серий писем.
	Succeed 	bool 	`json:"succeed" gorm:"type:bool;default:false;"`

	// С какой попытки было отправлено письмо. По этой колонке можно понять качество базы/рассылок.
	NumberOfAttempts uint `json:"numberOfAttempts" gorm:"type:smallint;"`

	// ####### Статистика #######

	// Статистика открытий
	Opens 		uint 		`json:"opens" gorm:"type:smallint;"`
	OpenedAt 	*time.Time  	`json:"openedAt"` // << время 1-го открытия

	// Отписался ли человек. По этому полю будет выборка (для сбора статистики)
	Unsubscribed 	bool 	`json:"unsubscribed" gorm:"type:bool;default:false;"`
	UnsubscribedAt 	*time.Time  `json:"unsubscribedAt"` // << время отписки
	UnsubscribedReason	string `json:"unsubscribedReason" gorm:"default:null"`

	// Ip адрес с которого человек открыл письмо. Может быть полезно для определения GeoLocation.
	NetIp	*net.IP `json:"ipAddr" gorm:"type:cidr;"`

	// Дата+время создания
	CreatedAt time.Time  `json:"createdAt"`
}

func (MTAHistory) PgSqlCreate() {
	db.CreateTable(&MTAHistory{})
	db.Model(&MTAHistory{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&MTAHistory{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	
	// db.Model(&MTAHistory{}).AddForeignKey("email_queue_id", "email_queues(id)", "CASCADE", "CASCADE")

	// todo: проработать модель удаления шаблона или связи
	db.Model(&MTAHistory{}).AddForeignKey("email_template_id", "email_templates(id)", "CASCADE", "CASCADE")

}
func (emailQueueWorkflowHistory *MTAHistory) BeforeCreate(scope *gorm.Scope) error {
	emailQueueWorkflowHistory.Id = 0

	return nil
}

func (emailQueueWorkflowHistory *MTAHistory) AfterCreate(scope *gorm.Scope) (error) {
	// event.AsyncFire(Event{}.PaymentCreated(emailQueueWorkflowHistory.AccountId, emailQueueWorkflowHistory.Id))
	return nil
}
func (emailQueueWorkflowHistory *MTAHistory) AfterUpdate(tx *gorm.DB) (err error) {

	// event.AsyncFire(Event{}.PaymentUpdated(emailQueueWorkflowHistory.AccountId, emailQueueWorkflowHistory.Id))

	return nil
}
func (emailQueueWorkflowHistory *MTAHistory) AfterDelete(tx *gorm.DB) (err error) {
	// event.AsyncFire(Event{}.PaymentDeleted(emailQueueWorkflowHistory.AccountId, emailQueueWorkflowHistory.Id))
	return nil
}
func (emailQueueWorkflowHistory *MTAHistory) AfterFind() (err error) {
	return nil
}

// ############# Entity interface #############
func (emailQueueWorkflowHistory MTAHistory) GetId() uint { return emailQueueWorkflowHistory.Id }
func (emailQueueWorkflowHistory *MTAHistory) setId(id uint) { emailQueueWorkflowHistory.Id = id }
func (emailQueueWorkflowHistory *MTAHistory) setPublicId(publicId uint) { }
func (emailQueueWorkflowHistory MTAHistory) GetAccountId() uint { return emailQueueWorkflowHistory.AccountId }
func (emailQueueWorkflowHistory *MTAHistory) setAccountId(id uint) { emailQueueWorkflowHistory.AccountId = id }
func (MTAHistory) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (emailQueueWorkflowHistory MTAHistory) create() (Entity, error)  {
	
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

func (MTAHistory) get(id uint) (Entity, error) {

	var emailQueueWorkflowHistory MTAHistory

	err := emailQueueWorkflowHistory.GetPreloadDb(false,false, true).First(&emailQueueWorkflowHistory, id).Error
	if err != nil {
		return nil, err
	}
	return &emailQueueWorkflowHistory, nil
}

func (emailQueueWorkflowHistory *MTAHistory) load() error {
	if emailQueueWorkflowHistory.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить MTAHistory - не указан  Id"}
	}

	err := emailQueueWorkflowHistory.GetPreloadDb(false,false, true).First(emailQueueWorkflowHistory,emailQueueWorkflowHistory.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailQueueWorkflowHistory *MTAHistory) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (MTAHistory) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return MTAHistory{}.getPaginationList(accountId, 0, 25, sortBy, "",nil)
}

func (MTAHistory) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	emailQueueHistories := make([]MTAHistory,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&MTAHistory{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueueHistories, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&MTAHistory{}).GetPreloadDb(true,false,true).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&MTAHistory{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueueHistories).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&MTAHistory{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailQueueHistories))
	for i,_ := range emailQueueHistories {
		entities[i] = &emailQueueHistories[i]
	}

	return entities, total, nil
}

func (emailQueueWorkflowHistory *MTAHistory) update(input map[string]interface{}) error {
	return emailQueueWorkflowHistory.GetPreloadDb(false,false,false).Where("id = ?", emailQueueWorkflowHistory.Id).Omit("id", "account_id").Updates(input).Error
}

func (emailQueueWorkflowHistory *MTAHistory) delete () error {

	return emailQueueWorkflowHistory.GetPreloadDb(true,false,false).Where("id = ?", emailQueueWorkflowHistory.Id).Delete(emailQueueWorkflowHistory).Error
}
// ######### END CRUD Functions ############

func (emailQueueWorkflowHistory *MTAHistory) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&emailQueueWorkflowHistory)
	} else {
		_db = _db.Model(&MTAHistory{})
	}

	if preload {
		// return _db.Preload("PaymentAmount")
		return _db
	} else {
		return _db
	}
}
