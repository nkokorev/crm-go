package models

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"strings"
	"time"
)

// активная база с реальными задачами по отправке
type MTAWorkflow struct {

	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// email_queues, email_campaigns, email_notifications
	OwnerType	EmailSenderType	`json:"ownerType" gorm:"varchar(32);not null;"` // << тип события: кампания, серия, уведомление
	OwnerId		uint	`json:"ownerId" gorm:"type:smallint;not null;"` // ID типа события: какая серия, компания или уведомление

	// К какой серии писем относится задача
	EmailQueue			EmailQueue `json:"_emailQueue" gorm:"-"`
	EmailNotification	EmailNotification `json:"_emailNotification" gorm:"-"`
	EmailCampaign		EmailCampaign `json:"_emailCampaign" gorm:"-"`

	// Номер необходимого шага в серии EmailQueue. Шаг определяется ситуационно в момент Expected Time. Если шага нет - серия завершается за пользователя.
	// После выполнения - № шага увеличивается на 1
	QueueExpectedStepId	uint	`json:"queueExpectedStepId" gorm:"type:int;not null;"`

	// Запланированное время отправки
	ExpectedTimeStart 	time.Time `json:"expectedTimeStart"`

	// Id пользователя, которому отправляется серия. Обязательный параметр.
	UserId 	uint `json:"userId" gorm:"type:int;not null;"`
	User	User `json:"user"`

	// Число попыток отправки. Если почему-то не удалось отправить письмо есть возможность перенести отправку и повторить попытку. При 2-3х попытках, завершается неудачей.
	NumberOfAttempts uint `json:"numberOfAttempts" gorm:"type:smallint;default:0;"`

	// Когда была последняя попытка отправки (обычно = CreatedAt time)
	LastTriedAt *time.Time  `json:"lastTriedAt"` // << may be null

	CreatedAt time.Time  `json:"createdAt"`
}

func (MTAWorkflow) PgSqlCreate() {
	db.CreateTable(&MTAWorkflow{})
	db.Model(&MTAWorkflow{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&MTAWorkflow{}).AddForeignKey("email_queue_id", "email_queues(id)", "CASCADE", "CASCADE")
	db.Model(&MTAWorkflow{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
}

func (mtaWorkflow *MTAWorkflow) BeforeCreate(scope *gorm.Scope) error {
	mtaWorkflow.Id = 0
	return nil
}
func (mtaWorkflow *MTAWorkflow) AfterCreate(scope *gorm.Scope) error {
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
func (mtaWorkflow *MTAWorkflow) AfterFind() (err error) {
	/*_t, err := time.Parse(time.RFC3339, mtaWorkflow.ExpectedTimeStart.String())
	if err != nil { return err }
	mtaWorkflow.ExpectedTimeStart = _t*/

	// fmt.Println("Найдено: ", _t)
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

func (MTAWorkflow) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return MTAWorkflow{}.getPaginationList(accountId, 0, 25, sortBy, "",nil)
}

func (MTAWorkflow) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	webHooks := make([]MTAWorkflow,0)
	var total uint

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
	input = utils.FixInputHiddenVars(input)
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

// Отправка элемента или перенос
func (mtaWorkflow *MTAWorkflow) Execute() error {

	// Локальные данные аккаунта, пользователя
	data := make(map[string]interface{})

	account, err := GetAccount(mtaWorkflow.AccountId)
	if err != nil {
		if err = mtaWorkflow.delete(); err != nil {
			log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
		}
		return err
	}

	user, err := account.GetUser(mtaWorkflow.UserId)
	if err != nil {
		if err = mtaWorkflow.delete(); err != nil {
			log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
		}
		return err
	}

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

	// wtf
	var QueueOrder uint

	if sender.GetType() == EmailSenderQueue {

		// Проверяем состоянии очереди
		if !sender.IsEnabled() {
			// Тут могут копиться люди в очереди, поэтому возвращаем ошибку
			return utils.Error{ Message: fmt.Sprintf("Невозможно отправить письмо, т.к. объект [id = %v] не запущен\n", sender.GetId()),
			}
		}

		emailQueue, ok := sender.(*EmailQueue)
		if !ok { return errors.New("Ошибка преобразования в EmailQueue")}
		
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
			if err = mtaWorkflow.delete(); err != nil {
				log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
			}
			return utils.Error{Message: "Этап отправки не активен"}
		}

		//////

		Subject = step.Subject
		PreviewText = step.PreviewText
		QueueOrder = step.Order

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

		if !sender.IsEnabled() {
			// Возвращаем с ошибкой, т.к. могут копиться люди в очереди
			return utils.Error{ Message: fmt.Sprintf("Невозможно отправить письмо, т.к. объект [id = %v] не запущен\n", sender.GetId()),
			}
		}

		emailNotification, ok := sender.(*EmailNotification)
		if !ok { return errors.New("Ошибка преобразования в email Notification")}

		err := account.LoadEntity(&emailTemplate, *emailNotification.EmailTemplateId)
		if err != nil {	return err }

		err = account.LoadEntity(&emailBox, *emailNotification.EmailBoxId)
		if err != nil {
			return err
		}
		
		Subject = emailNotification.Subject
		PreviewText = emailNotification.PreviewText
	}

	if sender.GetType() == EmailSenderCampaign {

		if !sender.IsEnabled() {
			// Возвращаем с ошибкой, т.к. могут копиться люди в очереди
			return utils.Error{ Message: fmt.Sprintf("Невозможно отправить письмо, т.к. кампания [id = %v] не запущена\n", sender.GetId()),
			}
		}

		emailCampaign, ok := sender.(*EmailCampaign)
		if !ok { return errors.New("Ошибка преобразования в email campaign")}

		err := account.LoadEntity(&emailTemplate, emailCampaign.EmailTemplateId)
		if err != nil {	return err }

		err = account.LoadEntity(&emailBox, emailCampaign.EmailBoxId)
		if err != nil {
			return err
		}

		Subject = emailCampaign.Subject
		PreviewText = emailCampaign.PreviewText
	}

	// **************************** //
	
	// Объект истории, который может быть дополнен позже
	history := &MTAHistory{
		HashId:  strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true)),
		AccountId: mtaWorkflow.AccountId,
		UserId: &user.Id,
		Email: user.Email,
		OwnerId: sender.GetId(),
		OwnerType: sender.GetType(),
		EmailTemplateId: utils.UINTp(emailTemplate.Id),
		QueueStepId: utils.UINTp(QueueOrder),
		QueueCompleted: false,	// по умолчанию
		NumberOfAttempts: mtaWorkflow.NumberOfAttempts + 1,
		Succeed: false, 	// по умолчанию
	}
	defer func() {
		_, _ = history.create()
	}()
	
	unsubscribeUrl := account.GetUnsubscribeUrl(*user, *history)
	pixelURL := account.GetPixelUrl(*history)

	// Подготавливаем данные для письма, чтобы можно было их использовать в шаблоне
	data["accountId"] = mtaWorkflow.AccountId
	data["Account"] = account.GetDepersonalizedData() // << хз
	data["userId"] = user.Id // << хз
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

	err = emailTemplate.SendMail(emailBox, user.Email, _subject, vData,  unsubscribeUrl)
	if err != nil {
		
		history.Succeed = false

		// Обновляем данные последней попытки
		timeNow := time.Now().UTC()
		if err2 := mtaWorkflow.update(map[string]interface{}{
			"last_tried_at": &timeNow,
			"number_of_attempts": mtaWorkflow.NumberOfAttempts + 1,
			"expected_time_start": time.Now().UTC().Add(time.Minute * 5),
		}); err2 != nil {
			fmt.Println("Err mtaWorkflow update: ", err2)
		}

		return err
	} else {

		// Ставим флаг успешного выполнения
		history.Succeed = true

		if sender.GetType() == EmailSenderQueue {
			// 1. Получаем следующий шаг
			emailQueue, ok := sender.(*EmailQueue)
			if !ok {
				log.Println("Ошибка преобразования в EmailQueue")
				history.QueueCompleted = true
				if err = mtaWorkflow.delete(); err != nil {
					log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
				}
				return nil
			}

			nextStep, err := emailQueue.GetNextActiveStep(QueueOrder)
			if err != nil {
				history.QueueCompleted = true
				// исключаем задачу, если не удалось ее обновить
				if err = mtaWorkflow.delete(); err != nil {
					log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
				}
				return nil
			}

			// Проверяем на его активность
			if !nextStep.Enabled {
				history.QueueCompleted = true
				if err = mtaWorkflow.delete(); err != nil {
					log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
				}
				return nil
			}

			// Обновляем задачу, вместо удаления / создания новой
			if err := mtaWorkflow.UpdateByNextStep(*nextStep); err != nil {

				// исключаем задачу, если не удалось ее обновить
				if err = mtaWorkflow.delete(); err != nil {
					log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
				}
				
				// Не удалось обновить следующий шаг, значит последний был до этого
				history.QueueCompleted = true
				return nil
			}

			// Если дошли сюда - значит, задача не завершена
			history.QueueCompleted = false

			return nil
		}

		if sender.GetType() == EmailSenderCampaign || sender.GetType() == EmailSenderNotification{
			if err = mtaWorkflow.delete(); err != nil {
				log.Printf("Невозможно исключить задачу [id = %v] по отпрваке: %v\n", mtaWorkflow.Id, err)
			}
			return nil
		}
		
	}

	return nil
}

func (mtaWorkflow *MTAWorkflow) UpdateByNextStep(expectedStep EmailQueueEmailTemplate) error {
	// Изменяется expected step, expected_time_start, number_attems, last_tried
	return mtaWorkflow.update(map[string]interface{}{
		"queue_expected_step_id": expectedStep.Order,
		"expected_time_start": time.Now().UTC().Add(expectedStep.DelayTime),
		"number_of_attempts": 0,
		"last_tried_at": nil,
	})

}
