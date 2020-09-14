package models

import (
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Данные поставки
type InventoryItem struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	InventoryId uint	`json:"inventory_id" gorm:"type:int;index;"`
	ProductId 	uint	`json:"product_id" gorm:"type:int;index;"`

	// Сколько ед.товара ожидалось / есть по факту
	VolumeExpected	*float64 `json:"volume_expected" gorm:"type:numeric;"`
	VolumeFact		*float64 `json:"volume_fact" gorm:"type:numeric;"`

	// Закрытие просчета в размере VolumeFact. При этом делается расчет и заносится VolumeExpected в строку.
	Posted		bool	`json:"posted" gorm:"type:bool;default:false"`

	// Краткое замечание к проверке
	Description *string	`json:"description" gorm:"type:varchar(255);"`

	Product 	Product 	`json:"product"`
	Inventory 	Inventory 	`json:"inventory"`

	// когда какие данные заносились
	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
}

func (InventoryItem) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&InventoryItem{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE inventory_items " +
		"ADD CONSTRAINT inventory_items_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT inventory_items_inventory_id_fkey FOREIGN KEY (inventory_id) REFERENCES inventories(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT inventory_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}



}
func (inventoryItem *InventoryItem) BeforeCreate(tx *gorm.DB) error {
	inventoryItem.Id = 0

	inventoryItem.CreatedAt = time.Now().UTC()

	return nil
}
func (inventoryItem *InventoryItem) AfterFind(tx *gorm.DB) (err error) {
	return nil
}
func (inventoryItem *InventoryItem) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(inventoryItem)
	} else {
		_db = _db.Model(&InventoryItem{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Product","Inventory"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}

// ############# Entity interface #############
func (inventoryItem InventoryItem) GetId() uint { return inventoryItem.Id }
func (inventoryItem *InventoryItem) setId(id uint) { inventoryItem.Id = id }
func (inventoryItem *InventoryItem) setPublicId(publicId uint) { }
func (inventoryItem InventoryItem) GetAccountId() uint { return inventoryItem.AccountId }
func (inventoryItem *InventoryItem) setAccountId(id uint) { inventoryItem.AccountId = id }
func (InventoryItem) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (inventoryItem InventoryItem) create() (Entity, error)  {

	en := inventoryItem

	if err := db.Create(&en).Error; err != nil {
		return nil, err
	}

	err := en.GetPreloadDb(false, false, nil).First(&en, en.Id).Error
	if err != nil {
		return nil, err
	}

	var newItem Entity = &en

	return newItem, nil
}
func (InventoryItem) get(id uint, preloads []string) (Entity, error) {

	var inventoryItem InventoryItem

	err := inventoryItem.GetPreloadDb(false,false,preloads).First(&inventoryItem, id).Error
	if err != nil {
		return nil, err
	}
	return &inventoryItem, nil
}
func (inventoryItem *InventoryItem) load(preloads []string) error {

	err := inventoryItem.GetPreloadDb(false,false,preloads).First(inventoryItem, inventoryItem.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (inventoryItem *InventoryItem) loadByPublicId(preloads []string) error {
	return utils.Error{Message: "Невозможно загрузить Inventory Product по public Id"}
}
func (InventoryItem) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return InventoryItem{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (InventoryItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	inventoryItems := make([]InventoryItem, 0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&InventoryItem{}).GetPreloadDb(false,false,preloads).
			Joins("LEFT JOIN products ON products.id = inventory_items.product_id").
			Select("products.*, inventory_items.*").
			Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&inventoryItems, "inventory_id ILIKE ?", search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&InventoryItem{}).Joins("LEFT JOIN products ON products.id = inventory_items.product_id").
			Select("products.*, inventory_items.*").
			Where( "account_id = ?", accountId).Where(filter).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&InventoryItem{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&inventoryItems).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&InventoryItem{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(inventoryItems))
	for i := range inventoryItems {
		entities[i] = &inventoryItems[i]
	}

	return entities, total, nil
}
func (inventoryItem *InventoryItem) update(input map[string]interface{}, preloads []string) error {

	delete(input,"product")
	delete(input,"inventory")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","inventory_id","product_id"}); err != nil {
		return err
	}
	// input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := inventoryItem.GetPreloadDb(false,false,nil).Where(" id = ?", inventoryItem.Id).
		Omit("id", "account_id","created_at").Updates(input).Error; err != nil {
		return err
	}

	err := inventoryItem.GetPreloadDb(false,false,preloads).First(inventoryItem, inventoryItem.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (inventoryItem *InventoryItem) delete () error {
	return inventoryItem.GetPreloadDb(true,false,nil).Where("id = ?", inventoryItem.Id).Delete(inventoryItem).Error
}
// ######### END CRUD Functions ############

// Закрытие инвентаризации позиции
func (inventoryItem *InventoryItem) SetPosted () error {
	// 1. Определяем VolumeExpected
	volumeExpected := 0

	// Обновляем данные модели, закрывая от редактирования
	return inventoryItem.update(map[string]interface{}{"posted":true,"volume_expected": volumeExpected},nil)
}
