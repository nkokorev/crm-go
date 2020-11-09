package models

import (
	"database/sql"
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type EmailQueue struct {

	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Имя очереди (Label)
	Name		string	`json:"name" gorm:"type:varchar(128);not null;"` // Welcome, Onboarding, ...

	Status 			WorkStatus 		`json:"status" gorm:"type:varchar(18);default:'pending'"`
	FailedStatus	*string 		`json:"failed_status" gorm:"type:varchar(255);"`

	// Сколько в очереди сейчас задач (выборка по MTAWorkflow) = сколько подписчиков еще проходят, в процессе
	Queue 		int64 `json:"_queue" gorm:"-"`

	// Из скольких активных писем состоит цепочка      activeEmailTemplates
	ActiveEmailTemplates int64 `json:"_active_email_templates" gorm:"-"`

	// Сколько прошло через нее. На это число навешивается статистика открытий / отписок / кликов
	Recipients 		uint 	`json:"_recipients" gorm:"-"` // <<< всего успешно отправлено писем
	Completed 		uint 	`json:"_completed" gorm:"-"` /// << число завершивших серию
	OpenRate 		float64 `json:"_open_rate" gorm:"-"`
	UnsubscribeRate float64 `json:"_unsubscribe_rate" gorm:"-"`

	EmailQueueEmailTemplates []EmailQueueEmailTemplate `json:"email_queue_email_templates"`

	// Внутреннее время
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (EmailQueue) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&EmailQueue{});err != nil {
		log.Fatal(err)
	}
	// db.Model(&EmailQueue{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE email_queues ADD CONSTRAINT email_queues_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error EmailQueue: ", err)
	}
}
func (emailQueue *EmailQueue) BeforeCreate(tx *gorm.DB) error {
	emailQueue.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&EmailQueue{}).Where("account_id = ?",  emailQueue.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	emailQueue.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}

func (emailQueue *EmailQueue) AfterCreate(tx *gorm.DB) (error) {
	// AsyncFire(*Event{}.PaymentCreated(emailQueue.AccountId, emailQueue.Id))
	return nil
}
func (emailQueue *EmailQueue) AfterUpdate(tx *gorm.DB) (err error) {

	// AsyncFire(*Event{}.PaymentUpdated(emailQueue.AccountId, emailQueue.Id))

	return nil
}
func (emailQueue *EmailQueue) AfterDelete(tx *gorm.DB) (err error) {
	// AsyncFire(*Event{}.PaymentDeleted(emailQueue.AccountId, emailQueue.Id))
	return nil
}
func (emailQueue *EmailQueue) AfterFind(tx *gorm.DB) (err error) {

	// Рассчитываем сколько активных писем в серии
	countTemplates := int64(0)
	err = db.Model(&EmailQueueEmailTemplate{}).Where("account_id = ? AND email_queue_id = ? AND enabled = 'true'", emailQueue.AccountId, emailQueue.Id).Count(&countTemplates).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	if err == gorm.ErrRecordNotFound {countTemplates = 0} else { emailQueue.ActiveEmailTemplates = countTemplates}

	// Рассчитываем сколько пользователей сейчас в очереди
	inQueue := int64(0)
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
func (emailQueue EmailQueue) IsEnabled() bool {
	// т.к. статус для обхода воркера
	if emailQueue.Status == WorkStatusPending || emailQueue.Status == WorkStatusFailed || emailQueue.Status == WorkStatusCancelled {
		return false
	}

	return true
}
func (emailQueue EmailQueue) IsActive() bool {
	return emailQueue.Status == WorkStatusActive
}
// ############# Entity interface #############

// ######### CRUD Functions ############
func (emailQueue EmailQueue) create() (Entity, error)  {
	
	wb := emailQueue
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	err := wb.GetPreloadDb(false,false, nil).First(&wb, wb.Id).Error
	if err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (EmailQueue) get(id uint, preloads []string) (Entity, error) {

	var emailQueue EmailQueue

	err := emailQueue.GetPreloadDb(false,false, preloads).First(&emailQueue, id).Error
	if err != nil {
		return nil, err
	}
	return &emailQueue, nil
}
func (emailQueue *EmailQueue) load(preloads []string) error {
	if emailQueue.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailQueue - не указан  Id"}
	}

	err := emailQueue.GetPreloadDb(false,false, preloads).First(emailQueue,emailQueue.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailQueue *EmailQueue) loadByPublicId(preloads []string) error {
	
	if emailQueue.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailQueue - не указан  Id"}
	}
	if err := emailQueue.GetPreloadDb(false,false, preloads).First(emailQueue, "account_id = ? AND public_id = ?", emailQueue.AccountId, emailQueue.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (EmailQueue) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return EmailQueue{}.getPaginationList(accountId, 0, 25, sortBy, "", nil,preload)
}
func (EmailQueue) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	emailQueues := make([]EmailQueue,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&EmailQueue{}).GetPreloadDb(false,false,preloads).
			/*Preload("MTAWorkflow", func(db *gorm.DB) *gorm.DB {
				return db.Select([]string{"id"})
			}).*/
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueues, "name ILIKE ?", search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailQueue{}).GetPreloadDb(false,false,nil).
			Where("account_id = ? AND name ILIKE ?", accountId, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&EmailQueue{}).GetPreloadDb(false,false,preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailQueues).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailQueue{}).GetPreloadDb(false,false,nil).Where("account_id = ?", accountId).Count(&total).Error
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
func (emailQueue *EmailQueue) update(input map[string]interface{}, preloads []string) error {
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}
	err := emailQueue.GetPreloadDb(true,false,nil).Where("id = ?", emailQueue.Id).
		Omit("id", "account_id", "created_at","public_id").Updates(input).Error
	if err != nil {	return err	}
	if err = emailQueue.GetPreloadDb(false,false,preloads).First(emailQueue).Error; err != nil {
		return err
	}

	return nil
}

type MassUpdateEmailQueueTemplate struct {
	Id uint `json:"id"`
	Step uint `json:"step"`
}

func (emailQueue *EmailQueue) UpdateOrderEmailTemplates(input []MassUpdateEmailQueueTemplate) error {
	for _,v := range input {

		if err := (&EmailQueueEmailTemplate{Id: v.Id }).update(map[string]interface{}{"step":v.Step},nil); err != nil {
			return err
		}
	}
	return nil
	// return emailQueue.GetPreloadDb(false,false,false).Where("id = ?", emailQueue.Id).Omit("id", "account_id").Updates(input).Error
}

func (emailQueue *EmailQueue) delete () error {

	return emailQueue.GetPreloadDb(true,false,nil).Where("id = ?", emailQueue.Id).Delete(emailQueue).Error
}
// ######### END CRUD Functions ############

func (emailQueue *EmailQueue) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(emailQueue)
	} else {
		_db = _db.Model(&EmailQueue{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

//////// ###### WORKER function ########## //////////

// Получает шаблон для stepId шага
func (emailQueue EmailQueue) GetStepById(step uint) (*EmailQueueEmailTemplate, error) {
	var eqet EmailQueueEmailTemplate
	 if err := db.Model(&eqet).Where("email_queue_id = ? AND email_queue_email_templates.step = ?", emailQueue.Id, step).First(&eqet).Error; err != nil {
	 	return nil, err
	 }

	 return &eqet, nil
}
func (emailQueue EmailQueue) GetFirstStep() (*EmailQueueEmailTemplate, error) {
	// var eqet EmailQueueEmailTemplate
	// var step = float64(0)
	var stepSql sql.NullInt64
	err := db.Model(&EmailQueueEmailTemplate{}).Where("email_queue_id = ? AND enabled = 'true'", emailQueue.Id).
		Select("min(email_queue_email_templates.step)").Row().Scan(&stepSql)
	if err != nil || !stepSql.Valid{
		return nil, utils.Error{Message: "Нет доступных писем для отправления (1)"}
	}

	eqet, err := emailQueue.GetStepById(uint(stepSql.Int64))
	if err != nil {
		return nil, err
	}

	return eqet, nil
}

// Возвращает ближайший шаг (может быть равен order) или ошибку
func (emailQueue EmailQueue) GetNearbyActiveStep(step uint) (*EmailQueueEmailTemplate, error) {

	var stepSql sql.NullInt64
	// err := db.Model(&EmailQueueEmailTemplate{}).Where("email_queue_id = ? AND enabled = 'true' AND email_queue_email_templates.step >= ?", emailQueue.Id, step).
	err := db.Model(&EmailQueueEmailTemplate{}).Where("email_queue_id = ? AND enabled = 'true' AND step >= ?", emailQueue.Id, step).
		Select("min(step)").Row().Scan(&stepSql)
	if err != nil || !stepSql.Valid {
		return nil, utils.Error{Message: "Нет доступных писем для отправления (2)"}
	}

	stepEqEt, err := emailQueue.GetStepById(uint(stepSql.Int64))
	if err != nil {
		return nil, err
	}

	return stepEqEt, nil
}

func (emailQueue EmailQueue) GetNextActiveStep(currentStep uint) (*EmailQueueEmailTemplate, error) {
	return emailQueue.GetNearbyActiveStep(currentStep+1)
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

	// todo: проверка на статус подписки пользователя
	

	// todo: проверка на запуск письма в серии.
	// ...
	
	// 2. Add user to MTAWorkflow
	mtaWorkflow := MTAWorkflow{
		AccountId: emailQueue.AccountId,
		OwnerId: emailQueue.Id,
		OwnerType: EmailSenderQueue,
		QueueExpectedStepId: &step.Step,
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
	step, err := emailQueue.GetStepById(order)
	if err != nil {return nil, err}

	if step.EmailTemplateId == nil {
		return nil, utils.Error{Message: "Отсутствует шаблон"}
	}

	var emailTemplate EmailTemplate
	if err = (Account{Id: emailQueue.AccountId}).LoadEntity(&emailTemplate, *step.EmailTemplateId,nil); err != nil {
		return nil, err
	}

	return &emailTemplate, nil
}

// Прямая функция для обновления, без проверок
func (emailQueue *EmailQueue) ChangeWorkStatus(status WorkStatus, reason... string) error {

	switch status {
	case WorkStatusPending:
		return emailQueue.SetPendingStatus()
	case WorkStatusPlanned:
		return emailQueue.SetPlannedStatus()
	case WorkStatusActive:
		return emailQueue.SetActiveStatus()
	case WorkStatusPaused:
		return emailQueue.SetPausedStatus()
	case WorkStatusFailed:
		_reason := ""
		if len(reason) > 0 {
			_reason = reason[0]
		}
		return emailQueue.SetFailedStatus(_reason)
	case WorkStatusCompleted:
		return emailQueue.SetCompletedStatus()
	case WorkStatusCancelled:
		return emailQueue.SetCancelledStatus()
	default:
		return utils.Error{Message: "Статус не опознан"}
	}

}

func (emailQueue *EmailQueue) updateWorkStatus(status WorkStatus, reason... string) error {
	_reason := ""
	if len(reason) > 0 {
		_reason = reason[0]
	}

	return emailQueue.update(map[string]interface{}{
		"status":	status,
		"failed_status": _reason,
	},nil)
}

// Функция не должна вызываться, т.к. статуса planned у EmailNotification не используется
func (emailQueue *EmailQueue) SetPendingStatus() error {

	// Возможен вызов из состояния planned: вернуть на доработку => pending
	if emailQueue.Status != WorkStatusPlanned {
		reason := "Невозможно установить статус,"
		switch emailQueue.Status {
		case WorkStatusPending:
			reason += "т.к. серия писем уже в разработке"
		case WorkStatusActive:
			reason += "т.к. серия писем в процессе работы"
		case WorkStatusPaused:
			reason += "т.к. серия писем на паузе, но в процессе работы"
		case WorkStatusFailed:
			reason += "т.к. серия писем завершено с ошибкой"
		case WorkStatusCompleted:
			reason += "т.к. серия писем уже завершено"
		case WorkStatusCancelled:
			reason += "т.к. серия писем отменено"
		}
		return utils.Error{Message: reason}
	}

	// fix: удаляем задачу в TaskScheduler
	if err := emailQueue.RemoveRunTask(); err != nil {
		return err
	}

	return emailQueue.updateWorkStatus(WorkStatusPending)
}

// Функция не должна вызываться, т.к. статус planned у EmailNotification не используется
func (emailQueue *EmailQueue) SetPlannedStatus() error {

	return utils.Error{Message: "Отложенный запуск серии писем не предусмотрен"}

	// Возможен вызов из состояния pending: запланировать кампанию => planned
	if emailQueue.Status != WorkStatusPending  {
		reason := "Невозможно запланировать серию писем,"
		switch emailQueue.Status {
		case WorkStatusPlanned:
			reason += "т.к. серию писем уже в плане"
		case WorkStatusActive:
			reason += "т.к. серию писем уже в процессе рассылки"
		case WorkStatusPaused:
			reason += "т.к. серию писем на паузе, но уже в процессе рассылки"
		case WorkStatusCompleted:
			reason += "т.к. серию писем уже завершена"
		case WorkStatusFailed:
			reason += "т.к. серию писем завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. серию писем отменена"
		}
		return utils.Error{Message: reason}
	}

	// Get Account
	account, err := GetAccount(emailQueue.AccountId)
	if err != nil {
		return utils.Error{Message: "Не удается найти аккаунт"}
	}

	// Проверяем кампанию и шаблон, чтобы не ставить в план не рабочую кампанию.
	if err := emailQueue.Validate(); err != nil { return err  }

	// На всякий случай удаляем задачу по запуску этой кампании, чтобы не было дубля
	if err := emailQueue.RemoveRunTask(); err != nil {
		return err
	}

	// Создаем новый объект task для добавлении кампании в TaskScheduler
	task := TaskScheduler {
		AccountId: emailQueue.AccountId,
		OwnerType: TaskEmailQueueRun,
		OwnerId: emailQueue.Id,
		// ExpectedTimeToStart: emailNotification.ScheduleRun.Add(-time.Minute*5),// запускаем задачу (но не уведомление) за 5 минут (= с запасом)
		IsSystem: true, // системная задача ли ?
		Status: WorkStatusPlanned,
	}

	// Создаем задачу по отправке рекламной кампании
	_, err = account.CreateEntity(&task)
	if err != nil {
		return utils.Error{Message: "Ошибка создания задания по запуску кампании"}
	}

	// Переводим в состояние "Запланирована", т.к. все проверки пройдены и можно ставить ее в планировщик
	return emailQueue.updateWorkStatus(WorkStatusPlanned)
}
func (emailQueue *EmailQueue) SetActiveStatus() error {

	// Возможен вызов из состояния planned или paused: запустить кампанию => active
	if emailQueue.Status != WorkStatusCompleted && emailQueue.Status != WorkStatusPending && emailQueue.Status != WorkStatusPlanned && emailQueue.Status != WorkStatusPaused {
		reason := "Невозможно запустить серию писем,"
		switch emailQueue.Status {
		case WorkStatusPending:
			reason += "т.к. серию писем еще в стадии разработки"
		case WorkStatusActive:
			reason += "т.к. серию писем уже в процессе работы"
		case WorkStatusCompleted:
			reason += "т.к. серию писем уже завершено"
		case WorkStatusFailed:
			reason += "т.к. серию писем завершено с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. серию писем отменено"
		}
		return utils.Error{Message: reason}
	}

	// Снова проверяем кампанию и шаблон
	if err := emailQueue.Validate(); err != nil { return err  }

	// Переводим в состояние "Активна", т.к. все проверки пройдены и можно продолжить ее выполнение
	return emailQueue.updateWorkStatus(WorkStatusActive)
}
func (emailQueue *EmailQueue) SetPausedStatus() error {

	// Возможен вызов из состояния active: приостановить кампанию => paused
	if emailQueue.Status != WorkStatusActive {
		reason := "Невозможно приостановить отправку серии писем,"
		switch emailQueue.Status {
		case WorkStatusPending:
			reason += "т.к. серия писем еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. серия писем уже в стадии планирования"
		case WorkStatusPaused:
			reason += "т.к. серия писем уже приостановлено"
		case WorkStatusCompleted:
			reason += "т.к. серия писем уже завершено"
		case WorkStatusFailed:
			reason += "т.к. серия писем завершено с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. серия писем отменено"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Приостановлена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailQueue.updateWorkStatus(WorkStatusPaused)
}
func (emailQueue *EmailQueue) SetCompletedStatus() error {

	// Возможен вызов из состояния active, paused: завершить кампанию => completed
	// Сбрасываются все задачи из очереди
	if emailQueue.Status != WorkStatusActive && emailQueue.Status != WorkStatusPaused {
		reason := "Невозможно завершить серию писем,"
		switch emailQueue.Status {
		case WorkStatusPending:
			reason += "т.к. серия писем еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. серия писем еще в стадии планирования"
		case WorkStatusCompleted:
			reason += "т.к. серия писем уже завершено"
		case WorkStatusFailed:
			reason += "т.к. серия писем завершено с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. серия писем отменено"
		}
		return utils.Error{Message: reason}
	}

	// Удаляем все задачи из WorkFlow
	err := db.Where("owner_id = ? AND owner_type = ?", emailQueue.Id, emailQueue.GetType()).Delete(&MTAWorkflow{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из MTAWorkflow: %v\b", err)
		return utils.Error{Message: "Ошибка завершения кампании - невозможно удалить остаток задач"}
	}

	// Переводим в состояние "Завершено", т.к. все проверки пройдены и можно приостановить уведомление
	return emailQueue.updateWorkStatus(WorkStatusCompleted)
}
func (emailQueue *EmailQueue) SetFailedStatus(reason string) error {

	// Удаляем все задачи из WorkFlow
	err := db.Where("owner_id = ? AND owner_type = ?", emailQueue.Id, emailQueue.GetType()).Delete(&MTAWorkflow{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из MTAWorkflow: %v\b", err)
		return utils.Error{Message: "Ошибка завершения кампании - невозможно удалить остаток задач"}
	}

	// На всякий случай удаляем задачу по запуску этого уведомления, чтобы не было дубля (прошлые не удаляются)
	if err := emailQueue.RemoveRunTask(); err != nil {
		return err
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailQueue.updateWorkStatus(WorkStatusFailed, reason)
}
func (emailQueue *EmailQueue) SetCancelledStatus() error {

	// Возможен вызов из состояния active, paused, planned: завершить кампанию => cancelled
	if emailQueue.Status != WorkStatusActive && emailQueue.Status != WorkStatusPaused && emailQueue.Status != WorkStatusPlanned {
		reason := "Невозможно отменить уведомление,"
		switch emailQueue.Status {
		case WorkStatusPending:
			reason += "т.к. серия писем еще в стадии разработки"
		case WorkStatusCompleted:
			reason += "т.к. серия писем уже завершено"
		case WorkStatusFailed:
			reason += "т.к. серия писем завершено с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. серия писем уже отменено"
		}
		return utils.Error{Message: reason}
	}

	// Удаляем все задачи из WorkFlow
	err := db.Where("owner_id = ? AND owner_type = ?", emailQueue.Id, emailQueue.GetType()).Delete(&MTAWorkflow{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из MTAWorkflow: %v\b", err)
		return utils.Error{Message: "Ошибка завершения кампании - невозможно удалить остаток задач"}
	}

	// На всякий случай удаляем задачу по запуску этой кампании, чтобы не было дубля
	if err := emailQueue.RemoveRunTask(); err != nil {
		return err
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailQueue.updateWorkStatus(WorkStatusCancelled)
}

// Проверяет возможность отправки email-уведомления
func (emailQueue *EmailQueue) Validate() error {

	_, err := GetAccount(emailQueue.AccountId)
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается найти аккаунт"}
	}

	if err := emailQueue.load([]string{"EmailQueueEmailTemplates","EmailQueueEmailTemplates.EmailTemplates","EmailQueueEmailTemplates.EmailBox"}); err != nil {
		return err
	}

	steps := make([]EmailQueueEmailTemplate,0)
	/*filter := map[string]interface{}{
		"account_id":emailQueue.AccountId,
	}*/
	//
	// err = db.Table("email_queue_email_templates").
	err = db.Model(&EmailQueueEmailTemplate{}).
		Preload("EmailTemplate").Preload("EmailBox").
		Where( "account_id = ? AND email_queue_id = ? AND enabled = 'true'", emailQueue.AccountId, emailQueue.Id).
		Find(&steps).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return err
	}
	if err == gorm.ErrRecordNotFound || len(steps) < 1 {
		return utils.Error{Message: "Для запуска необходимо минимум 1 активное письмо"}
	}

	// Проверяем каждый активный шаблон
	for i := range steps {
		tpl :=  steps[i]

		// Проверяем только активные
		if !tpl.IsActive() {
			continue
		}
		if err := tpl.Validate(); err != nil {
			 return err
		 }
	}


	return nil
}
// Удаляет связанную задачу по запуск
func (emailQueue *EmailQueue) RemoveRunTask() error {

	// Удаляем все задачи, которые можно "выполнить" в будущем
	err := db.Where("owner_id = ? AND owner_type = ? AND (status != ? OR status != ? OR status != ?)",
		emailQueue.Id, TaskEmailNotificationRun, WorkStatusCompleted, WorkStatusCancelled, WorkStatusFailed).Delete(&TaskScheduler{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из TaskScheduler: %v\b", err)
		return utils.Error{Message: "Ошибка удаления задачи по запуску кампании"}
	}

	return nil
}

func (emailQueue *EmailQueue) AppendAllUsers() (uint, error) {

	offset := int64(0)
	limit := uint(20)
	total := int64(1)

	emailQueueAdded := uint(0)

	for offset < total {

		_aUsers, _total, err := (AccountUser{}).getPaginationListByRole(emailQueue.AccountId, nil, int(offset), int(limit), "public_id")
		if err != nil {
			break
		}

		// Проверяем пользователей по истории
		for i := range _aUsers {
			
			if !emailQueue.ExistUserById(_aUsers[i].UserId) {
				_ = emailQueue.AppendUser(_aUsers[i].UserId)
				emailQueueAdded++
			}
		}

		// Обновляем данные по цепочке
		offset = offset + int64(len(_aUsers))
		total = _total
	}


	return emailQueueAdded, nil

}

// Проверяет по истории и по наличию в очереди
func (emailQueue *EmailQueue) ExistUserById(userId uint) bool {
	if emailQueue.Id < 1 || userId < 1 {
		return false
	}

	// Проверяет историю
	if (MTAHistory{}).ExistUserById(userId, emailQueue) {
		return true
	}

	// Проверяем очередь
	if (MTAWorkflow{}).ExistUserById(userId, emailQueue) {
		return true
	}

	return false
}