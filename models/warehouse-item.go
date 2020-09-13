package models

import (
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Складская единица учета на складе
type WarehouseItem struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	ProductId 	uint	`json:"product_id" gorm:"type:int;index;"`
	WarehouseId uint	`json:"warehouse_id" gorm:"type:int;index;"`

	// Сколько ед. в одном товаре ()
	AmountUnit 	float64 `json:"amount_unit" gorm:"type:numeric;"`

	// Остаток
	Stock 		float64 `json:"stock" gorm:"type:numeric;"`

	// Резерв
	Reservation	float64 `json:"reservation" gorm:"type:numeric;"`

	// Время хранения... - потом вызов уведомления (?)
	ExpiredAt 	time.Time  `json:"expired_at"`

	Product 	Product 	`json:"product"`
	Warehouse 	Warehouse 	`json:"warehouse"`

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

func (WarehouseItem) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&WarehouseItem{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE warehouse_items " +
		"ADD CONSTRAINT warehouse_items_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT warehouse_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT warehouse_items_web_site_id_fkey FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (warehouseItem *WarehouseItem) BeforeCreate(tx *gorm.DB) error {
	warehouseItem.Id = 0
	if warehouseItem.AccountId < 1 {
		return utils.Error{Message: "Техническая ошибка создания складской единицы: отсутствует account id"}
	}
	return nil
}
func (warehouseItem *WarehouseItem) AfterFind(tx *gorm.DB) (err error) {
	return nil
}
func (warehouseItem *WarehouseItem) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(warehouseItem)
	} else {
		_db = _db.Model(&WarehouseItem{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Product","Warehouse"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}

// ############# Entity interface #############
func (warehouseItem WarehouseItem) GetId() uint { return warehouseItem.Id }
func (warehouseItem *WarehouseItem) setId(id uint) { warehouseItem.Id = id }
func (warehouseItem *WarehouseItem) setPublicId(publicId uint) { }
func (warehouseItem WarehouseItem) GetAccountId() uint { return warehouseItem.AccountId }
func (warehouseItem *WarehouseItem) setAccountId(id uint) { warehouseItem.AccountId = id }
func (WarehouseItem) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (warehouseItem WarehouseItem) create() (Entity, error)  {

	en := warehouseItem

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
func (WarehouseItem) get(id uint, preloads []string) (Entity, error) {

	var warehouseItem WarehouseItem

	err := warehouseItem.GetPreloadDb(false,false,preloads).First(&warehouseItem, id).Error
	if err != nil {
		return nil, err
	}
	return &warehouseItem, nil
}
func (warehouseItem *WarehouseItem) load(preloads []string) error {

	err := warehouseItem.GetPreloadDb(false,false,preloads).First(warehouseItem, warehouseItem.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (warehouseItem *WarehouseItem) loadByPublicId(preloads []string) error {
	return utils.Error{Message: "Невозможно получить объект по публичному ID"}
}
func (WarehouseItem) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return WarehouseItem{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (WarehouseItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	warehouseItems := make([]WarehouseItem,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&WarehouseItem{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).
			Where(filter).Find(&warehouseItems, "account_id = ?", accountId).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WarehouseItem{}).Where("account_id = ?", accountId).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&WarehouseItem{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&warehouseItems).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WarehouseItem{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(warehouseItems))
	for i := range warehouseItems {
		entities[i] = &warehouseItems[i]
	}

	return entities, total, nil
}
func (warehouseItem *WarehouseItem) update(input map[string]interface{}, preloads []string) error {

	delete(input,"product")
	delete(input,"warehouse")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"product_id","warehouse_id","web_site_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := warehouseItem.GetPreloadDb(false,false,nil).Where(" id = ?", warehouseItem.Id).
		Omit("id", "account_id").Updates(input).Error; err != nil {
		return err
	}

	err := warehouseItem.GetPreloadDb(false,false,preloads).First(warehouseItem, warehouseItem.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (warehouseItem *WarehouseItem) delete () error {
	return warehouseItem.GetPreloadDb(true,false,nil).Where("id = ?", warehouseItem.Id).Delete(warehouseItem).Error
}
// ######### END CRUD Functions ############
