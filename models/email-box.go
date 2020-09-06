package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net/mail"
)

type EmailBox struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	WebSiteId 	uint 	`json:"web_site_id" gorm:"type:int;not null;"` // какой сайт обязательно!
	
	Default 	bool 	`json:"default" gorm:"type:bool;default:false"` // является ли дефолтным почтовым ящиком для домена
	Allowed 	bool 	`json:"allowed" gorm:"type:bool;default:true"` // прошел ли проверку домен на право отправлять с него почту
	
	Name 		string 	`json:"name" gorm:"type:varchar(32);not null;"` // от имени кого отправляется RatusCRM, Магазин 357 грамм..
	Box 		string 	`json:"box" gorm:"type:varchar(32);not null;"` // обратный адрес info@, news@, mail@...

	WebSite 	WebSite `json:"-"`
}

func (EmailBox) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	if err := db.Migrator().CreateTable(&EmailBox{});err != nil {log.Fatal(err)}
	// db.Model(&EmailBox{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&EmailBox{}).AddForeignKey("web_site_id", "web_sites(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE email_boxes " +
		"ADD CONSTRAINT email_boxes_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT email_boxes_web_site_id_fkey FOREIGN KEY (web_site_id) REFERENCES web_sites(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (emailBox *EmailBox) BeforeCreate(tx *gorm.DB) error {
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

	_item := emailBox
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,true, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (EmailBox) get(id uint, preloads []string) (Entity, error) {

	var item EmailBox

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (emailBox *EmailBox) load(preloads []string) error {

	if emailBox == nil {
		return utils.Error{Message: "Невозможно загрузить EmailBox - не указан  объект"}
	}

	if emailBox.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить EmailBox - не указан  Id"}
	}

	err := emailBox.GetPreloadDb(false, false, preloads).First(emailBox, emailBox.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (emailBox *EmailBox) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (EmailBox) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return EmailBox{}.getPaginationList(accountId,0,10,sortBy, "", nil,preload)
}

func (EmailBox) getListByWebSite(accountId uint, webSiteId uint, sortBy string) ([]EmailBox, error) {

	emailBoxes := make([]EmailBox,0)

	err := (&EmailBox{}).GetPreloadDb(false,true,nil).Limit(100).Order(sortBy).Where( "account_id = ? AND web_site_id = ?", accountId, webSiteId).
		Find(&emailBoxes).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}

	return emailBoxes, nil
}

func (EmailBox) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	webHooks := make([]EmailBox,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&CartItem{}).GetPreloadDb(false, false, preloads).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR box ILIKE ?", search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&CartItem{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND name ILIKE ? OR box ILIKE ?", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&CartItem{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&CartItem{}).GetPreloadDb(false, false, nil).Where("account_id = ?", accountId).Count(&total).Error
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

func (emailBox *EmailBox) update(input map[string]interface{}, preloads []string) error {
	delete(input,"web_site")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"web_site_id"}); err != nil {
		return err
	}
	
	if err := emailBox.GetPreloadDb(false, false, nil).Where("id = ?", emailBox.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := emailBox.GetPreloadDb(false,false, preloads).First(emailBox, emailBox.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (emailBox *EmailBox) delete () error {
	return emailBox.GetPreloadDb(true,false,nil).Where("id = ?", emailBox.Id).Delete(emailBox).Error
}


// ########### EmailBox FUNCTIONAL ###########
func (emailBox EmailBox) GetMailAddress() mail.Address {
	return mail.Address{Name: emailBox.Name, Address: emailBox.Box + "@" + emailBox.WebSite.Hostname}
}

// ########### END OF EmailBox FUNCTIONAL ###########

// ########## Work function ############
func (emailBox *EmailBox) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&emailBox)
	} else {
		_db = _db.Model(&EmailBox{})
	}

	if autoPreload {
		return db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"WebSite"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}