package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type OrderSource struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"accountID" gorm:"index,not null"` // аккаунт-владелец ключа

	Name		string 	`json:"name" gorm:"type:varchar(255);"` // "Заказ из корзины", "Заказ по телефону", "Пропущенный звонок", "Письмо.."
	Description string 	`json:"description" gorm:"type:varchar(255);"` // Описание назначения канала
	
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ############# Entity interface #############
func (orderSource OrderSource) GetID() uint { return orderSource.ID }
func (orderSource *OrderSource) setID(id uint) { orderSource.ID = id }
func (orderSource OrderSource) GetAccountID() uint { return orderSource.AccountID }
func (orderSource *OrderSource) setAccountID(id uint) { orderSource.AccountID = id }
func (orderSource OrderSource) SystemEntity() bool { return orderSource.AccountID == 1 }

// ############# Entity interface #############

func (OrderSource) PgSqlCreate() {
	if !db.HasTable(&OrderSource{}) {
		db.CreateTable(&OrderSource{})
	}
	db.Model(&OrderSource{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}
func (orderSource *OrderSource) BeforeCreate(scope *gorm.Scope) error {
	orderSource.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (orderSource OrderSource) create() (Entity, error)  {
	// if err := db.Create(&orderSource).Find(&orderSource, orderSource.ID).Error; err != nil {
	wb := orderSource
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (OrderSource) get(id uint) (Entity, error) {

	var orderSource OrderSource

	err := db.First(&orderSource, id).Error
	if err != nil {
		return nil, err
	}
	return &orderSource, nil
}
func (orderSource *OrderSource) load() error {
	if orderSource.ID < 1 {
		return utils.Error{Message: "Невозможно загрузить OrderSource - не указан  ID"}
	}

	err := db.First(orderSource).Error
	if err != nil {
		return err
	}
	return nil
}

func (OrderSource) getList(accountID uint, sortBy string) ([]Entity, uint, error) {

	orderSources := make([]OrderSource,0)
	var total uint

	err := db.Model(&OrderSource{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountID).
		Find(&orderSources).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&OrderSource{}).Where("account_id = ?", accountID).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(orderSources))
	for i,_ := range orderSources {
		entities[i] = &orderSources[i]
	}

	return entities, total, nil
}

func (OrderSource) getPaginationList(accountID uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	orderSources := make([]OrderSource,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&OrderSource{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountID).
			Find(&orderSources, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&OrderSource{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountID, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&OrderSource{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountID).
			Find(&orderSources).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&OrderSource{}).Where("account_id = ?", accountID).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(orderSources))
	for i,_ := range orderSources {
		entities[i] = &orderSources[i]
	}

	return entities, total, nil
}

func (OrderSource) getByEvent(eventName string) (*OrderSource, error) {

	wh := OrderSource{}

	if err := db.First(&wh, "event_type = ?", eventName).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}

func (orderSource *OrderSource) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(orderSource).Where("id", orderSource.ID).Omit("id", "account_id").Updates(input).Error
}

func (orderSource OrderSource) delete () error {
	return db.Model(OrderSource{}).Where("id = ?", orderSource.ID).Delete(orderSource).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

