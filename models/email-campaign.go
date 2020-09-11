package models

import (
	"database/sql"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"strings"
	"time"
)

type EmailCampaign struct {
	Id     			uint   		`json:"id" gorm:"primaryKey"`
	PublicId		uint   		`json:"public_id" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 		`json:"-" gorm:"type:int;index;not null;"`
	HashId 			string 		`json:"hash_id" gorm:"type:varchar(12);unique_index;not null;"`

	// Результат выполнения: planned / pending / active / completed / failed / cancelled .
	Status 			WorkStatus 	`json:"status" gorm:"type:varchar(18);default:'pending'"`
	FailedStatus	*string 		`json:"failed_status" gorm:"type:varchar(255);"`

	// Планируемое время старта
	ScheduleRun		*time.Time 	`json:"schedule_run"`

	// Имя кампании - Ежемесячный дайджест !
	Name 			string 	`json:"name" gorm:"type:varchar(128);default:'New campaign'"`

	// Тема сообщения и preview-текст, компилируются
	Subject			*string 	`json:"subject" gorm:"type:varchar(255);"`
	PreviewText		*string 	`json:"preview_text" gorm:"type:varchar(255);"`

	// Шаблон email-сообщения
	EmailTemplateId *uint 	`json:"email_template_id" gorm:"type:int;"`
	EmailTemplate 	EmailTemplate 	`json:"email_template"`

	// Отправитель, может устанавливаться в конце
	EmailBoxId		*uint 		`json:"email_box_id" gorm:"type:int;"`
	EmailBox 		EmailBox 	`json:"email_box"`

	// RecipientList	postgres.Jsonb 	`json:"recipientList" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	// UserSegments	[]UserSegment `json:"userSegments"`
	UsersSegmentId	*uint 		`json:"users_segment_id" gorm:"type:int;"`
	UsersSegment 	UsersSegment `json:"users_segment"`

	// Hidden vars
	Queue 			int64 	`json:"_queue" gorm:"-"` 	// сколько подписчиков еще в процессе отправки кампании
	Recipients 		uint 	`json:"_recipients" gorm:"-"` // << всего успешно отправлено писем
	OpenRate 		float64 `json:"_open_rate" gorm:"-"`
	UnsubscribeRate float64 `json:"_unsubscribe_rate" gorm:"-"`

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

// ############# Entity interface #############
func (emailCampaign EmailCampaign) GetId() uint { return emailCampaign.Id }
func (emailCampaign *EmailCampaign) setId(id uint) { emailCampaign.Id = id }
func (emailCampaign *EmailCampaign) setPublicId(publicId uint) { emailCampaign.PublicId = publicId }
func (emailCampaign EmailCampaign) GetAccountId() uint { return emailCampaign.AccountId }
func (emailCampaign *EmailCampaign) setAccountId(id uint) { emailCampaign.AccountId = id }
func (EmailCampaign) SystemEntity() bool { return false }
func (EmailCampaign) GetType() string { return "email_campaigns" }
func (emailCampaign EmailCampaign) IsEnabled() bool {

	// т.к. статус для обхода воркера
	if emailCampaign.Status == WorkStatusPending || emailCampaign.Status == WorkStatusFailed || emailCampaign.Status == WorkStatusCancelled {
		return false
	}

	return true
}
func (emailCampaign EmailCampaign) IsActive() bool {
	return emailCampaign.Status == WorkStatusActive
}
// ############# End Entity interface #############

func (EmailCampaign) PgSqlCreate() {
	db.Migrator().CreateTable(&EmailCampaign{})
	// db.Model(&EmailCampaign{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&EmailCampaign{}).AddForeignKey("email_template_id", "email_templates(id)", "SET NULL", "CASCADE")
	// db.Model(&EmailCampaign{}).AddForeignKey("email_box_id", "email_boxes(id)", "SET NULL", "CASCADE")
	// db.Model(&EmailCampaign{}).AddForeignKey("users_segment_id", "users_segments(id)", "SET NULL", "CASCADE")
	err := db.Exec("ALTER TABLE email_campaigns " +
		"ADD CONSTRAINT email_campaigns_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT email_campaigns_email_template_id_fkey FOREIGN KEY (email_template_id) REFERENCES email_templates(id) ON DELETE SET NULL ON UPDATE CASCADE," +
		"ADD CONSTRAINT email_campaigns_email_box_id_fkey FOREIGN KEY (email_box_id) REFERENCES email_boxes(id) ON DELETE SET NULL ON UPDATE CASCADE," +
		"ADD CONSTRAINT email_campaigns_users_segment_id_fkey FOREIGN KEY (users_segment_id) REFERENCES users_segments(id) ON DELETE SET NULL ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (emailCampaign *EmailCampaign) BeforeCreate(tx *gorm.DB) error {
	emailCampaign.Id = 0
	emailCampaign.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))

	if emailCampaign.ScheduleRun == nil {
		emailCampaign.ScheduleRun = utils.TimeP(time.Now().UTC())
	}

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&EmailCampaign{}).Where("account_id = ?",  emailCampaign.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	emailCampaign.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (emailCampaign *EmailCampaign) AfterFind(tx *gorm.DB) (err error) {

	// Рассчитываем сколько пользователей сейчас в очереди
	inQueue := int64(0)
	err = db.Model(&MTAWorkflow{}).Where("account_id = ? AND owner_id = ? AND owner_type = ?", emailCampaign.AccountId, emailCampaign.Id, EmailSenderCampaign).Count(&inQueue).Error
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	if err == gorm.ErrRecordNotFound {emailCampaign.Queue = 0} else { emailCampaign.Queue = inQueue}

	stat := struct {
		Recipients uint  	// << Всего отправок
		Opens uint    		// (opens >=1)
		Unsubscribed uint 	// (unsubscribed = true)
	}{0,0,0}
	if err = db.Raw("SELECT   \n--        COUNT(CASE WHEN succeed = true THEN 1 END) AS recipients, -- успешно отправленных   \n       COUNT(*) AS recipients, -- успешно отправленных   \n       COUNT(CASE WHEN opens >=1 THEN 1 END) AS opens, -- открытий среди успешно отправленных   \n       COUNT(CASE WHEN unsubscribed = true THEN 1 END) AS unsubscribed \nFROM mta_histories \nWHERE account_id = ? AND owner_id = ? AND owner_type = 'email_campaigns';", emailCampaign.AccountId, emailCampaign.Id).
		Scan(&stat).Error; err != nil {
		return err
	}

	emailCampaign.Recipients = stat.Recipients // << Сколько всего реально было отправлено писем.
	if stat.Opens > 0 && stat.Recipients > 0{
		emailCampaign.OpenRate = (float64(stat.Opens) / float64(stat.Recipients))*100
	} else {
		emailCampaign.OpenRate = 0
	}
	if stat.Unsubscribed > 0 && stat.Recipients > 0 {
		emailCampaign.UnsubscribeRate = (float64(stat.Unsubscribed) / float64(stat.Recipients))*100
	} else {
		emailCampaign.UnsubscribeRate = 0
	}

	return nil
}

// ######### CRUD Functions ############
func (emailCampaign EmailCampaign) create() (Entity, error)  {

	_item := emailCampaign
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (EmailCampaign) get(id uint, preloads []string) (Entity, error) {

	var emailCampaign EmailCampaign

	err := emailCampaign.GetPreloadDb(false,false,preloads).First(&emailCampaign, id).Error
	if err != nil {
		return nil, err
	}
	return &emailCampaign, nil
}
func (emailCampaign *EmailCampaign) load(preloads []string) error {
	if emailCampaign.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailCampaign - не указан  Id"}
	}

	err := emailCampaign.GetPreloadDb(false, false, preloads).First(emailCampaign, emailCampaign.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailCampaign *EmailCampaign) loadByPublicId(preloads []string) error {
	
	if emailCampaign.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailCampaign - не указан  Id"}
	}
	if err := emailCampaign.GetPreloadDb(false,false, preloads).First(emailCampaign, "account_id = ? AND public_id = ?", emailCampaign.AccountId, emailCampaign.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (EmailCampaign) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return EmailCampaign{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (EmailCampaign) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	emailCampaigns := make([]EmailCampaign,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&EmailCampaign{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailCampaigns, "name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailCampaign{}).GetPreloadDb(false,false,nil).
			Where("account_id = ? AND name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&EmailCampaign{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailCampaigns).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailCampaign{}).GetPreloadDb(false,false,nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailCampaigns))
	for i := range emailCampaigns {
		entities[i] = &emailCampaigns[i]
	}

	return entities, total, nil
}

func (emailCampaign *EmailCampaign) update(input map[string]interface{}, preloads []string) error {

	utils.FixInputHiddenVars(&input)
	input = utils.FixInputDataTimeVars(input,[]string{"scheduleRun"})
	delete(input,"email_template")
	delete(input,"email_box")
	delete(input,"users_segment")
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}

	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","email_template_id","email_box_id","users_segment_id"}); err != nil {
		return err
	}

	if err := emailCampaign.GetPreloadDb(false, false, nil).Where("id = ?", emailCampaign.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := emailCampaign.GetPreloadDb(false,false, preloads).First(emailCampaign, emailCampaign.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (emailCampaign *EmailCampaign) delete () error {
	return emailCampaign.GetPreloadDb(true,false,nil).Where("id = ?", emailCampaign.Id).Delete(emailCampaign).Error
}
// ######### END CRUD Functions ############

func (emailCampaign *EmailCampaign) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(emailCampaign)
	} else {
		_db = _db.Model(&EmailCampaign{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"EmailTemplate","EmailBox","UsersSegment"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// Подготавливает рассылку к первичному запуску: извлекает сегмент пользователей и добавляет пользователей в mta-workflow
func (emailCampaign *EmailCampaign) Execute() error {

	// 1. Проверяем все данные перед маршем -\0/-
	
	// Проверяем статус кампании, может быть только planned, т.к. это первоначальный (!!!) запуску кампании
	if emailCampaign.Status != WorkStatusPlanned {
		return utils.Error{Message: fmt.Sprintf("Кампания не может быть запущена т.к. находится в статусе - '%v'", emailCampaign.Status)}
	}

	// Get Account
	_, err := GetAccount(emailCampaign.AccountId)
	if err != nil {
		emailCampaign.SetFailedStatus(fmt.Sprintf("Ошибка определения аккаунта: %v\n", err.Error()))
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается найти аккаунт"}
	}

	if err := emailCampaign.Validate();err != nil {
		return err
	}

	// 2. Переводим Кампанию в состояние "Выполняется"
	if err := emailCampaign.SetActiveStatus(); err != nil {return err}

	// 3. Собираем всех пользователь сегмента из базы
	users, err := emailCampaign.getUsersBySegment()
	if err != nil {
		return err
	}

	// Шаблон-заготовка для каждого пользователя под задачу в mta-workflow
	mtaWorkflow := MTAWorkflow{
		AccountId: emailCampaign.AccountId,
		OwnerId: emailCampaign.Id,
		OwnerType: EmailSenderCampaign,
		ExpectedTimeStart: *emailCampaign.ScheduleRun, // время запуска - время запуска кампании
		// UserId: users[i].Id, // << ставим во время цикла
		NumberOfAttempts: 0,
	}

	// Создаем под каждого пользователя задачу в mta-workflow
	for i := range users {
		mtaWorkflow.UserId = users[i].Id
		if _, err = mtaWorkflow.create(); err != nil {
			log.Printf("Ошибка добавления пользователя [%v] в очередь при выполнении кампании: %v", users[i].Id, err)

		}
	}

	return nil
}

// Прямая функция для обновления, без проверок
func (emailCampaign *EmailCampaign) ChangeWorkStatus(status WorkStatus, reason... string) error {

	switch status {
	case WorkStatusPending:
		return emailCampaign.SetPendingStatus()
	case WorkStatusPlanned:
		return emailCampaign.SetPlannedStatus()
	case WorkStatusActive:
		return emailCampaign.SetActiveStatus()
	case WorkStatusPaused:
		return emailCampaign.SetPausedStatus()
	case WorkStatusFailed:
		_reason := "" 
		if len(reason) > 0 {
			_reason = reason[0]
		}
		return emailCampaign.SetFailedStatus(_reason)
	case WorkStatusCompleted:
		return emailCampaign.SetCompletedStatus()
	case WorkStatusCancelled:
		return emailCampaign.SetCancelledStatus()
	default:
		return utils.Error{Message: "Статус не опознан"}
	}

}
func (emailCampaign *EmailCampaign) updateWorkStatus(status WorkStatus, reason... string) error {
	_reason := ""
	if len(reason) > 0 {
		_reason = reason[0]
	}
	
	return emailCampaign.update(map[string]interface{}{
		"status":	status,
		"failed_status": _reason,
	},nil)
}

func (emailCampaign *EmailCampaign) SetPendingStatus() error {

	// Возможен вызов из состояния planned: вернуть на доработку => pending
	if emailCampaign.Status != WorkStatusPlanned {
		reason := "Невозможно установить статус,"
		switch emailCampaign.Status {
		case WorkStatusPending:
			reason += "т.к. кампания уже в разработке"
		case WorkStatusActive:
			reason += "т.к. кампания в процессе рассылки"
		case WorkStatusPaused:
			reason += "т.к. кампания на паузе, но в процессе рассылки"
		case WorkStatusFailed:
			reason += "т.к. кампания завершена с ошибкой"
		case WorkStatusCompleted:
			reason += "т.к. кампания уже завершена"
		case WorkStatusCancelled:
			reason += "т.к. кампания отменена"
		}
		return utils.Error{Message: reason}
	}

	// fix: удаляем задачу в TaskScheduler
	if err := emailCampaign.RemoveRunTask(); err != nil {
		return err
	}

	return emailCampaign.updateWorkStatus(WorkStatusPending)
}
func (emailCampaign *EmailCampaign) SetPlannedStatus() error {

	// Возможен вызов из состояния pending: запланировать кампанию => planned
	if emailCampaign.Status != WorkStatusPending  {
		reason := "Невозможно запланировать кампанию,"
		switch emailCampaign.Status {
		case WorkStatusPlanned:
			reason += "т.к. кампания уже в плане"
		case WorkStatusActive:
			reason += "т.к. кампания уже в процессе рассылки"
		case WorkStatusPaused:
			reason += "т.к. кампания на паузе, но уже в процессе рассылки"
		case WorkStatusCompleted:
			reason += "т.к. кампания уже завершена"
		case WorkStatusFailed:
			reason += "т.к. кампания завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. кампания отменена"
		}
		return utils.Error{Message: reason}
	}

	// Get Account
	account, err := GetAccount(emailCampaign.AccountId)
	if err != nil {
		return utils.Error{Message: "Не удается найти аккаунт"}
	}

	// Проверяем кампанию и шаблон, чтобы не ставить в план не рабочую кампанию.
	if err := emailCampaign.Validate(); err != nil { return err  }

	// На всякий случай удаляем задачу по запуску этой кампании, чтобы не было дубля
	if err := emailCampaign.RemoveRunTask(); err != nil {
		return err
	}

	// Создаем новый объект task для добавлении кампании в TaskScheduler
	task := TaskScheduler {
		AccountId: emailCampaign.AccountId,
		OwnerType: TaskEmailCampaignRun,
		OwnerId: emailCampaign.Id,
		ExpectedTimeToStart: emailCampaign.ScheduleRun.Add(-time.Minute*5),// запускаем задачу (но не кампанию) за 5 минут (= с запасом)
		IsSystem: true, // системная задача ли ?
		Status: WorkStatusPlanned,
	}

	// Создаем задачу по отправке рекламной кампании
	_, err = account.CreateEntity(&task)
	if err != nil {
		return utils.Error{Message: "Ошибка создания задания по запуску кампании"}
	}

	// Переводим в состояние "Запланирована", т.к. все проверки пройдены и можно ставить ее в планировщик
	return emailCampaign.updateWorkStatus(WorkStatusPlanned)
}
func (emailCampaign *EmailCampaign) SetActiveStatus() error {

	// Возможен вызов из состояния planned или paused: запустить кампанию => active
	if emailCampaign.Status != WorkStatusPlanned && emailCampaign.Status != WorkStatusPaused {
		reason := "Невозможно запустить кампанию,"
		switch emailCampaign.Status {
		case WorkStatusPending:
			reason += "т.к. кампания еще в стадии разработки"
		case WorkStatusActive:
			reason += "т.к. кампания уже в процессе рассылки"
		case WorkStatusCompleted:
			reason += "т.к. кампания уже завершена"
		case WorkStatusFailed:
			reason += "т.к. кампания завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. кампания отменена"
		}
		return utils.Error{Message: reason}
	}

	// Снова проверяем кампанию и шаблон
	if err := emailCampaign.Validate(); err != nil { return err  }

	// Переводим в состояние "Активна", т.к. все проверки пройдены и можно продолжить ее выполнение
	return emailCampaign.updateWorkStatus(WorkStatusActive)
}
func (emailCampaign *EmailCampaign) SetPausedStatus() error {

	// Возможен вызов из состояния active: приостановить кампанию => paused
	if emailCampaign.Status != WorkStatusActive {
		reason := "Невозможно приостановить кампанию,"
		switch emailCampaign.Status {
		case WorkStatusPending:
			reason += "т.к. кампания еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. кампания уже в стадии планирования"
		case WorkStatusPaused:
			reason += "т.к. кампания уже приостановлена"
		case WorkStatusCompleted:
			reason += "т.к. кампания уже завершена"
		case WorkStatusFailed:
			reason += "т.к. кампания завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. кампания отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Приостановлена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailCampaign.updateWorkStatus(WorkStatusPaused)
}
func (emailCampaign *EmailCampaign) SetCompletedStatus() error {

	// Возможен вызов из состояния active, paused: завершить кампанию => completed
	// Сбрасываются все задачи из очереди
	if emailCampaign.Status != WorkStatusActive && emailCampaign.Status != WorkStatusPaused {
		reason := "Невозможно завершить кампанию,"
		switch emailCampaign.Status {
		case WorkStatusPending:
			reason += "т.к. кампания еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. кампания еще в стадии планирования"
		case WorkStatusCompleted:
			reason += "т.к. кампания уже завершена"
		case WorkStatusFailed:
			reason += "т.к. кампания завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. кампания отменена"
		}
		return utils.Error{Message: reason}
	}

	// Удаляем все задачи из WorkFlow
	err := db.Where("owner_id = ? AND owner_type = ?", emailCampaign.Id, emailCampaign.GetType()).Delete(&MTAWorkflow{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из MTAWorkflow: %v\b", err)
		return utils.Error{Message: "Ошибка завершения кампании - невозможно удалить остаток задач"}
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailCampaign.updateWorkStatus(WorkStatusCompleted)
}
func (emailCampaign *EmailCampaign) SetFailedStatus(reason string) error {

	// Вызов из любого состояния (хотя это не так)
	// Сбрасываются все задачи из очереди

	// Удаляем все задачи из WorkFlow
	err := db.Where("owner_id = ? AND owner_type = ?", emailCampaign.Id, emailCampaign.GetType()).Delete(MTAWorkflow{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из MTAWorkflow: %v\b", err)
		return utils.Error{Message: "Ошибка завершения кампании - невозможно удалить остаток задач"}
	}

	// На всякий случай удаляем задачу по запуску этой кампании, чтобы не было дубля (прошлые не удаляются)
	if err := emailCampaign.RemoveRunTask(); err != nil {
		return err
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailCampaign.updateWorkStatus(WorkStatusFailed, reason)
}
func (emailCampaign *EmailCampaign) SetCancelledStatus() error {

	// Возможен вызов из состояния active, paused, planned: завершить кампанию => cancelled
	if emailCampaign.Status != WorkStatusActive && emailCampaign.Status != WorkStatusPaused && emailCampaign.Status != WorkStatusPlanned {
		reason := "Невозможно отменить кампанию,"
		switch emailCampaign.Status {
		case WorkStatusPending:
			reason += "т.к. кампания еще в стадии разработки"
		case WorkStatusCompleted:
			reason += "т.к. кампания уже завершена"
		case WorkStatusFailed:
			reason += "т.к. кампания завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. кампания уже отменена"
		}
		return utils.Error{Message: reason}
	}

	// Удаляем все задачи из WorkFlow
	err := db.Where("owner_id = ? AND owner_type = ?", emailCampaign.Id, emailCampaign.GetType()).Delete(MTAWorkflow{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из MTAWorkflow: %v\b", err)
		return utils.Error{Message: "Ошибка завершения кампании - невозможно удалить остаток задач"}
	}

	// На всякий случай удаляем задачу по запуску этой кампании, чтобы не было дубля
	if err := emailCampaign.RemoveRunTask(); err != nil {
		return err
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailCampaign.updateWorkStatus(WorkStatusCancelled)
}

func (emailCampaign *EmailCampaign) getUsersBySegment() ([]User, error) {

	if emailCampaign.UsersSegmentId == nil {
		return nil, utils.Error{Message: "Не выбран пользовательский сегмент"}
	}

	segment := UsersSegment{Id: *emailCampaign.UsersSegmentId}
	if err := segment.load(nil); err != nil {
		return nil, err
	}

	users := make([]User,0)
	offset := int64(0)
	limit := uint(100)
	total := int64(1)

	for offset < total {

		_users, _total, err := segment.ChunkUsers(offset, limit)
		if err != nil {
			break
		}

		// добавляем в общий массив пользователей
		users = append(users, _users...)
		offset = offset + int64(len(_users))
		total = _total
	}

	return users, nil
}

func (emailCampaign EmailCampaign) Validate() error {

	account, err := GetAccount(emailCampaign.AccountId)
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается найти аккаунт"}
	}
	

	// Проверяем тело сообщения (не должно быть пустое)
	if emailCampaign.Subject == nil {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет темы сообщения"}
	}
	if emailCampaign.PreviewText == nil {
		emailCampaign.PreviewText = utils.STRp("")
	}

	if emailCampaign.ScheduleRun == nil {
		emailCampaign.ScheduleRun = utils.TimeP(time.Now().UTC())
	}

	// Проверяем ключи и загружаем еще раз все данные для отправки сообщения
	if emailCampaign.EmailTemplateId == nil {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет установленного шаблона email-сообщения"}
	}
	err = account.LoadEntity(&emailCampaign.EmailTemplate, *emailCampaign.EmailTemplateId,nil)
	if err != nil {
		log.Printf("Ошибка загрузки шаблона email-сообщения для кампании [%v]: %v\n", emailCampaign.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки шаблона email-сообщения"}
	}

	if emailCampaign.EmailBoxId == nil  {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет установленного адреса отправителя"}
	}
	err = account.LoadEntity(&emailCampaign.EmailBox, *emailCampaign.EmailBoxId,nil)
	if err != nil {
		log.Printf("Ошибка загрузки адреса отправителя для кампании [%v]: %v\n", emailCampaign.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки адреса отправителя"}
	}

	// Проверяем DKIM подпись
	var webSite WebSite
	err = account.LoadEntity(&webSite, emailCampaign.EmailBox.WebSiteId,nil)
	if err != nil {
		log.Printf("Ошибка загрузки web site отправителя для кампании [%v]: %v\n", emailCampaign.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки адреса отправителя"}
	}

	if err := webSite.ValidateDKIM(); err != nil { return err }

	if *emailCampaign.UsersSegmentId < 1 {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет установленного сегмента пользователей"}
	}
	err = account.LoadEntity(&emailCampaign.UsersSegment, *emailCampaign.UsersSegmentId,nil)
	if err != nil {
		log.Printf("Ошибка загрузка сегмента пользователей для кампании [%v]: %v\n", emailCampaign.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки сегмента пользователей"}
	}

	// Тестовые данные
	// Локальные данные аккаунта, пользователя
	data := make(map[string]interface{})

	data["accountId"] = account.Id
	data["Account"] = account.GetDepersonalizedData() // << хз
	data["userId"] = 0  // example id, need test user
	data["User"] = User{Id: 1,Name: utils.STRp("TIvan"), Username: utils.STRp("TUsername"), Surname: utils.STRp("TSurname"), Patronymic: utils.STRp("TPatronymic"),
		PhoneRegion: utils.STRp("RU"), Phone: utils.STRp("+79251000000")} // << хз
	data["unsubscribeUrl"] = "/unsubscribe_url"
	
	viewData := ViewData {
		Subject: *emailCampaign.Subject,
		PreviewText: *emailCampaign.PreviewText,
		Data: data,
		Json: data,
		UnsubscribeURL: "",
		PixelURL: "",
		PixelHTML: "<div></div>",
	}

	return emailCampaign.EmailTemplate.Validate(&viewData)
}

// Удаляет связанную задачу по запуск
func (emailCampaign EmailCampaign) RemoveRunTask() error {

	// Удаляем все задачи, которые можно "выполнить" в будущем
	err := db.Where("owner_id = ? AND owner_type = ? AND (status != ? OR status != ? OR status != ?)",
		emailCampaign.Id, TaskEmailCampaignRun, WorkStatusCompleted, WorkStatusCancelled, WorkStatusFailed).Delete(TaskScheduler{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из TaskScheduler: %v\b", err)
		return utils.Error{Message: "Ошибка удаления задачи по запуску кампании"}
	}

	return nil
}
func (emailCampaign *EmailCampaign) CheckDoubleFromHistory() (uint, error) {

	// Get Account

	histories := make([]MTAHistory,0)

	offset := int64(0)
	limit := uint(100)
	total := int64(1)

	for offset < total {

		_histories, _total, err := (MTAHistory{}).getPaginationListByOwner(emailCampaign, int(offset), int(limit), "id")
		if err != nil {
			break
		}

		// добавляем в общий массив пользователей
		histories = append(histories, _histories...)
		offset = offset + int64(len(_histories))
		total = _total
	}

	count := int64(0)
	dobules := uint(0)

	// Создаем под каждого пользователя задачу в mta-workflow
	for i := range histories {
		email := histories[i].Email
		err := db.Model(&MTAHistory{}).Where("account_id = ? AND owner_id = ? AND owner_type = ? AND email = ?",
			emailCampaign.GetAccountId(), emailCampaign.GetId(), emailCampaign.GetType(), email).Count(&count).Error
		if err != nil {
			return 0, utils.Error{Message: fmt.Sprintf("Ошибка извлечения данных из истории рассылок: %v\n", err)}
		}
		if count > 1 {
			fmt.Printf("Дубль [%v]: %v\n", count, email)
			dobules++
		}
	}

	return dobules, nil
}

func (emailCampaign EmailCampaign) CheckDoubleFromHistoryTest() (uint, error) {
	users, err := emailCampaign.getUsersBySegment()
	return uint(len(users)), err
}
