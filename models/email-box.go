package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"net/mail"
)

type EmailBox struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId uint `json:"-" gorm:"type:int;index;not null;"`
	WebSiteId uint `json:"webSiteId" gorm:"type:int;not null;"` // какой сайт обязательно!
	
	Default bool `json:"default" gorm:"type:bool;default:false"` // является ли дефолтным почтовым ящиком для домена
	Allowed bool `json:"allowed" gorm:"type:bool;default:true"` // прошел ли проверку домен на право отправлять с него почту
	
	Name string `json:"name" gorm:"type:varchar(32);not null;"` // от имени кого отправляется RatusCRM, Магазин 357 грамм..
	Box string `json:"box" gorm:"type:varchar(32);not null;"` // обратный адрес info@, news@, mail@...

	WebSite WebSite `json:"-"`
}

func (EmailBox) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&EmailBox{})
	db.Model(&EmailBox{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&EmailBox{}).AddForeignKey("web_site_id", "web_sites(id)", "CASCADE", "CASCADE")
}
func (emailBox *EmailBox) BeforeCreate(scope *gorm.Scope) error {
	emailBox.Id = 0
	return nil
}

// ############# Entity interface #############
func (emailBox EmailBox) GetId() uint { return emailBox.Id }
func (emailBox *EmailBox) setId(id uint) { emailBox.Id = id }
func (emailBox *EmailBox) setPublicId(id uint) { }
func (emailBox EmailBox) GetAccountId() uint { return emailBox.AccountId }
func (emailBox *EmailBox) setAccountId(id uint) { emailBox.AccountId = id }
func (EmailBox) SystemEntity() bool { return false }

// ############# Entity interface #############
func (emailBox EmailBox) create() (Entity, error)  {
	if emailBox.Box == "" {
		return nil, utils.Error{Message: "Необходимо указать имя почтового ящика"}
	}

	eb := emailBox
	if err := db.Create(&eb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &eb

	return entity, nil
}

func (EmailBox) get(id uint) (Entity, error) {

	var emailBox EmailBox

	err := db.Preload("WebSite").First(&emailBox, id).Error
	if err != nil {
		return nil, err
	}
	return &emailBox, nil
}
func (emailBox *EmailBox) load() error {

	if emailBox.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailBox - не указан  Id"}
	}
	err := db.Preload("WebSite").First(emailBox, emailBox.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailBox *EmailBox) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (EmailBox) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	webHooks := make([]EmailBox,0)
	var total uint

	err := db.Model(&EmailBox{}).Limit(100).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&webHooks).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&EmailBox{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i,_ := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}

func (EmailBox) getListByWebSite(accountId uint, webSiteId uint, sortBy string) ([]EmailBox, error) {

	emailBoxes := make([]EmailBox,0)

	err := db.Model(&EmailBox{}).Limit(100).Order(sortBy).Where( "account_id = ? AND web_site_id = ?", accountId, webSiteId).
		Find(&emailBoxes).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}

	return emailBoxes, nil
}

func (EmailBox) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	webHooks := make([]EmailBox,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&EmailBox{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR box ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailBox{}).
			Where("account_id = ? AND name ILIKE ? OR box ILIKE ?", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&EmailBox{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&EmailBox{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i,_ := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}

func (emailBox *EmailBox) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(emailBox).Omit("id", "account_id").Updates(input).Error
}

func (emailBox *EmailBox) delete () error {
	return db.Model(EmailBox{}).Where("id = ?", emailBox.Id).Delete(emailBox).Error
}


// ########### EmailBox FUNCTIONAL ###########
func (ebox EmailBox) GetMailAddress() mail.Address {
	return mail.Address{Name: ebox.Name, Address: ebox.Box + "@" + ebox.WebSite.Hostname}
}

// ########### END OF EmailBox FUNCTIONAL ###########

