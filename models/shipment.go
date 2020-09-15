package models

import (
	"database/sql"
	"fmt"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Партия товаров (Поставка)
type Shipment struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	ParentId	*uint 	`json:"parent_id"`

	// Краткое имя поставки
	Name 		*string	`json:"name" gorm:"type:varchar(255);"`

	// External code
	Code 		*string	`json:"code" gorm:"type:varchar(255);"`

	// Поставщик (может быть не известен)
	CompanyId 	*uint	`json:"company_id" gorm:"type:int;index;"`

	// Склад назначения
	WarehouseId *uint	`json:"warehouse_id" gorm:"type:int;index;"`

	// Статус поставки: планируется, ожидается поставка, ожидает оприходования, завершена/отмена/фейл.
	Status 				WorkStatus 	`json:"status" gorm:"type:varchar(18);default:'pending'"`
	DecryptionStatus	*string 	`json:"decryption_status" gorm:"type:varchar(255);"`

	// Планируемая дата поставки
	DeliveryDate	*time.Time 	`json:"delivery_date"`

	// Сумма поставки - высчитывается в AfterFind
	PaymentAmountOrder	float64 `json:"_payment_amount_order" gorm:"-"`
	PaymentAmountFact	float64 `json:"_payment_amount_fact" gorm:"-"`
	// Товарных позиций - высчитывается
	ProductUnits	uint 	`json:"_product_units" gorm:"-"`

	// Фактический список товаров в поставке + объем + закупочная цены
	ShipmentItems 	[]ShipmentItem 	`json:"shipment_items"`

	// Список товаров в поставке
	Products 	[]Product 	`json:"products" gorm:"many2many:shipment_items"`

	Company 	Company 	`json:"company"`
	Warehouse 	Warehouse 	`json:"warehouse"`

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

func (Shipment) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&Shipment{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE shipments " +
		"ADD CONSTRAINT shipments_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT shipments_company_id_fkey FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL ON UPDATE CASCADE," +
		"ADD CONSTRAINT shipments_warehouse_id_fkey FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE SET NULL ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&Shipment{}, "Products", &ShipmentItem{})
	if err != nil {
		log.Fatal(err)
	}
}
func (shipment *Shipment) BeforeCreate(tx *gorm.DB) error {
	shipment.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&Shipment{}).Where("account_id = ?",  shipment.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	shipment.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (shipment *Shipment) AfterFind(tx *gorm.DB) (err error) {

	stat := struct {
		AmountOrder float64
		AmountFact float64
		Units uint
	}{0,0,0}
	if err = db.Raw("SELECT\n    COUNT(*) AS units,\n    sum(payment_amount * volume_order) AS amount_order,\n    sum(payment_amount * volume_fact) AS amount_fact\nFROM shipment_items\nWHERE account_id = ? AND shipment_id = ?;",
		shipment.AccountId, shipment.Id).
		Scan(&stat).Error; err != nil {
		return err
	}
	shipment.PaymentAmountOrder = stat.AmountOrder
	shipment.PaymentAmountFact 	= stat.AmountFact
	shipment.ProductUnits = stat.Units

	return nil
}
func (shipment *Shipment) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(shipment)
	} else {
		_db = _db.Model(&Shipment{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Product","Warehouse","ShipmentItems","ShipmentItems.Product"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}

// ############# Entity interface #############
func (shipment Shipment) GetId() uint { return shipment.Id }
func (shipment *Shipment) setId(id uint) { shipment.Id = id }
func (shipment *Shipment) setPublicId(publicId uint) { shipment.PublicId = publicId }
func (shipment Shipment) GetAccountId() uint { return shipment.AccountId }
func (shipment *Shipment) setAccountId(id uint) { shipment.AccountId = id }
func (Shipment) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (shipment Shipment) create() (Entity, error)  {

	en := shipment

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
func (Shipment) get(id uint, preloads []string) (Entity, error) {

	var shipment Shipment

	err := shipment.GetPreloadDb(false,false,preloads).First(&shipment, id).Error
	if err != nil {
		return nil, err
	}
	return &shipment, nil
}
func (shipment *Shipment) load(preloads []string) error {

	err := shipment.GetPreloadDb(false,false,preloads).First(shipment, shipment.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (shipment *Shipment) loadByPublicId(preloads []string) error {
	
	if shipment.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Shipment - не указан  Id"}
	}
	if err := shipment.GetPreloadDb(false,false, preloads).First(shipment, "account_id = ? AND public_id = ?", shipment.AccountId, shipment.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (Shipment) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Shipment{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (Shipment) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	shipments := make([]Shipment,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&Shipment{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&shipments, "name ILIKE ?", search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Shipment{}).
			Where("account_id = ? AND name ILIKE ?", accountId, search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&Shipment{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&shipments).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Shipment{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(shipments))
	for i := range shipments {
		entities[i] = &shipments[i]
	}

	return entities, total, nil
}
func (shipment *Shipment) update(input map[string]interface{}, preloads []string) error {

	delete(input,"products")
	delete(input,"shipment_items")
	delete(input,"company")
	delete(input,"warehouse")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","company_id","warehouse_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"delivery_date"})

	if err := shipment.GetPreloadDb(false,false,nil).Where(" id = ?", shipment.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := shipment.GetPreloadDb(false,false,preloads).First(shipment, shipment.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (shipment *Shipment) delete () error {
	return shipment.GetPreloadDb(true,false,nil).Where("id = ?", shipment.Id).Delete(shipment).Error
}
// ######### END CRUD Functions ############

// завершает поставку и переводит все товары в целевой warehouse

// Была ли оприходована поставка или нет
func (shipment Shipment) IsPostedWarehouse() bool {

	if shipment.Id < 1 {
		return false
	}

	return shipment.Status == WorkStatusCompleted
}
func (shipment Shipment) Validate() error {
	return nil
}

func (shipment *Shipment) updateWorkStatus(status WorkStatus, reason... string) error {
	_reason := ""
	if len(reason) > 0 {
		_reason = reason[0]
	}

	return shipment.update(map[string]interface{}{
		"status":	status,
		"decryption_status": _reason,
	},nil)
}
func (shipment *Shipment) SetPendingStatus() error {

	// Возможен вызов из состояния planned: вернуть на доработку => pending
	if shipment.Status != WorkStatusPlanned {
		reason := "Невозможно установить статус,"
		switch shipment.Status {
		case WorkStatusPending:
			reason += "т.к. поставка уже в разработке"
		case WorkStatusActive:
			reason += "т.к. поставка в процессе рассылки"
		case WorkStatusPaused:
			reason += "т.к. поставка на паузе, но в процессе рассылки"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusCancelled:
			reason += "т.к. поставка отменена"
		}
		return utils.Error{Message: reason}
	}

	return shipment.updateWorkStatus(WorkStatusPending)
}
func (shipment *Shipment) SetPlannedStatus() error {

	// Возможен вызов из состояния pending: запланировать кампанию => planned
	if shipment.Status != WorkStatusPending  {
		reason := "Невозможно запланировать кампанию,"
		switch shipment.Status {
		case WorkStatusPlanned:
			reason += "т.к. поставка уже в плане"
		case WorkStatusActive:
			reason += "т.к. поставка уже в процессе рассылки"
		case WorkStatusPaused:
			reason += "т.к. поставка на паузе, но уже в процессе рассылки"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. поставка отменена"
		}
		return utils.Error{Message: reason}
	}

	// Проверяем кампанию и шаблон, чтобы не ставить в план не рабочую кампанию.
	if err := shipment.Validate(); err != nil { return err  }

	// Переводим в состояние "Запланирована", т.к. все проверки пройдены и можно ставить ее в планировщик
	return shipment.updateWorkStatus(WorkStatusPlanned)
}
func (shipment *Shipment) SetActiveStatus() error {

	// Возможен вызов из состояния planned или paused: запустить кампанию => active
	if shipment.Status != WorkStatusPlanned && shipment.Status != WorkStatusPaused {
		reason := "Невозможно запустить кампанию,"
		switch shipment.Status {
		case WorkStatusPending:
			reason += "т.к. поставка еще в стадии разработки"
		case WorkStatusActive:
			reason += "т.к. поставка уже в процессе рассылки"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. поставка отменена"
		}
		return utils.Error{Message: reason}
	}

	// Снова проверяем кампанию и шаблон
	if err := shipment.Validate(); err != nil { return err  }

	// Переводим в состояние "Активна", т.к. все проверки пройдены и можно продолжить ее выполнение
	return shipment.updateWorkStatus(WorkStatusActive)
}
func (shipment *Shipment) SetPausedStatus() error {

	// Возможен вызов из состояния active: приостановить кампанию => paused
	if shipment.Status != WorkStatusActive {
		reason := "Невозможно приостановить кампанию,"
		switch shipment.Status {
		case WorkStatusPending:
			reason += "т.к. поставка еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. поставка уже в стадии планирования"
		case WorkStatusPaused:
			reason += "т.к. поставка уже приостановлена"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. поставка отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Приостановлена", т.к. все проверки пройдены и можно приостановить кампанию
	return shipment.updateWorkStatus(WorkStatusPaused)
}
func (shipment *Shipment) SetCompletedStatus() error {

	// Возможен вызов из состояния active, paused: завершить кампанию => completed
	// Сбрасываются все задачи из очереди
	if shipment.Status != WorkStatusActive && shipment.Status != WorkStatusPaused {
		reason := "Невозможно завершить кампанию,"
		switch shipment.Status {
		case WorkStatusPending:
			reason += "т.к. поставка еще в стадии разработки"
		case WorkStatusPlanned:
			reason += "т.к. поставка еще в стадии планирования"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. поставка отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return shipment.updateWorkStatus(WorkStatusCompleted)
}
func (shipment *Shipment) SetFailedStatus(reason string) error {
	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return shipment.updateWorkStatus(WorkStatusFailed, reason)
}
func (shipment *Shipment) SetCancelledStatus() error {

	// Возможен вызов из состояния active, paused, planned: завершить кампанию => cancelled
	if shipment.Status != WorkStatusActive && shipment.Status != WorkStatusPaused && shipment.Status != WorkStatusPlanned {
		reason := "Невозможно отменить кампанию,"
		switch shipment.Status {
		case WorkStatusPending:
			reason += "т.к. поставка еще в стадии разработки"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. поставка уже отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return shipment.updateWorkStatus(WorkStatusCancelled)
}
func (shipment *Shipment) SetShipmentStatus() error {

	// Возможен вызов из состояния active, paused, planned: завершить кампанию => cancelled
	if shipment.Status != WorkStatusActive && shipment.Status != WorkStatusPaused && shipment.Status != WorkStatusPlanned {
		reason := "Невозможно отменить кампанию,"
		switch shipment.Status {
		case WorkStatusPending:
			reason += "т.к. поставка еще в стадии разработки"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. поставка уже отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return shipment.updateWorkStatus(WorkStatusCancelled)
}
func (shipment *Shipment) SetPostingStatus() error {

	// Возможен вызов из состояния active, paused, planned: завершить кампанию => cancelled
	if shipment.Status != WorkStatusActive && shipment.Status != WorkStatusPaused && shipment.Status != WorkStatusPlanned {
		reason := "Невозможно отменить кампанию,"
		switch shipment.Status {
		case WorkStatusPending:
			reason += "т.к. поставка еще в стадии разработки"
		case WorkStatusCompleted:
			reason += "т.к. поставка уже завершена"
		case WorkStatusFailed:
			reason += "т.к. поставка завершена с ошибкой"
		case WorkStatusCancelled:
			reason += "т.к. поставка уже отменена"
		}
		return utils.Error{Message: reason}
	}

	// Переводим в состояние "Завершена", т.к. все проверки пройдены и можно приостановить кампанию
	return shipment.updateWorkStatus(WorkStatusCancelled)
}

// ######### Shipment Functions ############

// Добавить новую позицию товара
func (shipment Shipment) AppendProduct(product Product, volumeOrder, paymentAmount float64) error {

	fmt.Println("Добавляем товар")
	// 1. Проверяем статус поставки
	if shipment.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: не удается определить поставку"}
	}
	if err := shipment.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: не удается загрузить поставку"}
	}
	// нельзя добавить товар, если поставка завершена, закончена или отменена
	if shipment.Status == WorkStatusCompleted || shipment.Status == WorkStatusFailed || shipment.Status == WorkStatusCancelled {
		return utils.Error{Message: "Нельзя добавить товар, если поставка завершена, закончена или отменена"}
	}

	// 2. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, он не найден"}
	}

	// 3. Проверяем есть ли уже на этом складе этот продукт
	if shipment.ExistProduct(product.Id) {
		return nil
		// return utils.Error{Message: "Товар уже есть в текущей поставки"}
	}

	// 4. Добавляем запись в ShipmentProduct
	if err := db.Create(
		&ShipmentItem{AccountId: shipment.AccountId, ProductId: product.Id, ShipmentId: shipment.Id, VolumeOrder: volumeOrder, VolumeFact: 0, PaymentAmount: paymentAmount, WarehousePosted: false}).Error; err != nil {
		return err
	}

	// 5. Запускаем событие добавление товара в инвентаризацию (надо ли)
	account, err := GetAccount(shipment.AccountId)
	if err == nil && account != nil {
		event.AsyncFire(Event{}.InventoryItemProductAppended(account.Id, shipment.Id, product.Id))
	}

	return nil
}
// Удалить из поставки позицию товара
func (shipment Shipment) RemoveProduct(product Product) error {

	// 1. Проверяем статус поставки
	if shipment.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: не удается определить поставку"}
	}
	if err := shipment.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: не удается загрузить поставку"}
	}
	// нельзя удалить товар, если поставка завершена, закончена или отменена
	if shipment.Status == WorkStatusCompleted || shipment.Status == WorkStatusFailed || shipment.Status == WorkStatusCancelled {
		return utils.Error{Message: "Нельзя удалить товар из поставки, если поставка завершена, закончена или отменена"}
	}

	// 2. Загружаем продукт еще раз
	if err := product.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить продукт, он не найден"}
	}

	if shipment.AccountId < 1 || product.Id < 1 || shipment.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product id || shipment id == nil"}
	}

	if err := db.Where("account_id = ? AND product_id = ? AND shipment_id = ?", shipment.AccountId, product.Id, shipment.Id).Delete(
		&ShipmentItem{}).Error; err != nil {
		return err
	}


	return nil
}

func (shipment Shipment) ExistProduct(productId uint) bool {

	if productId < 1 {
		return false
	}

	sp := ShipmentItem{}
	result := db.Model(&ShipmentItem{}).First(&sp,"shipment_id = ? AND product_id = ?", shipment.Id, productId)

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
