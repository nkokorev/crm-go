package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type Warehouse struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;"`

	// Код склада (может быть external id)
	Code 		*string `json:"code" gorm:"type:varchar(255);"`

	// Имя склада
	Name 		*string `json:"name" gorm:"type:varchar(255);"`

	// Контактные данные склада
	Address 	*string `json:"address" gorm:"type:varchar(255);"`
	Phone	 	*string `json:"phone" gorm:"type:varchar(50);"`
	Email	 	*string `json:"email" gorm:"type:varchar(50);"`

	// Описание склада
	Description	*string `json:"description" gorm:"type:varchar(255);"`

	ProductCount	uint `json:"_product_count" gorm:"-"`

	WarehouseItems	[]WarehouseItem `json:"warehouse_items"`
	Products		[]Product 		`json:"products" gorm:"many2many:warehouse_item;"`

	CreatedAt 	time.Time 		`json:"created_at"`
	UpdatedAt 	time.Time 		`json:"updated_at"`
	DeletedAt 	gorm.DeletedAt 	`json:"-" sql:"index"`
}

func (Warehouse) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&Warehouse{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE warehouses " +
		"ADD CONSTRAINT warehouses_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&Warehouse{}, "Products", &WarehouseItem{})
	if err != nil {
		log.Fatal(err)
	}
}
func (warehouse *Warehouse) BeforeCreate(tx *gorm.DB) error {
	warehouse.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&Warehouse{}).Where("account_id = ?",  warehouse.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	warehouse.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (warehouse *Warehouse) AfterFind(tx *gorm.DB) (err error) {
	warehouse.ProductCount =  uint(db.Model(warehouse).Association("Products").Count())
	return nil
}
func (warehouse *Warehouse) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = db.Model(warehouse)
	} else {
		_db = db.Model(&Warehouse{})
	}

	if autoPreload {
		return db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"WarehouseItems","WarehouseItems.Product","Products"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}

// ############# Entity interface #############
func (warehouse Warehouse) GetId() uint { return warehouse.Id }
func (warehouse *Warehouse) setId(id uint) { warehouse.Id = id }
func (warehouse *Warehouse) setPublicId(publicId uint) { warehouse.PublicId = publicId }
func (warehouse Warehouse) GetAccountId() uint { return warehouse.AccountId }
func (warehouse *Warehouse) setAccountId(id uint) { warehouse.AccountId = id }
func (Warehouse) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (warehouse Warehouse) create() (Entity, error)  {

	en := warehouse

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
func (Warehouse) get(id uint, preloads []string) (Entity, error) {

	var warehouse Warehouse

	err := warehouse.GetPreloadDb(false,false,preloads).First(&warehouse, id).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}
func (warehouse *Warehouse) load(preloads []string) error {

	err := warehouse.GetPreloadDb(false,false,preloads).First(warehouse, warehouse.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (warehouse *Warehouse) loadByPublicId(preloads []string) error {

	if warehouse.PublicId < 1 || warehouse.AccountId < 1 {
		return utils.Error{Message: "Невозможно загрузить Warehouse - не указан  Id"}
	}
	if err := warehouse.GetPreloadDb(false,false, preloads).
		First(warehouse, "account_id = ? AND public_id = ?", warehouse.AccountId, warehouse.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (Warehouse) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Warehouse{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (Warehouse) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	warehouses := make([]Warehouse,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&Warehouse{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&warehouses, "name ILIKE ? OR code ILIKE ? OR address ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR description ILIKE ?", search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Warehouse{}).
			Where("name ILIKE ? OR code ILIKE ? OR address ILIKE ? OR phone ILIKE ? OR email ILIKE ? OR description ILIKE ?", search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&Warehouse{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&warehouses).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Warehouse{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(warehouses))
	for i := range warehouses {
		entities[i] = &warehouses[i]
	}

	return entities, total, nil
}
func (warehouse *Warehouse) update(input map[string]interface{}, preloads []string) error {

	delete(input,"products")
	delete(input,"warehouse_items")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}

	if err := warehouse.GetPreloadDb(false,false,nil).Where(" id = ?", warehouse.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := warehouse.GetPreloadDb(false,false,preloads).First(warehouse, warehouse.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (warehouse *Warehouse) delete () error {
	return warehouse.GetPreloadDb(true,false,nil).Where("id = ?", warehouse.Id).Delete(warehouse).Error
}
// ######### END CRUD Functions ############

// ######### WAREHOUSE ACTIONS ############

// Оприходовать поставку = завести все реальные товары из поставки на склад. Создавать ли позиции, которых нет?
func (warehouse *Warehouse) Shipment(shipment Shipment, createIfNotExist bool) error {

	// 1. Загружаем продукт еще раз
	if err := shipment.load([]string{"ShipmentProduct","ShipmentProduct.Product"}); err != nil {
		return utils.Error{Message: "Техническая ошибка оприходования поставки: поставка не найдена"}
	}

	// 2. Проверяем, была ли занесена эта поставка
	// if !warehouse.ExistShipment(shipment.Id) {
	if shipment.IsPostedWarehouse() {
		return utils.Error{Message: "Поставка уже занесена на склад"}
	}

	// Идем циклом по товаром с проверкой быи ли они занесены
	for i := range shipment.ShipmentItems {
		item := shipment.ShipmentItems[i]

		// Если товар был уже занесен
		if item.WarehousePosted {
			continue
		}

		// Проверяем, если товар на складе
		if !warehouse.ExistProduct(item.ProductId) {
			if err := warehouse.AppendProduct(item.Product); err != nil {
				log.Printf("Ошибка создания товара на складе: %v\n", err)
				return err
			}
		}

		// Пополняем запас товара по факту прихода
		if err := warehouse.ReplenishProduct(item.ProductId, item.VolumeFact); err != nil {
			log.Printf("Ошибка пополнения товара на складе: %v\n", err)
			return err
		}

		if err := item.SetWarehousePosted(); err != nil {
			log.Printf("Ошибка перевода позиции поставки в добавленную на склад: %v\n", err)
			return err
		}
	}

	if err := shipment.SetCompletedStatus(); err != nil {
		log.Printf("Ошибка перевода поставки в статус завершения: %v\n", err)
		return err
	}

	// todo: нужно потом добавить событие
	/*account, err := GetAccount(warehouse.AccountId)
	if err == nil && account != nil {
		AsyncFire(*Event{}.WarehouseItemProductAppended(account.Id, warehouse.Id, product.Id))
	}*/

	return nil
}

// Добавить новую позицию товара
func (warehouse Warehouse) AppendProduct(product Product, strict... bool) error {

	// 1. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, он не найден"}
	}
	
	// 2. Проверяем есть ли уже на этом складе этот продукт
	if warehouse.ExistProduct(product.Id) {
		if len(strict) > 1 {
			return utils.Error{Message: "Продукт уже числиться на складе"}
		} else {
			return nil
		}
	}
	
	if err := db.Model(&WarehouseItem{}).Create(
		&WarehouseItem{AccountId: warehouse.AccountId, ProductId: product.Id, WarehouseId: warehouse.Id, Stock: 0, Reservation: 0}).Error; err != nil {
		return err
	}

	account, err := GetAccount(warehouse.AccountId)
	if err == nil && account != nil {
		// AsyncFire(*Event{}.WarehouseItemProductAppended(account.Id, warehouse.Id, product.Id))
		// AsyncFire(NewEvent("UserCreated", map[string]interface{}{"account_id":user.IssuerAccountId, "user_id":user.Id}))
	}

	return nil
}
// Удалить со склада позицию товара
func (warehouse Warehouse) RemoveProduct(product Product) error {

	// 1. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя удалить продукт, он не найден"}
	}

	if warehouse.AccountId < 1 || product.Id < 1 || warehouse.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product id || warehouse id == nil"}
	}

	if err := db.Where("account_id = ? AND product_id = ? AND warehouse_id = ?", warehouse.AccountId, product.Id, warehouse.Id).Delete(
		&WarehouseItem{}).Error; err != nil {
		return err
	}


	return nil
}

// Потратить N у продукта
func (warehouse Warehouse) SpendProduct(productId uint, amount float64) error {

	return nil
}
// Пополнить N продукта
func (warehouse Warehouse) ReplenishProduct(productId uint, amount float64) error {

	return nil
}

// Забронировать продукцию
func (warehouse Warehouse) BookProduct(productId uint, amount float64) error {

	return nil
}
// Снять бронь у продукции
func (warehouse Warehouse) CancelReservationProduct(productId uint, amount float64) error {

	return nil
}

func (warehouse Warehouse) ExistProduct(productId uint) bool {

	if productId < 1 {
		return false
	}
	
	var count int64

	db.Model(&WarehouseItem{}).Where("warehouse_id = ? AND product_id = ?", warehouse.Id, productId).Count(&count)
	if count > 0 {
		return true
	}
	
	return false
}

func (warehouse Warehouse) GetProductById(productId uint) (*WarehouseItem, error) {

	if productId < 1 {
		return nil, utils.Error{Message: "Невозможно получить данные по товару: отсутствует id"}
	}

	warehouseItem, err := warehouse.GetByProductId(productId, nil)
	if err != nil {
        return nil, err
	}

	return warehouseItem, nil
}

// todo: поставки.. и инвентаризация

// ######### END CRUD Functions ############