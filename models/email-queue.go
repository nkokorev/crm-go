package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type EmailQueue struct {

	Id     		uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Имя очереди (Label)
	Name	string	`json:"name" gorm:"type:varchar(128);not null;"` // Welcome, Onboarding, ...

	// В работе серия или нет (== нужно ли ее обходить воркером)
	Status 	bool 	`json:"status" gorm:"type:bool;default:false;"`

	// EmailQueueEmailTemplate	[]EmailQueueEmailTemplate	`json:"-"`

	// EmailQueueWorkflow []EmailQueueWorkflow `json:"emailQueueWorkflow" gorm:"preload"`
	// EmailQueueWorkflowQuantity uint `json:"emailQueueWorkflowQuantity" gorm:"preload"`

	// Сколько в очереди сейчас задач (выборка по EmailQueueWorkflow) = сколько подписчиков еще проходят, в процессе
	Queue uint `json:"_queue" gorm:"-"`

	// Из скольких активных писем состоит цепочка      activeEmailTemplates
	ActiveEmailTemplates uint `json:"_activeEmailTemplates" gorm:"-"`

	// Сколько прошло через нее. На это число навешивается статистика открытий / отписок / кликов
	Recipients uint `json:"_recipients" gorm:"-"` // << число участников в серии
	EmailsSent uint `json:"_emailsSent" gorm:"-"` // << всего успешно отправлено писем
	OpenRate float64 `json:"_openRate" gorm:"-"`
	UnsubscribeRate float64 `json:"_unsubscribeRate" gorm:"-"`
		

	// Сколько пользователей завершило серию
	// Subscribers uint `json:"emailQueueWorkflowQuantity" gorm:"preload"`

	// Внутреннее время
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (EmailQueue) PgSqlCreate() {
	db.CreateTable(&EmailQueue{})
	db.Model(&EmailQueue{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&EmailQueue{}).AddForeignKey("amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
	// db.Model(&EmailQueue{}).AddForeignKey("income_amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
	// db.Model(&EmailQueue{}).AddForeignKey("refunded_amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
}
func (emailQueue *EmailQueue) BeforeCreate(scope *gorm.Scope) error {
	emailQueue.Id = 0

	// PublicId
	lastIdx := uint(0)
	var ord Order

	err := db.Where("account_id = ?", emailQueue.AccountId).Select("public_id").Last(&ord).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err}
	if err == gorm.ErrRecordNotFound {
		lastIdx = 0
	} else {
		lastIdx = ord.PublicId
	}
	emailQueue.PublicId = lastIdx + 1

	return nil
}

func (emailQueue *EmailQueue) AfterCreate(scope *gorm.Scope) (error) {
	// event.AsyncFire(Event{}.PaymentCreated(emailQueue.AccountId, emailQueue.Id))
	return nil
}
func (emailQueue *EmailQueue) AfterUpdate(tx *gorm.DB) (err error) {

	// event.AsyncFire(Event{}.PaymentUpdated(emailQueue.AccountId, emailQueue.Id))

	return nil
}
func (emailQueue *EmailQueue) AfterDelete(tx *gorm.DB) (err error) {
	// event.AsyncFire(Event{}.PaymentDeleted(emailQueue.AccountId, emailQueue.Id))
	return nil
}
func (emailQueue *EmailQueue) AfterFind() (err error) {

	// Рассчитываем сколько активных писем в серии
	countTemplates := uint(0)
	err = db.Model(&EmailQueueEmailTemplate{}).Where("account_id = ? AND email_queue_id = ? AND status = 'true'", emailQueue.AccountId, emailQueue.Id).Count(&countTemplates).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	if err == gorm.ErrRecordNotFound {countTemplates = 0} else { emailQueue.ActiveEmailTemplates = countTemplates}

	// Рассчитываем сколько пользователей сейчас в очереди
	countQueue := uint(0)
	err = db.Model(&EmailQueueWorkflow{}).Where("account_id = ? AND email_queue_id = ?", emailQueue.AccountId, emailQueue.Id).Count(&countQueue).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	if err == gorm.ErrRecordNotFound {countQueue = 0} else { emailQueue.Queue = countQueue}


	stat := struct {
		Recipients uint  	// << Успешных отправок (succeed = true)
		Completed uint   	// << Завершило серию (completed = true)
		Opens uint    		// (opens >=1)
		Unsubscribed uint 	// (unsubscribed = true)
	}{}
	if err = db.Raw("SELECT   \n       COUNT(CASE WHEN succeed = true THEN 1 END) AS recipients,  \n       COUNT(CASE WHEN completed = true THEN 1 END) AS completed,  \n       COUNT(CASE WHEN opens >=1 THEN 1 END) AS opens,   \n       COUNT(CASE WHEN unsubscribed = true THEN 1 END) AS unsubscribed \nFROM email_queue_workflow_histories \nWHERE account_id = ? AND email_queue_id = ?;", emailQueue.AccountId, emailQueue.Id).
		Scan(&stat).Error; err != nil {
			return err
	}

	emailQueue.Recipients = stat.Completed
	emailQueue.EmailsSent = stat.Recipients
	emailQueue.OpenRate = (float64(stat.Opens) / float64(stat.Recipients))*100
	emailQueue.UnsubscribeRate = (float64(stat.Unsubscribed) / float64(stat.Recipients))*100



	/*// |Дорогой запрос| Сколько прошло подписчиков
	countRecipients := uint(0)
	// err = db.Model(&EmailQueueWorkflowHistory{}).Where("account_id = ? AND email_queue_id = ? AND completed ='true'", emailQueue.AccountId, emailQueue.Id).Count(&countRecipients).Error;
	err = db.Model(&EmailQueueWorkflowHistory{}).Where("account_id = ? AND email_queue_id = ? AND completed = 'true'", emailQueue.AccountId, emailQueue.Id).
		Select("count(id)").Count(&countRecipients).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	if err == gorm.ErrRecordNotFound {countRecipients = 0} else { emailQueue.Recipients = countRecipients}

	// |Дорогой запрос| Сколько прошло подписчиков
	countOpens := uint(0)
	// err = db.Model(&EmailQueueWorkflowHistory{}).Where("account_id = ? AND email_queue_id = ? AND completed ='true'", emailQueue.AccountId, emailQueue.Id).Count(&countRecipients).Error;
	err = db.Model(&EmailQueueWorkflowHistory{}).Where("account_id = ? AND email_queue_id = ? AND opens >=1", emailQueue.AccountId, emailQueue.Id).
		Select("count(id)").Count(&countOpens).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	if err == gorm.ErrRecordNotFound {countOpens = 0} else { emailQueue.OpenRate = (float64(countOpens) / float64(countRecipients))*100 }

	// |Дорогой запрос| Каков % отписок
	countUnsubs := uint(0)
	// err = db.Model(&EmailQueueWorkflowHistory{}).Where("account_id = ? AND email_queue_id = ? AND completed ='true'", emailQueue.AccountId, emailQueue.Id).Count(&countRecipients).Error;
	err = db.Model(&EmailQueueWorkflowHistory{}).Where("account_id = ? AND email_queue_id = ? AND unsubscribed = 'true'", emailQueue.AccountId, emailQueue.Id).
		Select("count(id)").Count(&countUnsubs).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	if err == gorm.ErrRecordNotFound {countUnsubs = 0} else { emailQueue.UnsubscribeRate = (float64(countUnsubs) / float64(countRecipients))*100 }*/
	

	return nil
}

// ############# Entity interface #############
func (emailQueue EmailQueue) GetId() uint { return emailQueue.Id }
func (emailQueue *EmailQueue) setId(id uint) { emailQueue.Id = id }
func (emailQueue *EmailQueue) setPublicId(publicId uint) { emailQueue.PublicId = publicId }
func (emailQueue EmailQueue) GetAccountId() uint { return emailQueue.AccountId }
func (emailQueue *EmailQueue) setAccountId(id uint) { emailQueue.AccountId = id }
func (EmailQueue) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (emailQueue EmailQueue) create() (Entity, error)  {
	
	wb := emailQueue
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

func (EmailQueue) get(id uint) (Entity, error) {

	var emailQueue EmailQueue

	err := emailQueue.GetPreloadDb(false,false, true).First(&emailQueue, id).Error
	if err != nil {
		return nil, err
	}
	return &emailQueue, nil
}
func (EmailQueue) getByExternalId(externalId string) (*EmailQueue, error) {
	emailQueue := EmailQueue{}

	err := emailQueue.GetPreloadDb(false,false,true).First(&emailQueue, "external_id = ?", externalId).Error
	if err != nil {
		return nil, err
	}
	return &emailQueue, nil
}
func (emailQueue *EmailQueue) load() error {
	if emailQueue.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailQueue - не указан  Id"}
	}

	err := emailQueue.GetPreloadDb(false,false, true).First(emailQueue,emailQueue.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailQueue *EmailQueue) loadByPublicId() error {
	
	if emailQueue.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailQueue - не указан  Id"}
	}

	if err := emailQueue.GetPreloadDb(false,false, true).First(emailQueue, "public_id = ?", emailQueue.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (EmailQueue) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return EmailQueue{}.getPaginationList(accountId, 0, 25, sortBy, "")
}

func (EmailQueue) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	emailQueues := make([]EmailQueue,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&EmailQueue{}).GetPreloadDb(true,false,true).
			Preload("EmailQueueWorkflow", func(db *gorm.DB) *gorm.DB {
				return db.Select([]string{"id"})
			}).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueues, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailQueue{}).GetPreloadDb(true,false,true).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&EmailQueue{}).GetPreloadDb(false,false,true).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueues).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			fmt.Println(err)
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailQueue{}).GetPreloadDb(false,false,true).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailQueues))
	for i,_ := range emailQueues {
		entities[i] = &emailQueues[i]
	}

	return entities, total, nil
}

func (emailQueue *EmailQueue) update(input map[string]interface{}) error {
	return emailQueue.GetPreloadDb(false,false,false).Where("if = ?", emailQueue.Id).Omit("id", "account_id").Updates(input).Error
}

func (emailQueue *EmailQueue) delete () error {

	return emailQueue.GetPreloadDb(true,false,false).Where("id = ?", emailQueue.Id).Delete(emailQueue).Error
}
// ######### END CRUD Functions ############

func (emailQueue *EmailQueue) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&emailQueue)
	} else {
		_db = _db.Model(&EmailQueue{})
	}

	if preload {
		// return _db.Preload("EmailTemplates")
		return _db
	} else {
		return _db
	}
}

