package models

import (
	"errors"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net/mail"
	"strings"
	"time"
)

// активная база с реальными задачами по отправке
type MTAWorkflow struct {

	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// email_queues, email_campaigns, email_notifications
	OwnerType	EmailSenderType	`json:"owner_type" gorm:"varchar(32);not null;"` // << тип события: кампания, серия, уведомление
	OwnerId		uint			`json:"owner_id" gorm:"type:smallint;not null;"` // ID типа события: какая серия, компания или уведомление

	// Номер необходимого шага в серии EmailQueue. Шаг определяется ситуационно в момент Expected Time. Если шага нет - серия завершается за пользователя.
	// После выполнения - № шага увеличивается на 1
	QueueExpectedStepId	*uint	`json:"queue_expected_step_id" gorm:"type:int;"`

	// Запланированное время отправки
	ExpectedTimeStart 	time.Time `json:"expected_time_start"`

	// Id пользователя, которому отправляется серия. Обязательный параметр.
	UserId 	uint `json:"user_id" gorm:"type:int;not null;"`
	User	User `json:"user"`

	// Результат выполнения: planned / pending / completed / failed / cancelled => планируется / выполняется / выполнена / провалена / отмена
	// Status WorkStatus `json:"status" gorm:"type:varchar(18);default:'planned'"`
	
	// Число попыток отправки. Если почему-то не удалось отправить письмо есть возможность перенести отправку и повторить попытку. При 2-3х попытках, завершается неудачей.
	NumberOfAttempts uint `json:"number_of_attempts" gorm:"type:smallint;default:0;"`


	// Когда была последняя попытка отправки (обычно = CreatedAt time)
	LastTriedAt *time.Time  `json:"last_tried_at"` // << may be null

	CreatedAt 	time.Time  `json:"created_at"`
}

func (MTAWorkflow) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&MTAWorkflow{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&MTAWorkflow{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&MTAWorkflow{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE mta_workflows " +
		"ADD CONSTRAINT mta_workflows_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT mta_workflows_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func (mtaWorkflow *MTAWorkflow) BeforeCreate(tx *gorm.DB) error {
	mtaWorkflow.Id = 0
	return nil
}
func (mtaWorkflow *MTAWorkflow) AfterCreate(tx *gorm.DB) error {
	// AsyncFire(*Event{}.PaymentCreated(mtaWorkflow.AccountId, mtaWorkflow.Id))
	return nil
}
func (mtaWorkflow *MTAWorkflow) AfterUpdate(tx *gorm.DB) (err error) {

	// AsyncFire(*Event{}.PaymentUpdated(mtaWorkflow.AccountId, mtaWorkflow.Id))

	return nil
}
func (mtaWorkflow *MTAWorkflow) AfterDelete(tx *gorm.DB) (err error) {
	// AsyncFire(*Event{}.PaymentDeleted(mtaWorkflow.AccountId, mtaWorkflow.Id))
	return nil
}
func (mtaWorkflow *MTAWorkflow) AfterFind(tx *gorm.DB) (err error) {
	return nil
}

func (mtaWorkflow *MTAWorkflow) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(mtaWorkflow)
	} else {
		_db = _db.Model(&MTAWorkflow{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"User"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ############# Entity interface #############
func (mtaWorkflow MTAWorkflow) GetId() uint { return mtaWorkflow.Id }
func (mtaWorkflow *MTAWorkflow) setId(id uint) { mtaWorkflow.Id = id }
func (mtaWorkflow *MTAWorkflow) setPublicId(publicId uint) { }
func (mtaWorkflow MTAWorkflow) GetAccountId() uint { return mtaWorkflow.AccountId }
func (mtaWorkflow *MTAWorkflow) setAccountId(id uint) { mtaWorkflow.AccountId = id }
func (MTAWorkflow) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (mtaWorkflow MTAWorkflow) create() (Entity, error)  {

	_item := mtaWorkflow
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (MTAWorkflow) get(id uint, preloads []string) (Entity, error) {

	var item CartItem

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (mtaWorkflow *MTAWorkflow) load(preloads []string) error {
	if mtaWorkflow.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := mtaWorkflow.GetPreloadDb(false, false, preloads).First(mtaWorkflow, mtaWorkflow.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (mtaWorkflow *MTAWorkflow) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (MTAWorkflow) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return MTAWorkflow{}.getPaginationList(accountId, 0, 25, sortBy, "",nil,preload)
}
func (MTAWorkflow) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	webHooks := make([]MTAWorkflow,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&MTAWorkflow{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&MTAWorkflow{}).GetPreloadDb(false,false,nil).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&MTAWorkflow{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&MTAWorkflow{}).GetPreloadDb(false,false,nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}
func (mtaWorkflow *MTAWorkflow) update(input map[string]interface{}, preloads []string) error {
	delete(input,"user")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"owner_id","user_id","queue_expected_step_id","number_of_attempts"}); err != nil {
		return err
	}

	if err := mtaWorkflow.GetPreloadDb(false, false, nil).Where("id = ?", mtaWorkflow.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := mtaWorkflow.GetPreloadDb(false,false, preloads).First(mtaWorkflow, mtaWorkflow.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (mtaWorkflow *MTAWorkflow) delete () error {

	return mtaWorkflow.GetPreloadDb(true,false,nil).Where("id = ?", mtaWorkflow.Id).Delete(mtaWorkflow).Error
}
// ######### END CRUD Functions ############

// Создание пакета для добавления в буфер на отправку.
// При ошибке: реагирует соответствующим образом: либо останавливает EmailSender, либо завершает со статусом Failed и удаляет все задачи по отправке
func (mtaWorkflow *MTAWorkflow) Execute() error {

	// Локальные данные аккаунта, пользователя
	data := make(map[string]interface{})

	account, err := GetAccount(mtaWorkflow.AccountId)
	if err != nil {
		mtaWorkflow.stopEmailSender(fmt.Sprintf("Account not found: %v",err.Error()))
		return err
	}

	user, err := account.GetUser(mtaWorkflow.UserId)
	if err != nil {
		mtaWorkflow.removeWorkflowUser(mtaWorkflow.UserId)
		return err
	}

	// Проверяем статус подписки
	if !user.Subscribed {
		mtaWorkflow.removeWorkflowUser(mtaWorkflow.UserId)
		return utils.Error{Message: "Невозможно отправить письмо пользователю, т.к. он отписан от всех подписок"}
	}

	// ############ Готовим данные ############
	var sender EmailSender
	switch mtaWorkflow.OwnerType {
		case EmailSenderQueue:
			_e := EmailQueue{}
			err := account.LoadEntity(&_e, mtaWorkflow.OwnerId,nil)
			if err != nil {
				mtaWorkflow.stopEmailSender(fmt.Sprintf("Email Queue not found: %v",err.Error()))
				return err
			}
			sender = &_e
		case EmailSenderCampaign:
			_e := EmailCampaign{}
			err := account.LoadEntity(&_e, mtaWorkflow.OwnerId,nil)
			if err != nil {
				mtaWorkflow.stopEmailSender(fmt.Sprintf("Email Campaign not found: %v",err.Error()))
				return err
			}
			sender = &_e
		case EmailSenderNotification:
			_e := EmailNotification{}
			err := account.LoadEntity(&_e, mtaWorkflow.OwnerId,nil)
			if err != nil {
				mtaWorkflow.stopEmailSender(fmt.Sprintf("Email Notification not found: %v",err.Error()))
				return err
			}
			sender = &_e
	default:
		mtaWorkflow.stopEmailSender("Ошибка: неопознанный тип отправления сообщения")
		return utils.Error{Message: "Ошибка: неопознанный тип отправления сообщения"}
	}

	// тут у нас есть Конверт для отправки
	var emailBox EmailBox
	var emailTemplate EmailTemplate
	var Subject, PreviewText string
	var queueStepId = uint(0)

	if sender.GetType() == EmailSenderQueue {

		// Проверяем состоянии кампании/очереди/уведомления
		if !sender.IsActive() {
			// Тут могут копиться люди в очереди, поэтому возвращаем ошибку
			mtaWorkflow.pausedEmailSender("Статус отправителя - is not active!")
			return utils.Error{ Message: fmt.Sprintf("Невозможно отправить письмо, т.к. объект [id = %v] не запущен\n", sender.GetId())}
		}

		emailQueue, ok := sender.(*EmailQueue)
		if !ok {
			mtaWorkflow.stopEmailSender("Ошибка преобразования статуса отправителя в email queue")
			return errors.New("Ошибка преобразования в Email Queue")
		}

		if mtaWorkflow.QueueExpectedStepId == nil {
			mtaWorkflow.stopEmailSender("Данные не полные: QueueExpectedStepId == nil")
			return utils.Error{Message: "Данные не полные: QueueExpectedStepId == nil"}
		}

		step, err := emailQueue.GetNearbyActiveStep(*mtaWorkflow.QueueExpectedStepId)
		if err != nil {
			// Если нет доступных шагов - удаляем задачу
			if err2 := mtaWorkflow.delete(); err2 != nil {
				log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err2)
			}
			return err
		}

		if step.EmailBoxId == nil {
			mtaWorkflow.stopEmailSender("Данные не полные: не хватает email_box_id")
			return utils.Error{Message: "Данные не полные: не хватает emailBoxID"}
		}

		// Только теоретически такое возможно
		if !step.Enabled {
			return utils.Error{Message: "Этап отправки не активен"}
		}

		if step.Subject == nil {
			mtaWorkflow.stopEmailSender("Отсутствует Subject")
			return utils.Error{Message: "Отсутствует Subject"}
		}

		// queueStepId = step.Id
		queueStepId = step.Step

		Subject = *step.Subject
		_previewText := ""
		if step.PreviewText != nil  {
			_previewText = *step.PreviewText
		}
		PreviewText = _previewText

		// Находим шаблон письма
		if mtaWorkflow.QueueExpectedStepId == nil {
			mtaWorkflow.stopEmailSender("Данные не полные: QueueExpectedStepId == nil")
			return utils.Error{Message: "Данные не полные: QueueExpectedStepId == nil"}
		}
		_et, err := emailQueue.GetEmailTemplateByStep(*mtaWorkflow.QueueExpectedStepId)
		if err != nil {
			mtaWorkflow.stopEmailSender(fmt.Sprintf("Email template not found: %v",err.Error()))
			return err
		}
		emailTemplate = *_et

		// EmailBox
		err = account.LoadEntity(&emailBox, *step.EmailBoxId,[]string{"WebSite"})
		if err != nil {
			mtaWorkflow.stopEmailSender(fmt.Sprintf("Email box not found: %v",err.Error()))
			return err
		}
	}

	if sender.GetType() == EmailSenderNotification {

		if !sender.IsActive() {
			mtaWorkflow.pausedEmailSender("Статус отправителя - is not active!")
			return utils.Error{ Message: fmt.Sprintf("Невозможно отправить письмо, т.к. объект [id = %v] не запущен\n", sender.GetId()),
			}
		}

		emailNotification, ok := sender.(*EmailNotification)
		if !ok {
			mtaWorkflow.stopEmailSender("Ошибка преобразования статуса отправителя в email notification")
			return errors.New("Ошибка преобразования в email Notification")
		}

		if emailNotification.Subject == nil {
			mtaWorkflow.pausedEmailSender("Отсутствует тема сообщения")
			return errors.New("Отсутствует тема сообщения в Notification")
		}

		err := account.LoadEntity(&emailTemplate, *emailNotification.EmailTemplateId,nil)
		if err != nil {
			mtaWorkflow.stopEmailSender(fmt.Sprintf("Email template not found: %v",err.Error()))
			return err
		}

		err = account.LoadEntity(&emailBox, *emailNotification.EmailBoxId,[]string{"WebSite"})
		if err != nil {
			mtaWorkflow.stopEmailSender(fmt.Sprintf("Email box not found: %v",err.Error()))
			return err
		}

		previewText := ""
		if emailNotification.PreviewText != nil {
			PreviewText = *emailNotification.PreviewText
		} else {
			PreviewText = previewText
		}
		Subject = *emailNotification.Subject
	}

	if sender.GetType() == EmailSenderCampaign {

		if !sender.IsActive() {
			mtaWorkflow.pausedEmailSender("Статус отправителя - is not active!")
			return utils.Error{ Message: fmt.Sprintf("Невозможно отправить письмо, т.к. кампания [id = %v] не запущена\n", sender.GetId())}
		}

		emailCampaign, ok := sender.(*EmailCampaign)
		if !ok {
			mtaWorkflow.stopEmailSender("Ошибка преобразования статуса отправителя в email campaign")
			return errors.New("Ошибка преобразования в email campaign")
		}

		err := account.LoadEntity(&emailTemplate, *emailCampaign.EmailTemplateId,nil)
		if err != nil {
			mtaWorkflow.stopEmailSender(fmt.Sprintf("Email template not found: %v",err.Error()))
			return err
		}

		err = account.LoadEntity(&emailBox, *emailCampaign.EmailBoxId,[]string{"WebSite"})
		if err != nil {
			mtaWorkflow.stopEmailSender(fmt.Sprintf("Email box not found: %v",err.Error()))
			return err
		}

		if  emailCampaign.Subject == nil {
			mtaWorkflow.pausedEmailSender("Необходимо установить тему сообщения")
		} else {
			Subject = *emailCampaign.Subject
		}

		if  emailCampaign.PreviewText == nil {
			PreviewText = ""
		} else {
			PreviewText = *emailCampaign.PreviewText
		}

	}

	// **************************** //
	
	// Готовим переменные
	historyHashId := strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	unsubscribeUrl := account.GetUnsubscribeUrl(user.HashId, historyHashId)
	pixelURL := account.GetPixelUrl(historyHashId)

	// Подготавливаем данные для письма, чтобы можно было их использовать в шаблоне
	data["accountId"] = mtaWorkflow.AccountId
	data["Account"] = account.GetDepersonalizedData() // << хз
	data["userId"] = user.Id
	data["User"] = user.GetDepersonalizedData() // << хз
	data["unsubscribeUrl"] = unsubscribeUrl

	// Компилируем тему письма
	_subject, err := parseSubjectByData(Subject, data)
	if err != nil {
		mtaWorkflow.stopEmailSender(fmt.Sprintf("Ошибка темы сообщения: %v",err.Error()))
		return utils.Error{ Message: "Ошибка отправления Уведомления - не удается прочитать тему сообщения"}
	}
	if _subject == "" {
		_subject = fmt.Sprintf("Письмо #%v", mtaWorkflow.Id)
	}

	// Готовим еще раз полностью данные
	vData, err := emailTemplate.PrepareViewData(_subject, PreviewText, data, pixelURL, &unsubscribeUrl)
	if err != nil {
		mtaWorkflow.stopEmailSender(fmt.Sprintf("Ошибка подготовки данных для сообщения: %v",err.Error()))
		return utils.Error{ Message: "Ошибка отправления Уведомления - не удается подготовить данные для сообщения"}
	}

	var webSite WebSite
	if err = account.LoadEntity(&webSite, emailBox.WebSiteId,nil); err != nil {
		mtaWorkflow.stopEmailSender(fmt.Sprintf("Ошибка загрузки данных WebSite: %v",err.Error()))
		return utils.Error{ Message: "Ошибка отправления Уведомления - не удается загрузить данные по WebSite"}
	}

	name := ""
	if user.Name != nil { name = *user.Name }
	if user.Email == nil {
		return utils.Error{ Message: "Ошибка отправления: отсутствует почта у пользователя"}
	}

	var pkg = EmailPkg {
		// To: mail.Address{Name: *user.Name, Address: *user.Email},
		To: mail.Address{Name: name, Address: *user.Email},
		accountId: 	account.Id,
		userId: 	user.Id,
		workflowId: mtaWorkflow.Id, // мы не знаем
		webSite: 	&webSite,
		emailBox: 	&emailBox,
		emailSender: sender,
		emailTemplate: &emailTemplate,
		subject: 	_subject,
		viewData: 	vData,
		queueStepId: queueStepId,
	}

	// можно и через go
	SendEmail(pkg)

	// Проверяем не нужно ли закончить серию
	pkg.handleQueue()

	return nil
}
func (mtaWorkflow *MTAWorkflow) UpdateByNextStep(expectedStep EmailQueueEmailTemplate) error {
	// Изменяется expected step, expected_time_start, number_attems, last_tried
	return mtaWorkflow.update(map[string]interface{}{
		"queue_expected_step_id": expectedStep.Step,
		"expected_time_start": time.Now().UTC().Add(expectedStep.DelayTime),
		"number_of_attempts": 0,
		"last_tried_at": nil,
		// "status": WorkStatusPlanned,
	},nil)

}

// Останавливает отправителя и удаляет все задачи по отправке из-за ошибки
func (mtaWorkflow *MTAWorkflow) stopEmailSender(reason string) {

	account,err := GetAccount(mtaWorkflow.AccountId)
	if err != nil {
		log.Printf("stopEmailSender: Ошибка загрузки аккаунта: %v\n",err)
		return
	}

	var sender EmailSender
	switch mtaWorkflow.OwnerType {
	case EmailSenderQueue:
		_e := EmailQueue{}
		err := account.LoadEntity(&_e, mtaWorkflow.OwnerId,nil)
		if err != nil {return}
		sender = &_e
	case EmailSenderCampaign:
		_e := EmailCampaign{}
		err := account.LoadEntity(&_e, mtaWorkflow.OwnerId,nil)
		if err != nil {return}
		sender = &_e
	case EmailSenderNotification:
		_e := EmailNotification{}
		err := account.LoadEntity(&_e, mtaWorkflow.OwnerId,nil)
		if err != nil {return}
		sender = &_e
	default:
		return
	}

	_ = sender.ChangeWorkStatus(WorkStatusFailed, reason)

}
// Приостанавливает отправителя, не удаляя задачи по отправке
func (mtaWorkflow *MTAWorkflow) pausedEmailSender(reason string) {

	account,err := GetAccount(mtaWorkflow.AccountId)
	if err != nil {
		log.Printf("stopEmailSender: Ошибка загрузки аккаунта: %v\n",err)
		return
	}

	var sender EmailSender
	switch mtaWorkflow.OwnerType {
	case EmailSenderQueue:
		_e := EmailQueue{}
		err := account.LoadEntity(&_e, mtaWorkflow.OwnerId,nil)
		if err != nil {return}
		sender = &_e
	case EmailSenderCampaign:
		_e := EmailCampaign{}
		err := account.LoadEntity(&_e, mtaWorkflow.OwnerId,nil)
		if err != nil {return}
		sender = &_e
	case EmailSenderNotification:
		_e := EmailNotification{}
		err := account.LoadEntity(&_e, mtaWorkflow.OwnerId,nil)
		if err != nil {return}
		sender = &_e
	default:
		return
	}

	_ = sender.ChangeWorkStatus(WorkStatusPaused, reason)
}

// Удаляем все задачи для пользователя из workflows для этого аккаунта
func (mtaWorkflow *MTAWorkflow) removeWorkflowUser(userId uint) {

	// Удаляем все задачи из WorkFlow для этого типа отправителя
	db.Where("owner_id = ? AND owner_type = ? AND user_id = ?", mtaWorkflow.OwnerId, mtaWorkflow.OwnerType,userId).Delete(&MTAWorkflow{})

}
