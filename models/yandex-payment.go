package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

type YandexPayment struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя интеграции
	
	ApiKey	string	`json:"apiKey" gorm:"type:varchar(128);"` // ApiKey от яндекс кассы
	ShopId	uint	`json:"shopId" gorm:"type:int;"` // shop id от яндекс кассы

	Code		WebHookType `json:"code" gorm:"type:varchar(128);default:''"` // Имя события

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // обрабатывать ли вебхук
	
	Description 		string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание что к чему)
	URL 		string 	`json:"url" gorm:"type:varchar(255);"` // вызов, который совершается
	HttpMethod		string `json:"httpMethod" gorm:"type:varchar(15);default:'get';"` // Тип вызова (GET, POST, PUT, puth и т.д.)
	//URLTemplate 		template.Template 	`json:"url" gorm:"type:varchar(255);"` // вызов, который совершается
}

// ############# Entity interface #############
func (yandexPayment YandexPayment) getId() uint { return yandexPayment.ID }
func (yandexPayment *YandexPayment) setId(id uint) { yandexPayment.ID = id }
func (yandexPayment YandexPayment) GetAccountId() uint { return yandexPayment.AccountID }
func (yandexPayment *YandexPayment) setAccountId(id uint) { yandexPayment.AccountID = id }
func (YandexPayment) systemEntity() bool { return false }

// ############# Entity interface #############

func (YandexPayment) PgSqlCreate() {
	db.CreateTable(&YandexPayment{})
	db.Model(&YandexPayment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (yandexPayment *YandexPayment) BeforeCreate(scope *gorm.Scope) error {
	yandexPayment.ID = 0
	return nil
}
func (YandexPayment) TableName() string {
	return "web_hooks"
}

// ######### CRUD Functions ############
func (yandexPayment YandexPayment) create() (Entity, error)  {
	// if err := db.Create(&yandexPayment).Find(&yandexPayment, yandexPayment.ID).Error; err != nil {
	wb := yandexPayment
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (YandexPayment) get(id uint) (Entity, error) {

	var yandexPayment YandexPayment

	err := db.First(&yandexPayment, id).Error
	if err != nil {
		return nil, err
	}
	return &yandexPayment, nil
}
func (yandexPayment *YandexPayment) load() error {
	if yandexPayment.ID < 1 {
		return utils.Error{Message: "Невозможно загрузить YandexPayment - не указан  ID"}
	}

	err := db.First(yandexPayment).Error
	if err != nil {
		return err
	}
	return nil
}

func (YandexPayment) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	webHooks := make([]YandexPayment,0)
	var total uint

	err := db.Model(&YandexPayment{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&webHooks).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&YandexPayment{}).Where("account_id = ?", accountId).Count(&total).Error
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

func (YandexPayment) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	webHooks := make([]YandexPayment,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&YandexPayment{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&YandexPayment{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&YandexPayment{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&YandexPayment{}).Where("account_id = ?", accountId).Count(&total).Error
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

func (YandexPayment) getByEvent(eventName string) (*YandexPayment, error) {

	wh := YandexPayment{}

	if err := db.First(&wh, "event_type = ?", eventName).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}

func (yandexPayment *YandexPayment) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(yandexPayment).Where("id", yandexPayment.ID).Omit("id", "account_id").Updates(input).Error
}

func (yandexPayment YandexPayment) delete () error {
	return db.Model(YandexPayment{}).Where("id = ?", yandexPayment.ID).Delete(yandexPayment).Error
}
// ######### END CRUD Functions ############
