package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"html/template"
	"log"
	"reflect"
	"strings"
	"time"
)

type EmailNotification struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	Enabled 		bool 	`json:"enabled" gorm:"type:bool;default:false;"` // отключить / включить
	Delay			uint 	`json:"delay" gorm:"type:int;default:0"` // Задержка перед отправлением в минутах: [0-180]
	
	Name 			string 	`json:"name" gorm:"type:varchar(128);default:''"` // "Оповещение менеджера", "Оповещение клиента"

	Subject			string 	`json:"subject" gorm:"type:varchar(128);not null;"` // Тема сообщения, компилируются
	PreviewText		string 	`json:"previewText" gorm:"type:varchar(255);default:''"` // Тема сообщения, компилируются
	
	Description		string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание что к чему)

	EmailTemplateId *uint 	`json:"emailTemplateId" gorm:"type:int;default:null;"` // всегда должен быть шаблон, иначе смысла в нем нет
	EmailTemplate 	EmailTemplate 	`json:"emailTemplate" gorm:"preload:true"`

	EmailBoxId		*uint 	`json:"emailBoxId" gorm:"type:int;default:null;"` // С какого ящика идет отправка
	EmailBox		EmailBox `json:"emailBox" gorm:"preload:false"`
	// =============   Настройки получателей    ===================

	// Список пользователей позволяет сделать "рассылку" уведомления по email-адреса пользователей, до 10 человек.
	SendingToUsers		bool			`json:"sendingToUsers" gorm:"type:bool;default:false"` // Отправлять пользователем RatusCRM (на их почтовые адреса, при их наличии)
	RecipientUsersList	postgres.Jsonb 	`json:"recipientUsersList" gorm:"type:JSONB;DEFAULT '{}'::JSONB"` // список id пользователей, которые получат уведомление

	// Список фиксированных адресов позволяет сделать "рассылку" по своей базе, до 10 человек.
	SendingToFixedAddresses	bool	`json:"sendingToFixedAddresses" gorm:"type:bool;default:false"` // Отправлять ли на фиксированный адреса
	RecipientList			postgres.Jsonb	`json:"recipientList" gorm:"type:JSONB;DEFAULT '{}'::JSONB"` // фиксированный список адресов, на которые будет произведено уведомление
	
	// Динамический список пользователей
	ParseRecipientUser	bool	`json:"parseRecipientUser" gorm:"type:bool;default:false"` // Спарсить из контекста пользователя(ей) по userId / users: ['email@mail.ru']
	ParseRecipientCustomer	bool	`json:"parseRecipientCustomer" gorm:"type:bool;default:false"` // Спарсить из контекста пользователя по customerId / users: ['email@mail.ru']
	ParseRecipientManager	bool	`json:"parseRecipientManager" gorm:"type:bool;default:false"` // Спарсить из контекста пользователя по customerId / users: ['email@mail.ru']

	// ==========================================

	// Скрытый список пользователей для Data и фронтенда
	RecipientUsers []User	`json:"_recipientUsers" gorm:"-"`

	CreatedAt 		time.Time `json:"createdAt"`
	UpdatedAt 		time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (emailNotification EmailNotification) GetId() uint { return emailNotification.Id }
func (emailNotification *EmailNotification) setId(id uint) { emailNotification.Id = id }
func (emailNotification *EmailNotification) setPublicId(id uint) { }
func (emailNotification EmailNotification) GetAccountId() uint { return emailNotification.AccountId }
func (emailNotification *EmailNotification) setAccountId(id uint) { emailNotification.AccountId = id }
func (EmailNotification) SystemEntity() bool { return false }

// ############# Entity interface #############

func (EmailNotification) PgSqlCreate() {
	db.CreateTable(&EmailNotification{})
	db.Model(&EmailNotification{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&EmailNotification{}).AddForeignKey("email_template_id", "email_templates(id)", "CASCADE", "CASCADE")
}
func (emailNotification *EmailNotification) BeforeCreate(scope *gorm.Scope) error {
	emailNotification.Id = 0
	return nil
}
func (emailNotification *EmailNotification) AfterFind() (err error) {

	// Собираем пользователей
	b, err := emailNotification.RecipientUsersList.MarshalJSON()
	if err != nil {
		return err
	}

	var arr []uint
	err = json.Unmarshal(b, &arr)
	if err != nil { return err }

	err = db.Find(&emailNotification.RecipientUsers, "id IN (?)", arr).Error
	if err != nil  {return err}

	//////


	/////////////////////////////////////

	
	if reflect.DeepEqual(emailNotification.SendingToFixedAddresses, *new(postgres.Jsonb)) {
		emailNotification.RecipientUsersList = postgres.Jsonb{RawMessage: []byte("[]")}
	}

	if reflect.DeepEqual(emailNotification.RecipientUsersList, *new(postgres.Jsonb)) {
		emailNotification.RecipientUsersList = postgres.Jsonb{RawMessage: []byte("[]")}
	}


	return nil
}

// ######### CRUD Functions ############
func (emailNotification EmailNotification) create() (Entity, error)  {

	en := emailNotification

	if err := db.Create(&en).Error; err != nil {
		return nil, err
	}

	var newItem Entity = &en

	return newItem, nil
}

func (EmailNotification) get(id uint) (Entity, error) {

	var emailNotification EmailNotification

	err := db.Preload("EmailBox").First(&emailNotification, id).Error
	if err != nil {
		return nil, err
	}
	return &emailNotification, nil
}
func (emailNotification *EmailNotification) load() error {

	err := db.Preload("EmailBox").First(emailNotification, emailNotification.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*EmailNotification) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (EmailNotification) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return EmailNotification{}.getPaginationList(accountId, 0, 100, sortBy, "",nil)
}
func (EmailNotification) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	emailNotifications := make([]EmailNotification,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		// jsearch := search
		search = "%"+search+"%"

		err := db.Model(&EmailNotification{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&emailNotifications, "name ILIKE ? OR description ILIKE ?", search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailNotification{}).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&EmailNotification{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&emailNotifications).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailNotification{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailNotifications))
	for i,_ := range emailNotifications {
		entities[i] = &emailNotifications[i]
	}

	return entities, total, nil
}

func (emailNotification *EmailNotification) update(input map[string]interface{}) error {

	// Приводим в опрядок
	input = utils.FixJSONB_String(input, []string{"recipientList"})
	input = utils.FixJSONB_Uint(input, []string{"recipientUsersList"})

	// delete(input, "recipientList")
	// delete(input, "recipientUsersList")
	delete(input, "emailTemplate")
	delete(input, "emailBox")


	// if err := db.Model(EmailNotification{}).Where("id = ?", emailNotification.Id).Omit("id", "account_id").Updates(input).Error; err != nil {
	// 	return err
	// }
	// fmt.Println("emailNotification id", emailNotification.EmailBoxId)

/*	if err := db.Model(emailNotification).
		Omit("id", "account_id","created_at").Update(input).Error; err != nil {
		return err
	}*/

	// work!!!
	if err := db.Set("gorm:association_autoupdate", false).Model(EmailNotification{}).Where(" id = ?", emailNotification.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := db.Preload("EmailBox").Preload("EmailTemplate").First(emailNotification, emailNotification.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (emailNotification *EmailNotification) delete () error {
	return db.Model(EmailNotification{}).Where("id = ?", emailNotification.Id).Delete(emailNotification).Error
}
// ######### END CRUD Functions ############

// Вызов уведомления
func (emailNotification EmailNotification) Execute(data map[string]interface{}) error {

	// Проверяем статус уведомления
	if !emailNotification.Enabled {
		return utils.Error{Message: "Уведомление не может быть отправлено т.к. находится в статусе - 'Отключено'"}
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
	emailTemplateEntity, err := EmailTemplate{}.get(*emailNotification.EmailTemplateId)
	if err != nil {
		return err
	}
	if emailTemplateEntity.GetAccountId() != emailNotification.AccountId {
		return utils.Error{Message: "Ошибка отправления Уведомления - шаблон принадлежит другому аккаунту 2"}
	}
	emailTemplate, ok := emailTemplateEntity.(*EmailTemplate)
	if !ok {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удалось получить шаблон"}
	}

	// Проверяем, чтобы был почтовые ящики, с которого отправляем
	if emailNotification.EmailBox.Id < 1 {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается получить почтовый ящик"}
	}




	// Загружаем данные почтового ящика
	err = emailNotification.EmailBox.load()
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается загрузить данные WEbSite"}
	}

	// NEW NEW ============

	var users = make([]User,0)

	// 1. Собираем список пользователей
	if emailNotification.SendingToUsers {
		for i := range emailNotification.RecipientUsers {
			users = append(users, emailNotification.RecipientUsers[i])

		}
	}
	if emailNotification.ParseRecipientUser {
		if userSTR, ok := data["userId"]; ok {
			if userId, ok := userSTR.(uint); ok {
				user, err := account.GetUser(userId)
				if err == nil && user.Email != "" {
					users = append(users, *user)
				}
			}
		}
	}
	if emailNotification.ParseRecipientCustomer {
		if customerSTR, ok := data["customerId"]; ok {
			if customerId, ok := customerSTR.(uint); ok {
				customer, err := account.GetUser(customerId)
				if err == nil && customer.Email != "" {
					users = append(users, *customer)
				}
			}
		}
	}
	if emailNotification.ParseRecipientManager {
		if managerSTR, ok := data["managerId"]; ok {
			if managerId, ok := managerSTR.(uint); ok {
				manager, err := account.GetUser(managerId)
				if err == nil && manager.Email != "" {
					users = append(users, *manager)
				}
			}
		}
	}

	// 2. Отправляем списку
	for i := range users {

		history := &MTAHistory{
			HashId:  strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true)),
			AccountId: emailNotification.AccountId,
			UserId: &users[i].Id,
			Email: users[i].Email,
			OwnerId: emailNotification.Id,
			OwnerType: "email_notifications",
			EmailTemplateId: emailTemplate.Id,
			NumberOfAttempts: 1,
			Succeed: false,
		}

		unsubscribeUrl := account.GetUnsubscribeUrl(users[i], *history)
		pixelURL := account.GetPixelUrl(*history)

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

		err = emailTemplate.SendMail(emailNotification.EmailBox, users[i].Email, _subject, vData, unsubscribeUrl)
		if err != nil {
			history.Succeed = false
		} else {
			history.Succeed = true
		}

		_, _ = history.create()
	}

	// END NEW NEW =========


	// 2. Готовим список почтовых адресов и отправляем
	if emailNotification.SendingToFixedAddresses {
		// 2. Готовим список фиксированных адресов


		emailList := utils.ParseJSONBToString(emailNotification.RecipientList)

		for _,v := range emailList {

			history := &MTAHistory{
				HashId:  strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true)),
				AccountId: emailNotification.AccountId,
				Email: v,
				OwnerId: emailNotification.Id,
				OwnerType: "email_notifications",
				EmailTemplateId: emailTemplate.Id,
				NumberOfAttempts: 1,
				Succeed: false,
			}

			pixelURL := account.GetPixelUrl(*history)

			// Временные данные
		/*	_vData, err := emailTemplate.PrepareViewData(emailNotification.Subject, emailNotification.PreviewText, data, pixelURL, nil)
			if err != nil {
				return utils.Error{Message: "Ошибка отправления Уведомления - не удается подготовить данные для сообщения"}
			}*/

			// Компилируем тему письма
			_subject, err := parseSubjectByData(emailNotification.Subject, data)
			if err != nil {
				return utils.Error{Message: "Ошибка отправления Уведомления - не удается прочитать тему сообщения"}
			}
			if _subject == "" {
				_subject = fmt.Sprintf("Уведомление по почте #%v", emailNotification.Id)
			}

			// Рабочие данные
			vData, err := emailTemplate.PrepareViewData(_subject, emailNotification.PreviewText, data, pixelURL, nil)
			if err != nil {
				return utils.Error{Message: "Ошибка отправления Уведомления - не удается подготовить данные для сообщения"}
			}

			err = emailTemplate.SendMail(emailNotification.EmailBox, v, _subject, vData, "")
			if err != nil {
				history.Succeed = false
			} else {
				history.Succeed = true
			}
			_, _ = history.create()
		}
	}

	return nil
}

// func parseSubjectByData(tpl string, vData *ViewData) (string, error) {
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
