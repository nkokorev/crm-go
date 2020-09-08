package models

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/mail"
	"reflect"
	"strings"
	"time"
)

type EmailNotification struct {
	Id     			uint   	`json:"id" gorm:"primaryKey"`
	PublicId		uint   	`json:"public_id" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	Status 			WorkStatus `json:"status" gorm:"type:varchar(18);default:'pending'"`
	FailedStatus	*string 		`json:"failed_status" gorm:"type:varchar(255);"`

	// Delay			uint 	`json:"delay" gorm:"type:int;default:0"` // Задержка перед отправлением в минутах: [0-180]
	DelayTime		time.Duration `json:"delay_time" gorm:"type:int8;default:0"`// << учитывается только время [0-24]
	
	Name 			string 	`json:"name" gorm:"type:varchar(128);default:''"` // "Оповещение менеджера", "Оповещение клиента"

	Subject			*string 	`json:"subject" gorm:"type:varchar(255);"` // Тема сообщения, компилируются
	PreviewText		*string 	`json:"preview_text" gorm:"type:varchar(255);"` // Тема сообщения, компилируются

	EmailTemplateId *uint 	`json:"email_template_id" gorm:"type:int;"`
	EmailTemplate 	EmailTemplate 	`json:"email_template"`

	// EmailTemplateId *uint 	`json:"email_template_id" gorm:"type:int;"` // всегда должен быть шаблон, иначе смысла в нем нет
	// EmailTemplate 	EmailTemplate 	`json:"email_template" gorm:"preload:true"`

	EmailBoxId		*uint 	`json:"email_box_id" gorm:"type:int;"` // С какого ящика идет отправка
	EmailBox		EmailBox `json:"email_box"`
	// =============   Настройки получателей    ===================

	// Список пользователей позволяет сделать "рассылку" уведомления по email-адреса пользователей, до 10 человек.
	// SendingToUsers		bool			`json:"sendingToUsers" gorm:"type:bool;default:false"` // Отправлять пользователем RatusCRM (на их почтовые адреса, при их наличии)
	RecipientUsersList	datatypes.JSON	`json:"recipient_users_list"` // список id пользователей, которые получат уведомление

	// Динамический список пользователей
	ParseRecipientUser		bool	`json:"parse_recipient_user" gorm:"type:bool;default:false"` // Спарсить из контекста пользователя(ей) по userId / users: ['email@mail.ru']
	ParseRecipientCustomer	bool	`json:"parse_recipient_customer" gorm:"type:bool;default:false"` // Спарсить из контекста пользователя по customerId / users: ['email@mail.ru']
	ParseRecipientManager	bool	`json:"parse_recipient_manager" gorm:"type:bool;default:false"` // Спарсить из контекста пользователя по customerId / users: ['email@mail.ru']

	// ==========================================

	// Скрытый список пользователей для Data и фронтенда
	RecipientUsers []User	`json:"_recipient_users" gorm:"-"`

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

// ############# Entity interface #############
func (emailNotification EmailNotification) GetId() uint { return emailNotification.Id }
func (emailNotification *EmailNotification) setId(id uint) { emailNotification.Id = id }
func (emailNotification *EmailNotification) setPublicId(publicId uint) { emailNotification.PublicId = publicId }
func (emailNotification EmailNotification) GetAccountId() uint { return emailNotification.AccountId }
func (emailNotification *EmailNotification) setAccountId(id uint) { emailNotification.AccountId = id }
func (EmailNotification) SystemEntity() bool { return false }
func (EmailNotification) GetType() string { return "email_notifications" }
// статус для обхода воркера
func (emailNotification EmailNotification) IsEnabled() bool {

	if emailNotification.Status == WorkStatusPending || emailNotification.Status == WorkStatusFailed || emailNotification.Status == WorkStatusCancelled {
		return false
	}

	return true
}
func (emailNotification EmailNotification) IsActive() bool {
	return emailNotification.Status == WorkStatusActive
}

// ############# Entity interface #############

func (EmailNotification) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&EmailNotification{}); err != nil {log.Fatal(err)}
	// db.Model(&EmailNotification{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&EmailNotification{}).AddForeignKey("email_template_id", "email_templates(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE email_notifications " +
		"ADD CONSTRAINT email_notifications_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT email_notifications_email_template_id_fkey FOREIGN KEY (email_template_id) REFERENCES email_templates(id) ON DELETE SET NULL ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (emailNotification *EmailNotification) BeforeCreate(tx *gorm.DB) error {
	emailNotification.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&EmailNotification{}).Where("account_id = ?",  emailNotification.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	emailNotification.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (emailNotification *EmailNotification) AfterFind(tx *gorm.DB) (err error) {

	// Собираем пользователей
	var _users = make([]User,0)
	if emailNotification.RecipientUsersList != nil {
		var arr []uint
		err = json.Unmarshal(emailNotification.RecipientUsersList, &arr)
		if err != nil { return err }

		if len(arr) > 0 {
			err = db.Find(&emailNotification.RecipientUsers, "id IN (?)", arr).Error
			if err != nil  {return err}
		} else {
			emailNotification.RecipientUsers = _users
		}

	}  else {
		emailNotification.RecipientUsers = _users
	}


	// fix nono to ms
	// emailNotification.DelayTime = time.Millisecond*emailNotification.DelayTime


	/////////////////////////////////////

	if reflect.DeepEqual(emailNotification.RecipientUsersList, *new(postgres.Jsonb)) {
		emailNotification.RecipientUsersList = []byte("[]")
	}


	return nil
}

// ######### CRUD Functions ############
func (emailNotification EmailNotification) create() (Entity, error)  {

	_item := emailNotification
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (EmailNotification) get(id uint, preloads []string) (Entity, error) {

	var item EmailNotification

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (emailNotification *EmailNotification) load(preloads []string) error {
	if emailNotification.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailNotification - не указан  Id"}
	}

	err := emailNotification.GetPreloadDb(false, false, preloads).First(emailNotification, emailNotification.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailNotification *EmailNotification) loadByPublicId(preloads []string) error {
	
	if emailNotification.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailNotification - не указан  Id"}
	}
	if err := emailNotification.GetPreloadDb(false,false, preloads).First(emailNotification, "account_id = ? AND public_id = ?", emailNotification.AccountId, emailNotification.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (EmailNotification) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return EmailNotification{}.getPaginationList(accountId, 0, 25, sortBy, "",nil,preload)
}
func (EmailNotification) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	emailNotifications := make([]EmailNotification,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		// jsearch := search
		search = "%"+search+"%"

		err := (&EmailNotification{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailNotifications, "name ILIKE ? OR description ILIKE ?", search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailNotification{}).GetPreloadDb(false,false,nil).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&EmailNotification{}).GetPreloadDb(false,false,preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailNotifications).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&EmailNotification{}).GetPreloadDb(false,false,nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailNotifications))
	for i := range emailNotifications {
		entities[i] = &emailNotifications[i]
	}

	return entities, total, nil
}
func (emailNotification *EmailNotification) update(input map[string]interface{}, preloads []string) error {

	// Приводим в опрядок
	input = utils.FixJSONB_Uint(input, []string{"recipient_users_list"})

	delete(input, "email_template")
	delete(input, "email_box")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","email_template_id", "email_box_id","delay_time"}); err != nil {
		return err
	}

	if err := (&EmailNotification{}).GetPreloadDb(false,false,nil).Where(" id = ?", emailNotification.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := emailNotification.GetPreloadDb(false,false,preloads).First(emailNotification, emailNotification.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (emailNotification *EmailNotification) delete () error {
	return emailNotification.GetPreloadDb(true,false,nil).Where("id = ?", emailNotification.Id).Delete(emailNotification).Error
}
// ######### END CRUD Functions ############

func (emailNotification *EmailNotification) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&emailNotification)
	} else {
		_db = _db.Model(&EmailNotification{})
	}

	if autoPreload {
		return _db.Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
			return db.Select(EmailTemplate{}.SelectArrayWithoutData())
		}).Preload("EmailBox")
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"EmailBox","EmailTemplate"})

		for _,v := range allowed {
			if v == "EmailTemplate" {
				_db.Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
					return db.Select(EmailTemplate{}.SelectArrayWithoutData())
				})
			} else {
				_db.Preload(v)
			}
		}
		return _db
	}

}

// Вызов уведомления
func (emailNotification EmailNotification) Execute(data map[string]interface{}) error {

	// Проверяем статус уведомления
	if !emailNotification.IsActive() {
		return utils.Error{Message: fmt.Sprintf("Уведомление не может быть отправлено т.к. находится в статусе - %v\n",emailNotification.Status)}
	}

	// Проверяем возможность отправки. Может излишне, но при небольшой нагрузке - ок, об ошибках узнаем До отправки
	if err := emailNotification.Validate(); err != nil {
		return err
	}

	// Get Account
	account, err := GetAccount(emailNotification.AccountId)
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается найти аккаунт"}
	}

	// Находим шаблон письма
	var emailTemplate EmailTemplate
	if err = account.LoadEntity(&emailTemplate, *emailNotification.EmailTemplateId,nil); err != nil {
		log.Printf("Ошибка отправления Уведомления - не удается загрузить шаблон письма: %v\n", err)
		return err
	}

	// Проверяем, чтобы был почтовые ящики, с которого отправляем
	if emailNotification.EmailBox.Id < 1 {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается получить почтовый ящик"}
	}

	// Загружаем данные почтового ящика
	err = emailNotification.EmailBox.load(nil)
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается загрузить данные WEbSite"}
	}

	// NEW NEW ============

	var users = make([]User,0)

	// 1. Собираем список пользователей
	for i := range emailNotification.RecipientUsers {
		users = append(users, emailNotification.RecipientUsers[i])

	}
	if emailNotification.ParseRecipientUser {
		if userSTR, ok := data["userId"]; ok {
			if userId, ok := userSTR.(uint); ok {
				user, err := account.GetUser(userId)
				if err == nil && user.Email != nil {
					users = append(users, *user)
				}
			}
		}
	}
	if emailNotification.ParseRecipientCustomer {
		if customerSTR, ok := data["customerId"]; ok {
			if customerId, ok := customerSTR.(uint); ok {
				customer, err := account.GetUser(customerId)
				if err == nil && customer.Email != nil {
					users = append(users, *customer)
				}
			}
		}
	}
	if emailNotification.ParseRecipientManager {
		if managerSTR, ok := data["managerId"]; ok {
			if managerId, ok := managerSTR.(uint); ok {
				manager, err := account.GetUser(managerId)
				if err == nil && *manager.Email != "" {
					users = append(users, *manager)
				}
			}
		}
	}

	// 2. Отправляем списку
	for i := range users {

		historyHashId := strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))

		unsubscribeUrl := account.GetUnsubscribeUrl(users[i].HashId, historyHashId)
		pixelURL := account.GetPixelUrl(historyHashId)

		// Компилируем тему письма
		if emailNotification.Subject == nil || emailNotification.PreviewText == nil {
			break
		}
		_subject, err := parseSubjectByData(*emailNotification.Subject, data)
		if err != nil {
			log.Printf("Ошибка отправления Уведомления - не удается прочитать тему сообщения. emailNotificationId: %v\n", emailNotification.Id)
			continue
		}
		if _subject == "" {
			_subject = fmt.Sprintf("Уведомление по почте #%v", emailNotification.Id)
		}

		vData, err := emailTemplate.PrepareViewData(_subject, *emailNotification.PreviewText, data, pixelURL, &unsubscribeUrl)
		if err != nil {
			log.Printf("Ошибка отправления Уведомления - не удается подготовить данные для сообщения. emailNotificationId: %v\n", emailNotification.Id)
			continue
		}

		var webSite WebSite
		if err = account.LoadEntity(&webSite, emailNotification.EmailBox.WebSiteId,nil); err != nil {
			log.Printf("Ошибка отправления Уведомления - не удается загрузить данные по WebSite: %v\n", err)
			continue
		}

		// Пакет для добавления в MTA-сервер
		var pkg = EmailPkg {
			To: mail.Address{Name: *users[i].Name, Address: *users[i].Email},
			accountId: account.Id,
			userId: users[i].Id,
			workflowId: 0, // мы не знаем
			webSite: &webSite,
			emailBox: &emailNotification.EmailBox,
			emailSender: &emailNotification,
			emailTemplate: &emailTemplate,
			subject: _subject,
			viewData: vData,
		}

		if emailNotification.DelayTime == 0 {
			fmt.Println("Отправляем сейчас же!")
			SendEmail(pkg)
		} else {
			// ставим в очередь с записью в БД

			// 2. Add user to MTAWorkflow
			mtaWorkflow := MTAWorkflow{
				AccountId: emailNotification.AccountId,
				OwnerId: emailNotification.Id,
				OwnerType: EmailSenderNotification,
				ExpectedTimeStart: time.Now().UTC().Add(emailNotification.DelayTime),
				UserId: users[i].Id,
				NumberOfAttempts: 0,
			}

			if _, err := mtaWorkflow.create(); err != nil {
				return err
			}

			return nil
		}


		// _, _ = history.create()
	}

	return nil
}

func parseSubjectByData(tpl string, data map[string]interface{}) (string, error) {

	body := new(bytes.Buffer)

	tmpl, err := template.New("et.Name").Parse(tpl)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(body, data)
	if err != nil {
		return "", utils.Error{Message: fmt.Sprintf("Ошибка в заголовке шаблона")}
	}

	return body.String(), nil
}

// Прямая функция для обновления, без проверок
func (emailNotification *EmailNotification) ChangeWorkStatus(status WorkStatus, reason... string) error {

	switch status {
	case WorkStatusPending:
		return emailNotification.SetPendingStatus()
	case WorkStatusPlanned:
		return emailNotification.SetPlannedStatus()
	case WorkStatusActive:
		return emailNotification.SetActiveStatus()
	case WorkStatusPaused:
		return emailNotification.SetPausedStatus()
	case WorkStatusFailed:
		_reason := ""
		if len(reason) > 0 {
			_reason = reason[0]
		}
		return emailNotification.SetFailedStatus(_reason)
	case WorkStatusCompleted:
		return emailNotification.SetCompletedStatus()
	case WorkStatusCancelled:
		return emailNotification.SetCancelledStatus()
	default:
		return utils.Error{Message: "Статус не опознан"}
	}

}

func (emailNotification *EmailNotification) updateWorkStatus(status WorkStatus, reason... string) error {
	_reason := ""
	if len(reason) > 0 {
		_reason = reason[0]
	}

	return emailNotification.update(map[string]interface{}{
		"status":	status,
		"failed_status": _reason,
	},nil)
}

// Функция не должна вызываться, т.к. статуса planned у EmailNotification не используется
func (emailNotification *EmailNotification) SetPendingStatus() error {

	// Возможен вызов из состояния planned: вернуть на доработку => pending
	if emailNotification.Status != WorkStatusPlanned {
		reason := "Невозможно установить статус,"
		switch emailNotification.Status {
		case WorkStatusPending:
			reason += "т.к. уведомление уже в разработке"
		case WorkStatusActive:
			reason += "т.к. уведомление в процессе работы"
		case WorkStatusPaused:
			reason += "т.к. уведомление на паузе, но в процессе работы"
		case WorkStatusFailed:
			reason += "т.к. уведомление завершено с ошибкой"
		case WorkStatusCompleted:
			reason += "т.к. уведомление уже завершено"
		case WorkStatusCancelled:
			reason += "т.к. уведомление отменено"
		}
		return utils.Error{Message: reason}
	}

	// fix: удаляем задачу в TaskScheduler
	if err := emailNotification.RemoveRunTask(); err != nil {
		return err
	}

	return emailNotification.updateWorkStatus(WorkStatusPending)
}

// Функция не должна вызываться, т.к. статус planned у EmailNotification не используется
func (emailNotification *EmailNotification) SetPlannedStatus() error {

	return utils.Error{Message: "Отложенный запуск уведомлений не предусмотрен"}

	// Возможен вызов из состояния pending: запланировать кампанию => planned
	if emailNotification.Status != WorkStatusPending  {
		reason := "Невозможно запланировать кампанию,"
		switch emailNotification.Status {
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
	account, err := GetAccount(emailNotification.AccountId)
	if err != nil {
		return utils.Error{Message: "Не удается найти аккаунт"}
	}

	// Проверяем кампанию и шаблон, чтобы не ставить в план не рабочую кампанию.
	if err := emailNotification.Validate(); err != nil { return err  }

	// На всякий случай удаляем задачу по запуску этой кампании, чтобы не было дубля
	if err := emailNotification.RemoveRunTask(); err != nil {
		return err
	}

	// Создаем новый объект task для добавлении кампании в TaskScheduler
	task := TaskScheduler {
		AccountId: emailNotification.AccountId,
		OwnerType: TaskEmailCampaignRun,
		OwnerId: emailNotification.Id,
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
	return emailNotification.updateWorkStatus(WorkStatusPlanned)
}
func (emailNotification *EmailNotification) SetActiveStatus() error {

	// Возможен вызов из состояния planned или paused: запустить кампанию => active
	if emailNotification.Status != WorkStatusCompleted && emailNotification.Status != WorkStatusPending && emailNotification.Status != WorkStatusPlanned && emailNotification.Status != WorkStatusPaused {
		reason := "Невозможно запустить уведомление,"
		switch emailNotification.Status {
		case WorkStatusPending:
			reason += "т.к. уведомление еще в стадии разработки"
		case WorkStatusActive:
			reason += "т.к. уведомление уже в процессе работы"
		case WorkStatusCompleted:
			reason += "т.к. уведомление уже завершено"
		case WorkStatusFailed:
			reason += "т.к. уведомление завершено с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. уведомление отменено"
		}
		return utils.Error{Message: reason}
	}

	// Снова проверяем кампанию и шаблон
	if err := emailNotification.Validate(); err != nil { return err  }

	// Переводим в состояние "Активна", т.к. все проверки пройдены и можно продолжить ее выполнение
	return emailNotification.updateWorkStatus(WorkStatusActive)
}
func (emailNotification *EmailNotification) SetPausedStatus() error {

	// Возможен вызов из состояния active: приостановить кампанию => paused
	if emailNotification.Status != WorkStatusActive {
		reason := "Невозможно приостановить уведомление,"
		switch emailNotification.Status {
		case WorkStatusPending:
			reason += "т.к. уведомление еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. уведомление уже в стадии планирования"
		case WorkStatusPaused:
			reason += "т.к. уведомление уже приостановлено"
		case WorkStatusCompleted:
			reason += "т.к. уведомление уже завершено"
		case WorkStatusFailed:
			reason += "т.к. уведомление завершено с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. уведомление отменено"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Приостановлена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailNotification.updateWorkStatus(WorkStatusPaused)
}
func (emailNotification *EmailNotification) SetCompletedStatus() error {

	// Возможен вызов из состояния active, paused: завершить кампанию => completed
	// Сбрасываются все задачи из очереди
	if emailNotification.Status != WorkStatusActive && emailNotification.Status != WorkStatusPaused {
		reason := "Невозможно завершить уведомление,"
		switch emailNotification.Status {
		case WorkStatusPending:
			reason += "т.к. уведомление еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. уведомление еще в стадии планирования"
		case WorkStatusCompleted:
			reason += "т.к. уведомление уже завершено"
		case WorkStatusFailed:
			reason += "т.к. уведомление завершено с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. уведомление отменено"
		}
		return utils.Error{Message: reason}
	}

	// Удаляем все задачи из WorkFlow
	err := db.Where("owner_id = ? AND owner_type = ?", emailNotification.Id, emailNotification.GetType()).Delete(&MTAWorkflow{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из MTAWorkflow: %v\b", err)
		return utils.Error{Message: "Ошибка завершения кампании - невозможно удалить остаток задач"}
	}

	// Переводим в состояние "Завершено", т.к. все проверки пройдены и можно приостановить уведомление
	return emailNotification.updateWorkStatus(WorkStatusCompleted)
}
func (emailNotification *EmailNotification) SetFailedStatus(reason string) error {
	
	// Удаляем все задачи из WorkFlow
	err := db.Where("owner_id = ? AND owner_type = ?", emailNotification.Id, emailNotification.GetType()).Delete(MTAWorkflow{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из MTAWorkflow: %v\b", err)
		return utils.Error{Message: "Ошибка завершения кампании - невозможно удалить остаток задач"}
	}

	// На всякий случай удаляем задачу по запуску этого уведомления, чтобы не было дубля (прошлые не удаляются)
	if err := emailNotification.RemoveRunTask(); err != nil {
		return err
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailNotification.updateWorkStatus(WorkStatusFailed, reason)
}
func (emailNotification *EmailNotification) SetCancelledStatus() error {

	// Возможен вызов из состояния active, paused, planned: завершить кампанию => cancelled
	if emailNotification.Status != WorkStatusActive && emailNotification.Status != WorkStatusPaused && emailNotification.Status != WorkStatusPlanned {
		reason := "Невозможно отменить уведомление,"
		switch emailNotification.Status {
		case WorkStatusPending:
			reason += "т.к. уведомление еще в стадии разработки"
		case WorkStatusCompleted:
			reason += "т.к. уведомление уже завершено"
		case WorkStatusFailed:
			reason += "т.к. уведомление завершено с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. уведомление уже отменено"
		}
		return utils.Error{Message: reason}
	}

	// Удаляем все задачи из WorkFlow
	err := db.Where("owner_id = ? AND owner_type = ?", emailNotification.Id, emailNotification.GetType()).Delete(MTAWorkflow{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из MTAWorkflow: %v\b", err)
		return utils.Error{Message: "Ошибка завершения кампании - невозможно удалить остаток задач"}
	}

	// На всякий случай удаляем задачу по запуску этой кампании, чтобы не было дубля
	if err := emailNotification.RemoveRunTask(); err != nil {
		return err
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return emailNotification.updateWorkStatus(WorkStatusCancelled)
}

// Проверяет возможность отправки email-уведомления
func (emailNotification *EmailNotification) Validate() error {

	account, err := GetAccount(emailNotification.AccountId)
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается найти аккаунт"}
	}


	// Проверяем тело сообщения (не должно быть пустое)
	if emailNotification.Subject == nil {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет темы сообщения"}
	}
	if emailNotification.PreviewText == nil {
		emailNotification.PreviewText = utils.STRp("")
	}

	// Проверяем ключи и загружаем еще раз все данные для отправки сообщения
	if emailNotification.EmailTemplateId == nil {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет установленного шаблона email-сообщения"}
	}
	err = account.LoadEntity(&emailNotification.EmailTemplate, *emailNotification.EmailTemplateId,nil)
	if err != nil {
		log.Printf("Ошибка загрузки шаблона email-сообщения для кампании [%v]: %v\n", emailNotification.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки шаблона email-сообщения"}
	}

	if emailNotification.EmailBoxId == nil  {
		return utils.Error{Message: "Кампания не может быть запущена, т.к. нет установленного адреса отправителя"}
	}
	err = account.LoadEntity(&emailNotification.EmailBox, *emailNotification.EmailBoxId,nil)
	if err != nil {
		log.Printf("Ошибка загрузки адреса отправителя для кампании [%v]: %v\n", emailNotification.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки адреса отправителя"}
	}

	// Проверяем DKIM подпись
	var webSite WebSite
	err = account.LoadEntity(&webSite, emailNotification.EmailBox.WebSiteId,nil)
	if err != nil {
		log.Printf("Ошибка загрузки web site отправителя для кампании [%v]: %v\n", emailNotification.Id, err)
		return utils.Error{Message: "Кампания не может быть запущена - ошибка загрузки адреса отправителя"}
	}

	if err := webSite.ValidateDKIM(); err != nil { return err }

	// Тестовые данные
	// Локальные данные аккаунта, пользователя
	data := make(map[string]interface{})

	data["accountId"] = account.Id
	data["Account"] = account.GetDepersonalizedData() // << хз
	data["userId"] = 0
	data["User"] = User{Id: 1,Name: utils.STRp("TIvan"), Username: utils.STRp("TUsername"), Surname: utils.STRp("TSurname"), Patronymic: utils.STRp("TPatronymic"),
		PhoneRegion: utils.STRp("RU"), Phone: utils.STRp("+79251000000")} // << хз
	data["unsubscribeUrl"] = "/unsubscribe_url"



	viewData := ViewData {
		Subject: *emailNotification.Subject,
		PreviewText: *emailNotification.PreviewText,
		Data: data,
		Json: data,
		UnsubscribeURL: "",
		PixelURL: "",
		PixelHTML: "<div></div>",
	}

	return emailNotification.EmailTemplate.Validate(&viewData)
}
// Удаляет связанную задачу по запуск
func (emailNotification *EmailNotification) RemoveRunTask() error {

	// Удаляем все задачи, которые можно "выполнить" в будущем
	err := db.Where("owner_id = ? AND owner_type = ? AND (status != ? OR status != ? OR status != ?)",
		emailNotification.Id, TaskEmailNotificationRun, WorkStatusCompleted, WorkStatusCancelled, WorkStatusFailed).Delete(TaskScheduler{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Ошибка удаления очереди из TaskScheduler: %v\b", err)
		return utils.Error{Message: "Ошибка удаления задачи по запуску кампании"}
	}

	return nil
}