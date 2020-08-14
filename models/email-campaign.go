package models

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"reflect"
	"strings"
	"time"
)

type EmailCampaign struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	Enabled 		bool 	`json:"enabled" gorm:"type:bool;default:false;"` // отключить / включить

	// Delay			uint 	`json:"delay" gorm:"type:int;default:0"` // Задержка перед отправлением в минутах: [0-180]
	DelayTime		time.Duration `json:"delayTime" gorm:"type:int8;default:0"`// << учитывается только время [0-24]
	
	Name 			string 	`json:"name" gorm:"type:varchar(128);default:''"` // "Оповещение менеджера", "Оповещение клиента"

	Subject			string 	`json:"subject" gorm:"type:varchar(128);not null;"` // Тема сообщения, компилируются
	PreviewText		string 	`json:"previewText" gorm:"type:varchar(255);default:''"` // Тема сообщения, компилируются
	
	EmailTemplateId *uint 	`json:"emailTemplateId" gorm:"type:int;default:null;"` // всегда должен быть шаблон, иначе смысла в нем нет
	EmailTemplate 	EmailTemplate 	`json:"emailTemplate" gorm:"preload:true"`

	EmailBoxId		*uint 	`json:"emailBoxId" gorm:"type:int;default:null;"` // С какого ящика идет отправка
	EmailBox		EmailBox `json:"emailBox" gorm:"preload:false"`
	// =============   Настройки получателей    ===================

	// Список пользователей позволяет сделать "рассылку" уведомления по email-адреса пользователей, до 10 человек.
	// SendingToUsers		bool			`json:"sendingToUsers" gorm:"type:bool;default:false"` // Отправлять пользователем RatusCRM (на их почтовые адреса, при их наличии)
	RecipientUsersList	postgres.Jsonb 	`json:"recipientUsersList" gorm:"type:JSONB;DEFAULT '{}'::JSONB"` // список id пользователей, которые получат уведомление

	// Список фиксированных адресов позволяет сделать "рассылку" по своей базе, до 10 человек.
	// SendingToFixedAddresses	bool	`json:"sendingToFixedAddresses" gorm:"type:bool;default:false"` // Отправлять ли на фиксированный адреса
	// RecipientList			postgres.Jsonb	`json:"recipientList" gorm:"type:JSONB;DEFAULT '{}'::JSONB"` // фиксированный список адресов, на которые будет произведено уведомление
	
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
func (emailCampaign EmailCampaign) GetId() uint { return emailCampaign.Id }
func (emailCampaign *EmailCampaign) setId(id uint) { emailCampaign.Id = id }
func (emailCampaign *EmailCampaign) setPublicId(publicId uint) { emailCampaign.PublicId = publicId }
func (emailCampaign EmailCampaign) GetAccountId() uint { return emailCampaign.AccountId }
func (emailCampaign *EmailCampaign) setAccountId(id uint) { emailCampaign.AccountId = id }
func (EmailCampaign) SystemEntity() bool { return false }
func (EmailCampaign) GetType() string { return "email_campaigns" }
func (emailCampaign EmailCampaign) IsEnabled() bool { return emailCampaign.Enabled }

// ############# Entity interface #############

func (EmailCampaign) PgSqlCreate() {
	db.CreateTable(&EmailCampaign{})
	db.Model(&EmailCampaign{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&EmailCampaign{}).AddForeignKey("email_template_id", "email_templates(id)", "RESTRICT", "CASCADE")
}
func (emailCampaign *EmailCampaign) BeforeCreate(scope *gorm.Scope) error {
	emailCampaign.Id = 0

	// PublicId
	lastIdx := uint(0)
	var eq EmailCampaign

	err := db.Where("account_id = ?", emailCampaign.AccountId).Select("public_id").Last(&eq).Error
	if err != nil && err != gorm.ErrRecordNotFound { return err}
	if err == gorm.ErrRecordNotFound {
		lastIdx = 0
	} else {
		lastIdx = eq.PublicId
	}
	emailCampaign.PublicId = lastIdx + 1

	return nil
}
func (emailCampaign *EmailCampaign) AfterFind() (err error) {

	// Собираем пользователей
	b, err := emailCampaign.RecipientUsersList.MarshalJSON()
	if err != nil {
		return err
	}

	var arr []uint
	err = json.Unmarshal(b, &arr)
	if err != nil { return err }

	err = db.Find(&emailCampaign.RecipientUsers, "id IN (?)", arr).Error
	if err != nil  {return err}


	/////////////////////////////////////

	if reflect.DeepEqual(emailCampaign.RecipientUsersList, *new(postgres.Jsonb)) {
		emailCampaign.RecipientUsersList = postgres.Jsonb{RawMessage: []byte("[]")}
	}


	return nil
}

// ######### CRUD Functions ############
func (emailCampaign EmailCampaign) create() (Entity, error)  {

	en := emailCampaign

	if err := db.Create(&en).Error; err != nil {
		return nil, err
	}

	err := en.GetPreloadDb(false,false, true).First(&en, en.Id).Error
	if err != nil {
		return nil, err
	}

	var newItem Entity = &en

	return newItem, nil
}

func (EmailCampaign) get(id uint) (Entity, error) {

	var emailCampaign EmailCampaign

	err := db.Preload("EmailBox").First(&emailCampaign, id).Error
	if err != nil {
		return nil, err
	}
	return &emailCampaign, nil
}
func (emailCampaign *EmailCampaign) load() error {

	err := db.Preload("EmailBox").First(emailCampaign, emailCampaign.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailCampaign *EmailCampaign) loadByPublicId() error {
	
	if emailCampaign.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailCampaign - не указан  Id"}
	}
	if err := emailCampaign.GetPreloadDb(false,false, true).First(emailCampaign, "account_id = ? AND public_id = ?", emailCampaign.AccountId, emailCampaign.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (EmailCampaign) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return EmailCampaign{}.getPaginationList(accountId, 0, 100, sortBy, "",nil)
}
func (EmailCampaign) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, uint, error) {

	emailNotifications := make([]EmailCampaign,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		// jsearch := search
		search = "%"+search+"%"

		err := db.Model(&EmailCampaign{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&emailNotifications, "name ILIKE ? OR description ILIKE ?", search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailCampaign{}).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&EmailCampaign{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&emailNotifications).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailCampaign{}).Where("account_id = ?", accountId).Count(&total).Error
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

func (emailCampaign *EmailCampaign) update(input map[string]interface{}) error {

	// Приводим в опрядок
	input = utils.FixJSONB_String(input, []string{"recipientList"})
	input = utils.FixJSONB_Uint(input, []string{"recipientUsersList"})

	// delete(input, "recipientList")
	// delete(input, "recipientUsersList")
	delete(input, "emailTemplate")
	delete(input, "emailBox")


	// if err := db.Model(EmailCampaign{}).Where("id = ?", emailCampaign.Id).Omit("id", "account_id").Updates(input).Error; err != nil {
	// 	return err
	// }
	// fmt.Println("emailCampaign id", emailCampaign.EmailBoxId)

/*	if err := db.Model(emailCampaign).
		Omit("id", "account_id","created_at").Update(input).Error; err != nil {
		return err
	}*/

	// work!!!
	if err := db.Set("gorm:association_autoupdate", false).Model(EmailCampaign{}).Where(" id = ?", emailCampaign.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := db.Preload("EmailBox").Preload("EmailTemplate").First(emailCampaign, emailCampaign.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (emailCampaign *EmailCampaign) delete () error {
	return db.Model(EmailCampaign{}).Where("id = ?", emailCampaign.Id).Delete(emailCampaign).Error
}
// ######### END CRUD Functions ############

func (emailCampaign *EmailCampaign) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(emailCampaign)
	} else {
		_db = _db.Model(&EmailCampaign{})
	}

	if preload {
		return _db.Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
			return db.Select(EmailTemplate{}.SelectArrayWithoutData())
		}).Preload("EmailBox")
	} else {
		return _db
	}
}

// Вызов уведомления
func (emailCampaign EmailCampaign) Execute(data map[string]interface{}) error {

	// Проверяем статус уведомления
	if !emailCampaign.Enabled {
		return utils.Error{Message: "Уведомление не может быть отправлено т.к. находится в статусе - 'Отключено'"}
	}

	// Проверяем тело сообщения (не должно быть пустое)
	if emailCampaign.Subject == "" {
		return utils.Error{Message: "Уведомление не может быть отправлено т.к. нет темы сообщения"}
	}

	// Get Account
	account, err := GetAccount(emailCampaign.AccountId)
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается найти аккаунт"}
	}

	// Находим шаблон письма
	emailTemplateEntity, err := EmailTemplate{}.get(*emailCampaign.EmailTemplateId)
	if err != nil {
		return err
	}
	if emailTemplateEntity.GetAccountId() != emailCampaign.AccountId {
		return utils.Error{Message: "Ошибка отправления Уведомления - шаблон принадлежит другому аккаунту 2"}
	}
	emailTemplate, ok := emailTemplateEntity.(*EmailTemplate)
	if !ok {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удалось получить шаблон"}
	}

	// Проверяем, чтобы был почтовые ящики, с которого отправляем
	if emailCampaign.EmailBox.Id < 1 {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается получить почтовый ящик"}
	}

	// Загружаем данные почтового ящика
	err = emailCampaign.EmailBox.load()
	if err != nil {
		return utils.Error{Message: "Ошибка отправления Уведомления - не удается загрузить данные WEbSite"}
	}

	// NEW NEW ============

	var users = make([]User,0)

	// 1. Собираем список пользователей
	for i := range emailCampaign.RecipientUsers {
		users = append(users, emailCampaign.RecipientUsers[i])

	}
	if emailCampaign.ParseRecipientUser {
		if userSTR, ok := data["userId"]; ok {
			if userId, ok := userSTR.(uint); ok {
				user, err := account.GetUser(userId)
				if err == nil && user.Email != "" {
					users = append(users, *user)
				}
			}
		}
	}
	if emailCampaign.ParseRecipientCustomer {
		if customerSTR, ok := data["customerId"]; ok {
			if customerId, ok := customerSTR.(uint); ok {
				customer, err := account.GetUser(customerId)
				if err == nil && customer.Email != "" {
					users = append(users, *customer)
				}
			}
		}
	}
	if emailCampaign.ParseRecipientManager {
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
			AccountId: emailCampaign.AccountId,
			UserId: &users[i].Id,
			Email: users[i].Email,
			OwnerId: emailCampaign.Id,
			OwnerType: "email_notifications",
			EmailTemplateId: utils.UINTp(emailTemplate.Id),
			NumberOfAttempts: 1,
			Succeed: false,
		}

		unsubscribeUrl := account.GetUnsubscribeUrl(users[i], *history)
		pixelURL := account.GetPixelUrl(*history)

		// Компилируем тему письма
		_subject, err := parseSubjectByData(emailCampaign.Subject, data)
		if err != nil {
			log.Printf("Ошибка отправления Уведомления - не удается прочитать тему сообщения. emailNotificationId: %v\n", emailCampaign.Id)
			continue
		}
		if _subject == "" {
			_subject = fmt.Sprintf("Уведомление по почте #%v", emailCampaign.Id)
		}

		vData, err := emailTemplate.PrepareViewData(_subject, emailCampaign.PreviewText, data, pixelURL, &unsubscribeUrl)
		if err != nil {
			log.Printf("Ошибка отправления Уведомления - не удается подготовить данные для сообщения. emailNotificationId: %v\n", emailCampaign.Id)
			continue
		}

		// if now ==
		if emailCampaign.DelayTime == 0 {
			err = emailTemplate.SendMail(emailCampaign.EmailBox, users[i].Email, _subject, vData, unsubscribeUrl)
			if err != nil {
				history.Succeed = false
			} else {
				history.Succeed = true
			}
		} else {
			// ставим в очередь
		}


		_, _ = history.create()
	}

	return nil
}
