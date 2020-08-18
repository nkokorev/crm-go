package models

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type EmailCampaign struct {
	Id     			uint   	`json:"id" gorm:"primary_key"`
	PublicId		uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"`
	AccountId 		uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Запущена ли кампания
	Enabled 		bool 	`json:"enabled" gorm:"type:bool;default:false;"`

	// Планируемое время старта
	ScheduleRun		time.Time `json:"scheduleRun" gorm:"type:int8;"`

	// Ежемесячный дайджест !
	Name 			string 	`json:"name" gorm:"type:varchar(128);default:''"`

	// Тема сообщения и preview-текст, компилируются
	Subject			string 	`json:"subject" gorm:"type:varchar(128);not null;"`
	PreviewText		string 	`json:"previewText" gorm:"type:varchar(255);default:''"`

	// Шаблон email-сообщения
	EmailTemplateId uint 	`json:"emailTemplateId" gorm:"type:int;not null;"`

	// Отправитель, может устанавливаться в конце
	EmailBoxId		uint 	`json:"emailBoxId" gorm:"type:int;not null;"`

	// RecipientList	postgres.Jsonb 	`json:"recipientList" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	UserSegments	[]UserSegment `json:"userSegments"`

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
// ############# End Entity interface #############

func (EmailCampaign) PgSqlCreate() {
	db.CreateTable(&EmailCampaign{})
	db.Model(&EmailCampaign{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&EmailCampaign{}).AddForeignKey("email_template_id", "email_templates(id)", "RESTRICT", "CASCADE")
}
func (emailCampaign *EmailCampaign) BeforeCreate(scope *gorm.Scope) error {
	emailCampaign.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&EmailCampaign{}).Where("account_id = ?",  emailCampaign.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	emailCampaign.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (emailCampaign *EmailCampaign) AfterFind() (err error) {
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

	err := emailCampaign.GetPreloadDb(false,false,true).First(&emailCampaign, id).Error
	if err != nil {
		return nil, err
	}
	return &emailCampaign, nil
}
func (emailCampaign *EmailCampaign) load() error {

	err := emailCampaign.GetPreloadDb(false,false,true).First(emailCampaign, emailCampaign.Id).Error
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

	emailCampaigns := make([]EmailCampaign,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&EmailCampaign{}).GetPreloadDb(true,false,true).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&emailCampaigns, "name ILIKE ? OR subject ILIKE ? OR preview_text ILIKE ?", search,search,search).Error

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

		err := (&EmailCampaign{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailTemplate", func(db *gorm.DB) *gorm.DB {
				return db.Select(EmailTemplate{}.SelectArrayWithoutData())
			}).Preload("EmailBox").
			Find(&emailCampaigns).Error
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
	entities := make([]Entity,len(emailCampaigns))
	for i := range emailCampaigns {
		entities[i] = &emailCampaigns[i]
	}

	return entities, total, nil
}

func (emailCampaign *EmailCampaign) update(input map[string]interface{}) error {

	if err := emailCampaign.GetPreloadDb(true,false,false).Where(" id = ?", emailCampaign.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := emailCampaign.GetPreloadDb(true,false,false).First(emailCampaign, emailCampaign.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (emailCampaign *EmailCampaign) delete () error {
	return emailCampaign.GetPreloadDb(true,true,false).Where("id = ?", emailCampaign.Id).Delete(emailCampaign).Error
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

// Отправка кампании
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
	emailTemplateEntity, err := EmailTemplate{}.get(emailCampaign.EmailTemplateId)
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

	fmt.Println(account, emailTemplate)

	return nil
}
