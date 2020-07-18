package models

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"reflect"
	"time"
)

type EmailNotification struct {
	ID     			uint   	`json:"id" gorm:"primary_key"`
	AccountID 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	Enabled 		bool 	`json:"enabled" gorm:"type:bool;default:true"` // отключить / включить
	Delay			uint 	`json:"delay" gorm:"type:int;default:0"` // Задержка перед отправлением в минутах: [0-180]
	
	Name 			string 	`json:"name" gorm:"type:varchar(128);default:''"` // "Оповещение менеджера", "Оповещение клиента"
	Description		string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание что к чему)

	EmailTemplateId uint 	`json:"emailTemplateId" gorm:"type:int;"` // всегда должен быть шаблон, иначе смысла в нем нет
	EmailTemplate 	EmailTemplate 	`json:"-" gorm:"preload:true"`


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

	RecipientUsers []User	`json:"_recipientUsers" gorm:"-"`

	CreatedAt 		time.Time `json:"createdAt"`
	UpdatedAt 		time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (emailNotification EmailNotification) getId() uint { return emailNotification.ID }
func (emailNotification *EmailNotification) setId(id uint) { emailNotification.ID = id }
func (emailNotification EmailNotification) GetAccountId() uint { return emailNotification.AccountID }
func (emailNotification *EmailNotification) setAccountId(id uint) { emailNotification.AccountID = id }
func (EmailNotification) systemEntity() bool { return false }

// ############# Entity interface #############

func (EmailNotification) PgSqlCreate() {
	db.CreateTable(&EmailNotification{})
	db.Model(&EmailNotification{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&EmailNotification{}).AddForeignKey("email_template_id", "email_templates(id)", "CASCADE", "CASCADE")
}
func (emailNotification *EmailNotification) BeforeCreate(scope *gorm.Scope) error {
	emailNotification.ID = 0
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
	var newItem Entity = &emailNotification

	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (EmailNotification) get(id uint) (Entity, error) {

	var emailNotification EmailNotification

	err := db.First(&emailNotification, id).Error
	if err != nil {
		return nil, err
	}
	return &emailNotification, nil
}
func (emailNotification *EmailNotification) load() error {

	err := db.First(emailNotification).Error
	if err != nil {
		return err
	}
	return nil
}

func (EmailNotification) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	emailNotifications := make([]EmailNotification,0)
	var total uint

	err := db.Model(&EmailNotification{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&emailNotifications).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&EmailNotification{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailNotifications))
	for i,_ := range emailNotifications {
		entities[i] = &emailNotifications[i]
	}

	return entities, total, nil
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

func (EmailNotification) getListByAccount(accountId uint) ([]EmailNotification, error) {

	emailNotifications := make([]EmailNotification,0)

	err := db.Find(&emailNotifications, "accountId = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return emailNotifications, nil

}

func (emailNotification *EmailNotification) update(input map[string]interface{}) error {

	input = utils.FixJSONB(input, []string{"recipientList"})
	delete(input, "recipientList")
	delete(input, "recipientUsersList")

	if err := db.Model(emailNotification).Update(input).Error; err != nil {
		return err
	}

	return db.Set("gorm:association_autoupdate", false).Model(emailNotification).Omit("id", "account_id").Update(input).Error
}

func (emailNotification EmailNotification) delete () error {
	return db.Model(EmailNotification{}).Where("id = ?", emailNotification.ID).Delete(emailNotification).Error
}
// ######### END CRUD Functions ############

// Вызов уведомления
func (emailNotification EmailNotification) Execute(data map[string]interface{}) error {

	if data == nil {
		fmt.Println("Execute EmailNotification of data[] is null!")
	} else {
		fmt.Println("Execute EmailNotification of data[] not null!!")
	}

	return nil
}
