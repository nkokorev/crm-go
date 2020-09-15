package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Инвентаризация товаров
type Inventory struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Склад назначения (обязательный параметр для запуска)
	WarehouseId *uint	`json:"warehouse_id" gorm:"type:int;index;"`

	// Поставщик (может быть не известен)
	CompanyId 	*uint	`json:"company_id" gorm:"type:int;"`

	// Краткое имя инвентаризации: "Пересчет перед НГ", "Просчет подарков"
	Name 		*string	`json:"name" gorm:"type:varchar(255);"`

	// Статус поставки: планируется, ожидается инвентаризация, ожидает оприходования, завершена/отмена/фейл.
	Status 				WorkStatus 	`json:"status" gorm:"type:varchar(18);default:'pending'"`
	DecryptionStatus	*string 	`json:"decryption_status" gorm:"type:varchar(255);"`

	// Время (возможно планируемое) инвентаризации
	StartDate	*time.Time 	`json:"start_date"`
	EndDate		*time.Time 	`json:"end_date"`

	// Примерная сумма расхождения по числиться / факт - высчитывается в AfterFind  
	DifferenceAmount	float64 `json:"_difference_amount" gorm:"-"`
	// Число инвентаризированных товарных позиций - высчитывается
	ProductUnits	uint 	`json:"_product_units" gorm:"-"`

	// Фактический список товаров в инвентаризации + объем факт / ожидание (его фиксируем в момент записи)
	InventoryItem 	[]InventoryItem 	`json:"inventory_items"`
	Products 	[]Product 	`json:"products" gorm:"many2many:inventory_items"`

	Company 	Company 	`json:"company"`
	Warehouse 	Warehouse 	`json:"warehouse"`

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

func (Inventory) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&Inventory{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE inventories " +
		"ADD CONSTRAINT inventories_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT inventories_company_id_fkey FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT inventories_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&Inventory{}, "Products", &InventoryItem{})
	if err != nil {
		log.Fatal(err)
	}
}
func (inventory *Inventory) BeforeCreate(tx *gorm.DB) error {
	inventory.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&Inventory{}).Where("account_id = ?",  inventory.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	inventory.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (inventory *Inventory) AfterFind(tx *gorm.DB) (err error) {

	return nil
}
func (inventory *Inventory) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(inventory)
	} else {
		_db = _db.Model(&Inventory{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"InventoryItems","Products","InventoryItems.Product","Company","Warehouse"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}

// ############# Entity interface #############
func (inventory Inventory) GetId() uint { return inventory.Id }
func (inventory *Inventory) setId(id uint) { inventory.Id = id }
func (inventory *Inventory) setPublicId(publicId uint) { inventory.PublicId = publicId }
func (inventory Inventory) GetAccountId() uint { return inventory.AccountId }
func (inventory *Inventory) setAccountId(id uint) { inventory.AccountId = id }
func (Inventory) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (inventory Inventory) create() (Entity, error)  {

	en := inventory

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
func (Inventory) get(id uint, preloads []string) (Entity, error) {

	var inventory Inventory

	err := inventory.GetPreloadDb(false,false,preloads).First(&inventory, id).Error
	if err != nil {
		return nil, err
	}
	return &inventory, nil
}
func (inventory *Inventory) load(preloads []string) error {

	err := inventory.GetPreloadDb(false,false,preloads).First(inventory, inventory.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (inventory *Inventory) loadByPublicId(preloads []string) error {
	
	if inventory.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Inventory - не указан  Id"}
	}
	if err := inventory.GetPreloadDb(false,false, preloads).First(inventory, "account_id = ? AND public_id = ?", inventory.AccountId, inventory.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (Inventory) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Inventory{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (Inventory) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	inventories := make([]Inventory,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&Inventory{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&inventories, "name ILIKE ?", search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Inventory{}).
			Where("account_id = ? AND name ILIKE ?", accountId, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&Inventory{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&inventories).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Inventory{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(inventories))
	for i := range inventories {
		entities[i] = &inventories[i]
	}

	return entities, total, nil
}
func (inventory *Inventory) update(input map[string]interface{}, preloads []string) error {

	delete(input,"image")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","parent_id","company_id","warehouse_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"delivery_date"})

	if err := inventory.GetPreloadDb(false,false,nil).Where(" id = ?", inventory.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := inventory.GetPreloadDb(false,false,preloads).First(inventory, inventory.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (inventory *Inventory) delete () error {
	return inventory.GetPreloadDb(true,false,nil).Where("id = ?", inventory.Id).Delete(inventory).Error
}
// ######### END CRUD Functions ############

// завершает инвентаризацию и переводит все товары в целевой warehouse

// Была ли оприходована инвентаризация или нет
func (inventory Inventory) IsPostedWarehouse() bool {

	if inventory.Id < 1 {
		return false
	}

	return inventory.Status == WorkStatusCompleted
}
func (inventory Inventory) Validate() error {
	return nil
}

func (inventory *Inventory) updateWorkStatus(status WorkStatus, reason... string) error {
	_reason := ""
	if len(reason) > 0 {
		_reason = reason[0]
	}

	return inventory.update(map[string]interface{}{
		"status":	status,
		"decryption_status": _reason,
	},nil)
}
func (inventory *Inventory) SetPendingStatus() error {

	// Возможен вызов из состояния planned: вернуть на доработку => pending
	if inventory.Status != WorkStatusPlanned {
		reason := "Невозможно установить статус,"
		switch inventory.Status {
		case WorkStatusPending:
			reason += "т.к. инвентаризация уже в разработке"
		case WorkStatusActive:
			reason += "т.к. инвентаризация в процессе рассылки"
		case WorkStatusPaused:
			reason += "т.к. инвентаризация на паузе, но в процессе рассылки"
		case WorkStatusFailed:
			reason += "т.к. инвентаризация завершена с ошибкой"
		case WorkStatusCompleted:
			reason += "т.к. инвентаризация уже завершена"
		case WorkStatusCancelled:
			reason += "т.к. инвентаризация отменена"
		}
		return utils.Error{Message: reason}
	}

	return inventory.updateWorkStatus(WorkStatusPending)
}
func (inventory *Inventory) SetPlannedStatus() error {

	// Возможен вызов из состояния pending: запланировать кампанию => planned
	if inventory.Status != WorkStatusPending  {
		reason := "Невозможно запланировать кампанию,"
		switch inventory.Status {
		case WorkStatusPlanned:
			reason += "т.к. инвентаризация уже в плане"
		case WorkStatusActive:
			reason += "т.к. инвентаризация уже в процессе рассылки"
		case WorkStatusPaused:
			reason += "т.к. инвентаризация на паузе, но уже в процессе рассылки"
		case WorkStatusCompleted:
			reason += "т.к. инвентаризация уже завершена"
		case WorkStatusFailed:
			reason += "т.к. инвентаризация завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. инвентаризация отменена"
		}
		return utils.Error{Message: reason}
	}

	// Проверяем кампанию и шаблон, чтобы не ставить в план не рабочую кампанию.
	if err := inventory.Validate(); err != nil { return err  }

	// Переводим в состояние "Запланирована", т.к. все проверки пройдены и можно ставить ее в планировщик
	return inventory.updateWorkStatus(WorkStatusPlanned)
}
func (inventory *Inventory) SetActiveStatus() error {

	// Возможен вызов из состояния planned или paused: запустить кампанию => active
	if inventory.Status != WorkStatusPlanned && inventory.Status != WorkStatusPaused {
		reason := "Невозможно запустить кампанию,"
		switch inventory.Status {
		case WorkStatusPending:
			reason += "т.к. инвентаризация еще в стадии разработки"
		case WorkStatusActive:
			reason += "т.к. инвентаризация уже в процессе рассылки"
		case WorkStatusCompleted:
			reason += "т.к. инвентаризация уже завершена"
		case WorkStatusFailed:
			reason += "т.к. инвентаризация завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. инвентаризация отменена"
		}
		return utils.Error{Message: reason}
	}

	// Снова проверяем кампанию и шаблон
	if err := inventory.Validate(); err != nil { return err  }

	// Переводим в состояние "Активна", т.к. все проверки пройдены и можно продолжить ее выполнение
	return inventory.updateWorkStatus(WorkStatusActive)
}
func (inventory *Inventory) SetPausedStatus() error {

	// Возможен вызов из состояния active: приостановить кампанию => paused
	if inventory.Status != WorkStatusActive {
		reason := "Невозможно приостановить кампанию,"
		switch inventory.Status {
		case WorkStatusPending:
			reason += "т.к. инвентаризация еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. инвентаризация уже в стадии планирования"
		case WorkStatusPaused:
			reason += "т.к. инвентаризация уже приостановлена"
		case WorkStatusCompleted:
			reason += "т.к. инвентаризация уже завершена"
		case WorkStatusFailed:
			reason += "т.к. инвентаризация завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. инвентаризация отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Приостановлена", т.к. все проверки пройдены и можно приостановить кампанию
	return inventory.updateWorkStatus(WorkStatusPaused)
}
func (inventory *Inventory) SetCompletedStatus() error {

	// Возможен вызов из состояния active, paused: завершить кампанию => completed
	// Сбрасываются все задачи из очереди
	if inventory.Status != WorkStatusActive && inventory.Status != WorkStatusPaused {
		reason := "Невозможно завершить кампанию,"
		switch inventory.Status {
		case WorkStatusPending:
			reason += "т.к. инвентаризация еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. инвентаризация еще в стадии планирования"
		case WorkStatusCompleted:
			reason += "т.к. инвентаризация уже завершена"
		case WorkStatusFailed:
			reason += "т.к. инвентаризация завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. инвентаризация отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return inventory.updateWorkStatus(WorkStatusCompleted)
}
func (inventory *Inventory) SetFailedStatus(reason string) error {
	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return inventory.updateWorkStatus(WorkStatusFailed, reason)
}
func (inventory *Inventory) SetCancelledStatus() error {

	// Возможен вызов из состояния active, paused, planned: завершить кампанию => cancelled
	if inventory.Status != WorkStatusActive && inventory.Status != WorkStatusPaused && inventory.Status != WorkStatusPlanned {
		reason := "Невозможно отменить кампанию,"
		switch inventory.Status {
		case WorkStatusPending:
			reason += "т.к. инвентаризация еще в стадии разработки"
		case WorkStatusCompleted:
			reason += "т.к. инвентаризация уже завершена"
		case WorkStatusFailed:
			reason += "т.к. инвентаризация завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. инвентаризация уже отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return inventory.updateWorkStatus(WorkStatusCancelled)
}
func (inventory *Inventory) SetInventoryStatus() error {

	// Возможен вызов из состояния active, paused, planned: завершить кампанию => cancelled
	if inventory.Status != WorkStatusActive && inventory.Status != WorkStatusPaused && inventory.Status != WorkStatusPlanned {
		reason := "Невозможно отменить кампанию,"
		switch inventory.Status {
		case WorkStatusPending:
			reason += "т.к. инвентаризация еще в стадии разработки"
		case WorkStatusCompleted:
			reason += "т.к. инвентаризация уже завершена"
		case WorkStatusFailed:
			reason += "т.к. инвентаризация завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. инвентаризация уже отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return inventory.updateWorkStatus(WorkStatusCancelled)
}
func (inventory *Inventory) SetPostingStatus() error {

	// Возможен вызов из состояния active, paused, planned: завершить кампанию => cancelled
	if inventory.Status != WorkStatusActive && inventory.Status != WorkStatusPaused && inventory.Status != WorkStatusPlanned {
		reason := "Невозможно отменить кампанию,"
		switch inventory.Status {
		case WorkStatusPending:
			reason += "т.к. инвентаризация еще в стадии разработки"
		case WorkStatusCompleted:
			reason += "т.к. инвентаризация уже завершена"
		case WorkStatusFailed:
			reason += "т.к. инвентаризация завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. инвентаризация уже отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return inventory.updateWorkStatus(WorkStatusCancelled)
}

// ######### Inventory Functions ############

// Добавить новую позицию товара
func (inventory Inventory) AppendProduct(product Product, volumeFact float64) error {

	// 1. Проверяем статус инвентаризации
	if inventory.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: не удается определить инвентаризацию"}
	}
	if err := inventory.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: не удается загрузить инвентаризацию"}
	}
	// нельзя добавить товар, если инвентаризация завершена, закончена или отменена
	if inventory.Status == WorkStatusCompleted || inventory.Status == WorkStatusFailed || inventory.Status == WorkStatusCancelled {
		return utils.Error{Message: "Нельзя добавить товар, если инвентаризация завершена, закончена или отменена"}
	}

	// 2. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, он не найден"}
	}

	// 3. Проверяем есть ли уже на этом складе этот продукт
	if inventory.ExistProduct(product.Id) {
		return utils.Error{Message: "Товар уже есть в текущей поставки"}
	}

	// 4. Узнаем текущий остаток на складе
	volumeExpected := float64(0)
	if inventory.WarehouseId != nil {
		warehouse := Warehouse{Id: *inventory.WarehouseId}
		if err := warehouse.load(nil); err != nil {
			return err
		}

		warehouseItem, err := warehouse.GetProductById(product.Id)
		if err != nil {
			return err
		}

		volumeExpected = warehouseItem.Stock
	}


	// 5. Добавляем запись в InventoryItem
	if err := db.Create(
		&InventoryItem{
			AccountId: inventory.AccountId, ProductId: product.Id, InventoryId: inventory.Id, VolumeFact: &volumeFact, Posted: false, VolumeExpected: &volumeExpected}).Error; err != nil {
		return err
	}

	// 5. Запускаем событие добавление товара в инвентаризацию (надо ли)
	account, err := GetAccount(inventory.AccountId)
	if err == nil && account != nil {
		event.AsyncFire(Event{}.InventoryItemProductAppended(account.Id, inventory.Id, product.Id))
	}

	return nil
}
// Удалить из поставки позицию товара
func (inventory Inventory) RemoveProduct(product Product) error {

	// 1. Проверяем статус поставки
	if inventory.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: не удается определить инвентаризацию"}
	}
	if err := inventory.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: не удается загрузить инвентаризацию"}
	}
	// нельзя удалить товар, если инвентаризация завершена, закончена или отменена
	if inventory.Status == WorkStatusCompleted || inventory.Status == WorkStatusFailed || inventory.Status == WorkStatusCancelled {
		return utils.Error{Message: "Нельзя удалить товар из поставки, если инвентаризация завершена, закончена или отменена"}
	}

	// 2. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, он не найден"}
	}

	if inventory.AccountId < 1 || product.Id < 1 || inventory.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product id || inventory id == nil"}
	}

	if err := db.Where("account_id = ? AND product_id = ? AND inventory_id = ?", inventory.AccountId, product.Id, inventory.Id).
		Delete(&WarehouseItem{}).Error; err != nil {
		return err
	}


	return nil
}

func (inventory Inventory) ExistProduct(productId uint) bool {

	if productId < 1 {
		return false
	}

	sp := InventoryItem{}
	result := db.Model(&InventoryItem{}).First(&sp,"inventory_id = ? AND product_id = ?", inventory.Id, productId)

	// check error ErrRecordNotFound
	// errors.Is(result.Error, gorm.ErrRecordNotFound)
	if result.Error != nil {
		return false
	}
	if result.RowsAffected > 0 {
		return true
	}


	return false
}
