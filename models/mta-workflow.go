package models

import (
	"errors"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
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

	// К какой серии писем относится задача (gorm - т.к. это AfterFind загружается)
	EmailQueue			EmailQueue 			`json:"_email_queue" gorm:"-"`
	EmailNotification	EmailNotification 	`json:"_email_notification" gorm:"-"`
	EmailCampaign		EmailCampaign 		`json:"_email_campaign" gorm:"-"`

	// Номер необходимого шага в серии EmailQueue. Шаг определяется ситуационно в момент Expected Time. Если шага нет - серия завершается за пользователя.
	// После выполнения - № шага увеличивается на 1
	QueueExpectedStepId	uint	`json:"queue_expected_step_id" gorm:"type:int;not null;"`

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
	// event.AsyncFire(Event{}.PaymentCreated(mtaWorkflow.AccountId, mtaWorkflow.Id))
	return nil
}
func (mtaWorkflow *MTAWorkflow) AfterUpdate(tx *gorm.DB) (err error) {

	// event.AsyncFire(Event{}.PaymentUpdated(mtaWorkflow.AccountId, mtaWorkflow.Id))

	return nil
}
func (mtaWorkflow *MTAWorkflow) AfterDelete(tx *gorm.DB) (err error) {
	// event.AsyncFire(Event{}.PaymentDeleted(mtaWorkflow.AccountId, mtaWorkflow.Id))
	return nil
}
func (mtaWorkflow *MTAWorkflow) AfterFind(tx *gorm.DB) (err error) {
	return nil
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
	
	wb := mtaWorkflow
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

func (MTAWorkflow) get(id uint) (Entity, error) {

	var mtaWorkflow MTAWorkflow

	err := mtaWorkflow.GetPreloadDb(false,false, true).First(&mtaWorkflow, id).Error
	if err != nil {
		return nil, err
	}
	return &mtaWorkflow, nil
}
func (MTAWorkflow) getByExternalId(externalId string) (*MTAWorkflow, error) {
	mtaWorkflow := MTAWorkflow{}

	err := mtaWorkflow.GetPreloadDb(false,false,true).First(&mtaWorkflow, "external_id = ?", externalId).Error
	if err != nil {
		return nil, err
	}
	return &mtaWorkflow, nil
}
func (mtaWorkflow *MTAWorkflow) load() error {
	if mtaWorkflow.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить MTAWorkflow - не указан  Id"}
	}

	err := mtaWorkflow.GetPreloadDb(false,false, true).First(mtaWorkflow,mtaWorkflow.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (mtaWorkflow *MTAWorkflow) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (MTAWorkflow) getList(accountId uint, sortBy string) ([]Entity, int64, error) {
	return MTAWorkflow{}.getPaginationList(accountId, 0, 25, sortBy, "",nil)
}

func (MTAWorkflow) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, int64, error) {

	webHooks := make([]MTAWorkflow,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&MTAWorkflow{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&MTAWorkflow{}).GetPreloadDb(true,false,true).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&MTAWorkflow{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&MTAWorkflow{}).Where("account_id = ?", accountId).Count(&total).Error
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

func (mtaWorkflow *MTAWorkflow) update(input map[string]interface{}) error {
	utils.FixInputHiddenVars(&input)
	return mtaWorkflow.GetPreloadDb(false,false,false).Where("id = ?", mtaWorkflow.Id).Omit("id", "account_id").Updates(input).Error
}

func (mtaWorkflow *MTAWorkflow) delete () error {

	return mtaWorkflow.GetPreloadDb(true,false,false).Where("id = ?", mtaWorkflow.Id).Delete(mtaWorkflow).Error
}
// ######### END CRUD Functions ############

func (mtaWorkflow *MTAWorkflow) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&mtaWorkflow)
	} else {
		_db = _db.Model(&MTAWorkflow{})
	}

	if preload {
		// return _db.Preload("PaymentAmount")
		return _db
	} else {
		return _db
	}
}

// Создание пакета для добавления в буфер на отправку
func (mtaWorkflow *MTAWorkflow) Execute() error {

	// Локальные данные аккаунта, пользователя
	data := make(map[string]interface{})

	account, err := GetAccount(mtaWorkflow.AccountId)
	if err != nil { return err }

	user, err := account.GetUser(mtaWorkflow.UserId)
	if err != nil { return err }

	// Проверяем статус подписки
	if !user.Subscribed {
		if err = mtaWorkflow.delete(); err != nil {
			log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
		}
		return utils.Error{Message: "Невозможно отправить письмо пользователю, т.к. он отписан от всех подписок"}
	}

	// ############ Готовим данные ############
	var sender EmailSender
	switch mtaWorkflow.OwnerType {
		case EmailSenderQueue:
			_e := EmailQueue{}
			err := account.LoadEntity(&_e, mtaWorkflow.OwnerId)
			if err != nil {return err}
			sender = &_e
		case EmailSenderCampaign:
			_e := EmailCampaign{}
			err := account.LoadEntity(&_e, mtaWorkflow.OwnerId)
			if err != nil {return err}
			sender = &_e
		case EmailSenderNotification:
			_e := EmailNotification{}
			err := account.LoadEntity(&_e, mtaWorkflow.OwnerId)
			if err != nil {return err}
			sender = &_e
	default:
		return utils.Error{Message: "Ошибка опознания типа сообщения"}
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
			return utils.Error{ Message: fmt.Sprintf("Невозможно отправить письмо, т.к. объект [id = %v] не запущен\n", sender.GetId())}
		}

		emailQueue, ok := sender.(*EmailQueue)
		if !ok { return errors.New("Ошибка преобразования в Email Queue")}
		
		step, err := emailQueue.GetNearbyActiveStep(mtaWorkflow.QueueExpectedStepId)
		if err != nil {
			// Если нет доступных шагов - удаляем задачу
			if err = mtaWorkflow.delete(); err != nil {
				log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
			}
			return err
		}

		if step.EmailBoxId == nil {
			return utils.Error{Message: "Данные не полные: не хватает emailBoxID"}
		}

		// Только теоретически такое возможно
		if !step.Enabled {
			return utils.Error{Message: "Этап отправки не активен"}
		}

		queueStepId = step.Id
		
		Subject = step.Subject
		PreviewText = step.PreviewText

		// Находим шаблон письма
		_et, err := emailQueue.GetEmailTemplateByStep(mtaWorkflow.QueueExpectedStepId)
		if err != nil {	return err }
		emailTemplate = *_et

		// EmailBox
		err = account.LoadEntity(&emailBox, *step.EmailBoxId)
		if err != nil {
			return err
		}
	}

	if sender.GetType() == EmailSenderNotification {

		if !sender.IsActive() {
			// Возвращаем с ошибкой, т.к. могут копиться люди в очереди
			return utils.Error{ Message: fmt.Sprintf("Невозможно отправить письмо, т.к. объект [id = %v] не запущен\n", sender.GetId()),
			}
		}

		emailNotification, ok := sender.(*EmailNotification)
		if !ok { return errors.New("Ошибка преобразования в email Notification")}

		err := account.LoadEntity(&emailTemplate, *emailNotification.EmailTemplateId)
		if err != nil {	return err }

		err = account.LoadEntity(&emailBox, *emailNotification.EmailBoxId)
		if err != nil {	return err }
		
		Subject = emailNotification.Subject
		PreviewText = emailNotification.PreviewText
	}

	if sender.GetType() == EmailSenderCampaign {

		if !sender.IsActive() {
			return utils.Error{ Message: fmt.Sprintf("Невозможно отправить письмо, т.к. кампания [id = %v] не запущена\n", sender.GetId())}
		}

		emailCampaign, ok := sender.(*EmailCampaign)
		if !ok { return errors.New("Ошибка преобразования в email campaign")}

		err := account.LoadEntity(&emailTemplate, *emailCampaign.EmailTemplateId)
		if err != nil {	return err }

		err = account.LoadEntity(&emailBox, *emailCampaign.EmailBoxId)
		if err != nil {	return err}

		Subject = emailCampaign.Subject
		PreviewText = emailCampaign.PreviewText
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
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается прочитать тему сообщения"}
	}
	if _subject == "" {
		_subject = fmt.Sprintf("Письмо #%v", mtaWorkflow.Id)
	}

	// Готовим еще раз полностью данные
	vData, err := emailTemplate.PrepareViewData(_subject, PreviewText, data, pixelURL, &unsubscribeUrl)
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается подготовить данные для сообщения"}
	}

	var webSite WebSite
	if err = account.LoadEntity(&webSite, emailBox.WebSiteId); err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается загрузить данные по WebSite"}
	}

	var pkg = EmailPkg {
		To: mail.Address{Name: *user.Name, Address: *user.Email},
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
	})

}
