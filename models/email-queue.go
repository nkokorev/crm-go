package models

import (
	"database/sql"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type EmailQueue struct {

	Id     		uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Имя очереди (Label)
	Name		string	`json:"name" gorm:"type:varchar(128);not null;"` // Welcome, Onboarding, ...

	// В работе серия или нет (== нужно ли ее обходить воркером)
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:false;"`

	// Сколько в очереди сейчас задач (выборка по MTAWorkflow) = сколько подписчиков еще проходят, в процессе
	Queue 		uint `json:"_queue" gorm:"-"`

	// Из скольких активных писем состоит цепочка      activeEmailTemplates
	ActiveEmailTemplates uint `json:"_activeEmailTemplates" gorm:"-"`

	// Сколько прошло через нее. На это число навешивается статистика открытий / отписок / кликов
	Recipients uint `json:"_recipients" gorm:"-"` // <<< всего успешно отправлено писем
	Completed uint `json:"_completed" gorm:"-"` /// << число завершивших серию
	OpenRate 		float64 `json:"_openRate" gorm:"-"`
	UnsubscribeRate float64 `json:"_unsubscribeRate" gorm:"-"`

	// Внутреннее время
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (EmailQueue) PgSqlCreate() {
	db.CreateTable(&EmailQueue{})
	db.Model(&EmailQueue{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (emailQueue *EmailQueue) BeforeCreate(scope *gorm.Scope) error {
	emailQueue.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&EmailQueue{}).Where("account_id = ?",  emailQueue.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	emailQueue.PublicId = 1 + uint(lastIdx.Int64)

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
	err = db.Model(&EmailQueueEmailTemplate{}).Where("account_id = ? AND email_queue_id = ? AND enabled = 'true'", emailQueue.AccountId, emailQueue.Id).Count(&countTemplates).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	if err == gorm.ErrRecordNotFound {countTemplates = 0} else { emailQueue.ActiveEmailTemplates = countTemplates}

	// Рассчитываем сколько пользователей сейчас в очереди
	inQueue := uint(0)
	err = db.Model(&MTAWorkflow{}).Where("account_id = ? AND owner_id = ? AND owner_type = ?", emailQueue.AccountId, emailQueue.Id, EmailSenderQueue).Count(&inQueue).Error
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	if err == gorm.ErrRecordNotFound {emailQueue.Queue = 0} else { emailQueue.Queue = inQueue}

	// todo: можно добавить % доставки
	stat := struct {
		Recipients uint  	// << Всего отправок
		Completed uint   	// << Завершило серию (completed = true)
		Opens uint    		// (opens >=1)
		Unsubscribed uint 	// (unsubscribed = true)
	}{0,0,0,0}
	if err = db.Raw("SELECT   \n--        COUNT(CASE WHEN succeed = true THEN 1 END) AS recipients, -- фактически отправленных писем (успешно)\n       COUNT(*) AS recipients, -- успешно отправленных\n       COUNT(CASE WHEN queue_completed = true THEN 1 END) AS completed, -- завершивших серию \n       COUNT(CASE WHEN opens >=1 THEN 1 END) AS opens, -- открытий среди успешно отправленных   \n       COUNT(CASE WHEN unsubscribed = true THEN 1 END) AS unsubscribed \nFROM mta_histories \nWHERE account_id = ? AND owner_id = ? AND owner_type = 'email_queues';", emailQueue.AccountId, emailQueue.Id).
		Scan(&stat).Error; err != nil {
			return err
	}

	emailQueue.Recipients = stat.Recipients // Сколько всего реально было отправлено писем.
	emailQueue.Completed = stat.Completed //  queue_completed = true
	
	if stat.Opens > 0 && stat.Recipients > 0{
		emailQueue.OpenRate = (float64(stat.Opens) / float64(stat.Recipients))*100
	} else {
		emailQueue.OpenRate = 0
	}
	if stat.Unsubscribed > 0 && stat.Recipients > 0{
		emailQueue.UnsubscribeRate = (float64(stat.Unsubscribed) / float64(stat.Recipients))*100
	} else {
		emailQueue.UnsubscribeRate = 0
	}
	
	return nil
}

// ############# Entity interface #############
func (emailQueue EmailQueue) GetId() uint { return emailQueue.Id }
func (emailQueue *EmailQueue) setId(id uint) { emailQueue.Id = id }
func (emailQueue *EmailQueue) setPublicId(publicId uint) { emailQueue.PublicId = publicId }
func (emailQueue EmailQueue) GetAccountId() uint { return emailQueue.AccountId }
func (emailQueue *EmailQueue) setAccountId(id uint) { emailQueue.AccountId = id }
func (EmailQueue) SystemEntity() bool { return false }
func (EmailQueue) GetType() string { return "email_queues" }
func (emailQueue EmailQueue) IsEnabled() bool { return emailQueue.Enabled }
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
	if err := emailQueue.GetPreloadDb(false,false, true).First(emailQueue, "account_id = ? AND public_id = ?", emailQueue.AccountId, emailQueue.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (EmailQueue) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return EmailQueue{}.getPaginationList(accountId, 0, 25, sortBy, "", nil)
}

func (EmailQueue) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	emailQueues := make([]EmailQueue,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&EmailQueue{}).GetPreloadDb(true,false,true).
			/*Preload("MTAWorkflow", func(db *gorm.DB) *gorm.DB {
				return db.Select([]string{"id"})
			}).*/
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueues, "name ILIKE ?", search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailQueue{}).GetPreloadDb(true,false,true).
			Where("account_id = ? AND name ILIKE ?", accountId, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&EmailQueue{}).GetPreloadDb(false,false,true).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueues).Error
		if err != nil && err != gorm.ErrRecordNotFound{
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
	for i := range emailQueues {
		entities[i] = &emailQueues[i]
	}

	return entities, total, nil
}

func (emailQueue *EmailQueue) update(input map[string]interface{}) error {
	input = utils.FixInputHiddenVars(input)
	err := emailQueue.GetPreloadDb(true,false,false).Where("id = ?", emailQueue.Id).
		Omit("id", "account_id", "created_at").Update(input).Error
	if err != nil {	return err	}
	if err = emailQueue.GetPreloadDb(false,true,true).First(emailQueue).Error; err != nil {
		return err
	}

	return nil
}

type MassUpdateEmailQueueTemplate struct {
	Id uint `json:"id"`
	Order uint `json:"order"`
}

func (emailQueue *EmailQueue) UpdateOrderEmailTemplates(input []MassUpdateEmailQueueTemplate) error {
	for _,v := range (input) {

		if err := (&EmailQueueEmailTemplate{Id: v.Id }).update(map[string]interface{}{"order":v.Order}); err != nil {
			return err
		}
	}
	return nil
	// return emailQueue.GetPreloadDb(false,false,false).Where("id = ?", emailQueue.Id).Omit("id", "account_id").Updates(input).Error
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

//////// ###### WORKER function ########## //////////

// Получает шаблон для stepId шага
func (emailQueue EmailQueue) GetStepByOrder(order uint) (*EmailQueueEmailTemplate, error) {
	var eqet EmailQueueEmailTemplate
	 if err := db.Model(&eqet).Where("email_queue_id = ? AND email_queue_email_templates.order = ?", emailQueue.Id, order).First(&eqet).Error; err != nil {
	 	return nil, err
	 }

	 return &eqet, nil
}
func (emailQueue EmailQueue) GetFirstStep() (*EmailQueueEmailTemplate, error) {
	// var eqet EmailQueueEmailTemplate
	var order = uint(0)
	err := db.Model(&EmailQueueEmailTemplate{}).Where("email_queue_id = ? AND enabled = 'true'", emailQueue.Id).
		Select("min(email_queue_email_templates.order)").Row().Scan(&order)
	if err != nil {
		return nil, utils.Error{Message: "Нет доступных писем для отправления"}
	}

	eqet, err := emailQueue.GetStepByOrder(order)
	if err != nil {
		return nil, err
	}

	return eqet, nil
}

// Возвращает ближайший шаг (может быть равен order) или ошибку
func (emailQueue EmailQueue) GetNearbyActiveStep(order uint) (*EmailQueueEmailTemplate, error) {

	// var eqet EmailQueueEmailTemplate
	var _order = uint(0)
	err := db.Model(&EmailQueueEmailTemplate{}).Where("email_queue_id = ? AND enabled = 'true' AND email_queue_email_templates.order >= ?", emailQueue.Id, order).
		Select("min(email_queue_email_templates.order)").Row().Scan(&_order)
	if err != nil {
		return nil, utils.Error{Message: "Нет доступных писем для отправления"}
	}

	step, err := emailQueue.GetStepByOrder(_order)
	if err != nil {
		return nil, err
	}

	return step, nil
}

func (emailQueue EmailQueue) GetNextActiveStep(order uint) (*EmailQueueEmailTemplate, error) {
	return emailQueue.GetNearbyActiveStep(order+1)
}

func (emailQueue EmailQueue) AppendUser(userId uint) error {
	// adding user to worker

	// 1. Check some
	if emailQueue.Id < 1 && userId < 1 {
		return errors.New("Ошибка добавление пользователя в серию писем")
	}

	// 2. Get Step
	step, err := emailQueue.GetFirstStep()
	if err != nil {
		return err
	}
	

	// todo: проверка на запуск письма в серии.
	// ...
	
	// 2. Add user to MTAWorkflow
	mtaWorkflow := MTAWorkflow{
		AccountId: emailQueue.AccountId,
		OwnerId: emailQueue.Id,
		OwnerType: EmailSenderQueue,
		QueueExpectedStepId: step.Order,
		ExpectedTimeStart: time.Now().UTC().Add(step.DelayTime),
		UserId: userId, 
		NumberOfAttempts: 0, // << пока так
	}

	if _, err := mtaWorkflow.create(); err != nil {
		return err
	}

	return nil
}

func (emailQueue EmailQueue) GetEmailTemplateByStep(order uint) (*EmailTemplate, error) {
	step, err := emailQueue.GetStepByOrder(order)
	if err != nil {return nil, err}

	var emailTemplate EmailTemplate
	if err = (Account{Id: emailQueue.AccountId}).LoadEntity(&emailTemplate, step.EmailTemplateId); err != nil {
		return nil, err
	}

	return &emailTemplate, nil
}
