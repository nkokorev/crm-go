package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"html/template"
	"reflect"
	"time"
)

type EmailNotification struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	Enabled 		bool 	`json:"enabled" gorm:"type:bool;default:false;"` // отключить / включить
	Delay			uint 	`json:"delay" gorm:"type:int;default:0"` // Задержка перед отправлением в минутах: [0-180]
	
	Name 			string 	`json:"name" gorm:"type:varchar(128);default:''"` // "Оповещение менеджера", "Оповещение клиента"

	Subject			string 	`json:"subject" gorm:"type:varchar(128);default:''"` // Тема сообщения, компилируются
	
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
	ParseRecipientUsers	bool	`json:"parseRecipientUsers" gorm:"type:bool;default:false"` // Спарсить из контекста пользователя(ей) по userId / users: ['email@mail.ru']

	// ==========================================

	// Скрытый список пользователей для Data и фронтенда
	RecipientUsers []User	`json:"_recipientUsers" gorm:"-"`

	CreatedAt 		time.Time `json:"createdAt"`
	UpdatedAt 		time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (emailNotification EmailNotification) GetId() uint { return emailNotification.Id }
func (emailNotification *EmailNotification) setId(id uint) { emailNotification.Id = id }
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


func (EmailNotification) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return EmailNotification{}.getPaginationList(accountId, 0, 100, sortBy, "")
}

func (EmailNotification) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

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

func (emailNotification EmailNotification) delete () error {
	return db.Model(EmailNotification{}).Where("id = ?", emailNotification.Id).Delete(emailNotification).Error
}
// ######### END CRUD Functions ############

// Вызов уведомления
func (emailNotification EmailNotification) Execute(data map[string]interface{}) error {

	if !emailNotification.Enabled {
		return utils.Error{Message: "Уведомление не может быть отправлено т.к. находится в статусе - 'Отключено'"}
	}

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

	vData, err := emailTemplate.PrepareViewData(data)

	if emailNotification.EmailBox.Id < 1 {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается получить почтовый ящик"}
	}

	_subject, err := parseSubjectByData(emailNotification.Subject, vData)
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается прочитать тему сообщения"}
	}
	if _subject == "" {
		_subject = fmt.Sprintf("Уведомление по почте #%v", emailNotification.Id)
	}

	// 1. Список пользователей
	var userEmails = make([]string,0)
	for i,_ := range emailNotification.RecipientUsers {
		userEmails = append(userEmails,emailNotification.RecipientUsers[i].Email)
	}

	// 2. список фиксированных адресов
	emailList := utils.ParseJSONBToString(emailNotification.RecipientList)

	// fmt.Println("emailNotification.EmailBox: ", emailNotification.EmailBox.Id)
	// return nil


	err = emailNotification.EmailBox.load()
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается загрузить данные WEbSite"}
	}


	// Проверяем и отправляем 1
	if emailNotification.SendingToUsers {
		for _,v := range(userEmails) {
			err = emailTemplate.SendMail(emailNotification.EmailBox, v, _subject, vData)
			if err != nil {
				fmt.Println("Ошибка отправления: ", err)
			}
		}
	}

	// Проверяем и отправляем 2
	if emailNotification.SendingToFixedAddresses {
		for _,v := range(emailList) {
			err = emailTemplate.SendMail(emailNotification.EmailBox, v, _subject, vData)
			if err != nil {
				fmt.Println("Ошибка отправления: ", err)
			}
		}
	}


	return nil
}

func parseSubjectByData(tpl string, vData *ViewData) (string, error) {

	body := new(bytes.Buffer)

	tmpl, err := template.New("et.Name").Parse(tpl)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(body, vData)
	if err != nil {
		return "", utils.Error{Message: fmt.Sprintf("Ошибка в заголовке шаблона")}
	}

	return body.String(), nil
}
