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
	FailedStatus	string 		`json:"failed_status" gorm:"type:varchar(255);"`

	// Delay			uint 	`json:"delay" gorm:"type:int;default:0"` // Задержка перед отправлением в минутах: [0-180]
	DelayTime		time.Duration `json:"delay_time" gorm:"type:int8;default:0"`// << учитывается только время [0-24]
	
	Name 			string 	`json:"name" gorm:"type:varchar(128);default:''"` // "Оповещение менеджера", "Оповещение клиента"

	Subject			string 	`json:"subject" gorm:"type:varchar(128);not null;"` // Тема сообщения, компилируются
	PreviewText		string 	`json:"preview_text" gorm:"type:varchar(255);default:''"` // Тема сообщения, компилируются

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
	ParseRecipientUser	bool	`json:"parse_recipient_user" gorm:"type:bool;default:false"` // Спарсить из контекста пользователя(ей) по userId / users: ['email@mail.ru']
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

	if err := _item.GetPreloadDb(false,true, nil).First(&_item,_item.Id).Error; err != nil {
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

	// Проверяем тело сообщения (не должно быть пустое)
	if emailNotification.Subject == "" {
		return utils.Error{Message: "Уведомление не может быть отправлено т.к. нет темы сообщения"}
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
		_subject, err := parseSubjectByData(emailNotification.Subject, data)
		if err != nil {
			log.Printf("Ошибка отправления Уведомления - не удается прочитать тему сообщения. emailNotificationId: %v\n", emailNotification.Id)
			continue
		}
		if _subject == "" {
			_subject = fmt.Sprintf("Уведомление по почте #%v", emailNotification.Id)
		}

		vData, err := emailTemplate.PrepareViewData(_subject, emailNotification.PreviewText, data, pixelURL, &unsubscribeUrl)
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

func (emailNotification *EmailNotification) changeWorkStatus(status WorkStatus, reason... string) error {
	_reason := "" // обнуление
	if len(reason) > 0 {
		_reason = reason[0]
	}
	return emailNotification.update(map[string]interface{}{
		"status":	status,
		"failed_status": _reason,
	},nil)
}
