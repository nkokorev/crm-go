package models

import (
	"errors"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Корзины товаров
type CartItem struct {
	
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint	`json:"account_id" gorm:"index;not null;"` // аккаунт-владелец ключа
	OrderId 	uint	`json:"order_id" gorm:"index;not null"` // заказ, к которому относится корзина

	ProductId	uint 	`json:"product_id" gorm:"type:int;not null;"`// Id позиции товара
	Description	string 	`json:"description" gorm:"type:varchar(255);not null;"` // краткое описание товара?
	Quantity	float64	`json:"quantity" gorm:"type:int;default:0;"`// число ед. товара

	// Фиксируем стоимость 
	Amount		PaymentAmount	`json:"amount" gorm:"-"`
	Cost		float64 		`json:"cost" gorm:"type:numeric;default:0"` // Результирующая стоимость 1 ед. товара!

	// Признак предмета расчета
	PaymentSubjectId		uint			`json:"payment_subject_id" gorm:"type:int;not null;default:1"`// товар или услуга ? [вид номенклатуры]
	PaymentSubject 			PaymentSubject 	`json:"payment_subject_yandex"`
	PaymentSubjectYandex	string 		`json:"payment_subject"` // << after find

	// Признак способа расчета
	PaymentModeId		uint	`json:"payment_mode_id" gorm:"type:int;not null;default:1"`//
	PaymentMode 		PaymentMode `json:"payment_mode_yandex"`
	PaymentModeYandex 	string `json:"payment_mode"`

	// Ставка НДС
	VatCode				uint	`json:"vat_code"` // << не vat_code_id т.к. в яндексе просто 'vat_code'

	// В резерве или списан товар: true - в резерве, false - списан
	Reserved			bool `json:"reserved" gorm:"type:bool;default:false;"`

	// Была ли списана ли позиция со склада
	Wasted			bool `json:"wasted" gorm:"type:bool;default:false;"`

	// warehouseItemId -   
	WarehouseId		*uint `json:"warehouse_id" gorm:"type:int;"`
	Warehouse		*Warehouse	`json:"warehouse"`
	WarehouseItems	[]WarehouseItem	`json:"warehouse_items" gorm:"-"` // AfterFind

	Product Product `json:"product"`
	Order	Order `json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CartItem) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&CartItem{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&CartItem{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&CartItem{}).AddForeignKey("order_id", "orders(id)", "CASCADE", "CASCADE")
	// db.Model(&CartItem{}).AddForeignKey("payment_subject_id", "payment_subjects(id)", "RESTRICT", "CASCADE")
	// db.Model(&CartItem{}).AddForeignKey("payment_mode_id", "payment_modes(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE cart_items " +
		"ADD CONSTRAINT cart_items_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT cart_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT cart_items_payment_subject_id_fkey FOREIGN KEY (payment_subject_id) REFERENCES payment_subjects(id) ON DELETE RESTRICT ON UPDATE CASCADE," +
		"DROP CONSTRAINT IF EXISTS fk_cart_items_product," +
		"ADD CONSTRAINT cart_items_payment_mode_id_fkey FOREIGN KEY (payment_mode_id) REFERENCES payment_modes(id) ON DELETE RESTRICT ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

}
func (cartItem *CartItem) BeforeCreate(tx *gorm.DB) error {
	cartItem.Id = 0
	// cartItem.Amount.AccountId = cartItem.AccountId
	return nil
}
func (cartItem *CartItem) AfterFind(tx *gorm.DB) (err error) {

	cartItem.PaymentSubjectYandex = cartItem.PaymentSubject.Code
	cartItem.Amount.Value = cartItem.Cost
	cartItem.Amount.Currency = "RUB"

	cartItem.WarehouseItems = GetAvailabilityProductWarehouseItems(cartItem.AccountId, cartItem.ProductId)

	return nil
}
func (cartItem *CartItem) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(cartItem)
	} else {
		_db = _db.Model(&CartItem{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Product","Product.MeasurementUnit","PaymentSubject","PaymentAmount","PaymentMode","Warehouse"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

// ############# Entity interface #############
func (cartItem CartItem) GetId() uint { return cartItem.Id }
func (cartItem *CartItem) setId(id uint) { cartItem.Id = id }
func (cartItem *CartItem) setPublicId(publicId uint) { }
func (cartItem CartItem) GetAccountId() uint { return cartItem.AccountId }
func (cartItem *CartItem) setAccountId(id uint) { cartItem.AccountId = id }
func (cartItem CartItem) SystemEntity() bool { return false }
// ############# End of Entity interface #############


// ######### CRUD Functions ############
func (cartItem CartItem) create() (Entity, error)  {

	_item := cartItem
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	if err := (Order{Id: _item.OrderId, AccountId: _item.AccountId}).UpdateDeliveryData(); err != nil {
		log.Println("Error update cart item: ", err)
	}
	if err := (Order{Id: _item.OrderId, AccountId: _item.AccountId}).UpdateCost(); err != nil {
		log.Println("Error update cart item: ", err)
	}

	var entity Entity = &_item

	return entity, nil
}
func (CartItem) get(id uint, preloads []string) (Entity, error) {

	var item CartItem

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (cartItem *CartItem) load(preloads []string) error {
	if cartItem.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := cartItem.GetPreloadDb(false, false, preloads).First(cartItem, cartItem.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (cartItem *CartItem) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности найти объект по public Id")
}
func (CartItem) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return CartItem{}.getPaginationList(accountId, 0,100,sortBy,"",nil,preload)
}
func (CartItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	orderChannels := make([]CartItem,0)
	var total int64

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&CartItem{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orderChannels, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&CartItem{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&CartItem{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&orderChannels).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&CartItem{}).GetPreloadDb(false, false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}
	
	entities := make([]Entity,len(orderChannels))
	for i := range orderChannels {
		entities[i] = &orderChannels[i]
	}

	return entities, total, nil
}
func (cartItem *CartItem) update(input map[string]interface{}, preloads []string) error {

	delete(input,"amount")
	delete(input,"payment_subject")
	delete(input,"payment_mode")
	delete(input,"product")
	delete(input,"order")
	delete(input,"warehouse_item")
	utils.FixInputHiddenVars(&input)
	/*if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}*/
	if err := utils.ConvertMapVarsToUINT(&input, []string{"quantity"}); err != nil {
		return err
	}

	// Делаем загрузку товара:
	if err := cartItem.load([]string{"Product"}); err != nil {
		return err
	}

	// Делаем перерасчет общей стоимости:
	if err := cartItem.GetPreloadDb(false, false, nil).Where("id = ?", cartItem.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	if err := (Order{Id: cartItem.OrderId, AccountId: cartItem.AccountId}).UpdateDeliveryData(); err != nil {
		log.Println("Error UpdateDeliveryData 1: ", err)
	}
	if err := (Order{Id: cartItem.OrderId, AccountId: cartItem.AccountId}).UpdateCost(); err != nil {
		log.Println("Error UpdateCost 2: ", err)
	}

	err := cartItem.GetPreloadDb(false,false, preloads).First(cartItem, cartItem.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (cartItem *CartItem) delete () error {
	if err := cartItem.cancelReserve(); err != nil {
		return err
	}
	if err := cartItem.wastedRollBack(); err != nil {
		return err
	}

	if err := cartItem.GetPreloadDb(true,false,nil).Where("id = ?", cartItem.Id).Delete(cartItem).Error; err != nil {
		return err
	}

	if err := (Order{Id: cartItem.OrderId, AccountId: cartItem.AccountId}).UpdateDeliveryData(); err != nil {
		log.Println("Error update cart item: ", err)
	}
	if err := (Order{Id: cartItem.OrderId, AccountId: cartItem.AccountId}).UpdateCost(); err != nil {
		log.Println("Error update order: ", err)
	}

	return nil

}
// ######### END CRUD Functions ############


// ########## Work function ############
type ReserveCartItem struct {
	WarehouseId *uint `json:"warehouse_id"` // Склад на котором содержится резерв или произведено списание.
	Quantity	*float64 `json:"quantity"`	// Объем резерва / списания
	Reserved	*bool `json:"reserved"`		// Есть резерв или нет (??)
	Wasted		*bool `json:"wasted"`		// Есть резерв или нет (??)
}

// func (cartItem *CartItem) UpdateReserve (warehouseId uint, quantity float64, reserved bool) error {
func (cartItem *CartItem) UpdateReserve (data ReserveCartItem) error {

	if cartItem.Id < 1 { return utils.Error{Message: "Техническая ошибка: не указан id of CartItem"}}
	
	// фикс. объем
	quantity := float64(0)
	if data.Quantity != nil {
		quantity = *data.Quantity
	}

	// 1. Если нужно изменить или установить warehouse_id
	if (data.WarehouseId != nil && cartItem.WarehouseId == nil) || (data.WarehouseId != nil && cartItem.WarehouseId != nil && *cartItem.WarehouseId != *data.WarehouseId){


		// Снимаем резерв, если он есть
		if cartItem.Reserved {
			if err := cartItem.cancelReserve(); err != nil { return err }
		}

		// Ставим новый резерв (в любом случае)
		if cartItem.WarehouseId != nil || data.WarehouseId != nil {
			whId := uint(0) // new id of warehouse
			if cartItem.WarehouseId != nil { whId = *cartItem.WarehouseId }
			if data.WarehouseId != nil { whId = *data.WarehouseId }

			if err := cartItem.setReserve(whId, quantity); err != nil {
				return err
			}
		}

	} else {  // Если не надо менять
		// Меняем локально статус, если нам известен новый статус и не меняется warehouseId
		if data.Reserved != nil || ((data.WarehouseId != nil && data.Reserved != nil && (cartItem.Reserved != *data.Reserved)) && cartItem.WarehouseId != nil) {
			// Ставим новый резерв
			if *data.Reserved {

				if cartItem.WarehouseId == nil {
					return utils.Error{Message: "Тех. ошибка: 27r3 - отсутствует склад id"}
				}
				if err := cartItem.setReserve(*cartItem.WarehouseId, quantity); err != nil {
					return err
				}
			} else {
				// снимаем резерв
				if err := cartItem.cancelReserve(); err != nil { return err }
			}
		}

		if (data.Wasted != nil && data.WarehouseId == nil) && (cartItem.Wasted != *data.Wasted) && cartItem.WarehouseId != nil {

			// Новое значение - возврат
			if *data.Wasted {
				if err := cartItem.wasted(); err != nil {
					return err
				}
			} else {
				// Делаем возврат товара на склад
				if err := cartItem.wastedRollBack(); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Резервирует товар на warehouseId в нужном объеме, НЕ снимая старый резерв
func (cartItem *CartItem) setReserve(warehouseId uint, quantity float64) error {

	if cartItem.Id < 1 || cartItem.ProductId < 1 { return utils.Error{Message: "Тех.ошибка cartItem.Id < 1"}}


	// Получаем product источник(и) / может быть is_kit
	var product Product
	if err := (Account{Id: cartItem.AccountId}).LoadEntity(&product, cartItem.ProductId, []string{"SourceItems"}); err != nil {return err}

	// 1. Идем по исходникам и находим wh_item на нужном складе с проверкой на доступный объем
	for _,v := range product.SourceItems {
		// fmt.Printf("Source item: id: %v || q: %v \n", v.SourceId, v.Quantity)

		// Объем для резерва
		var warehouseItem WarehouseItem
		
		// Находим нужный Item
		err := db.Model(&WarehouseItem{}).Where("product_id = ? AND warehouse_id = ? AND stock >= ?", v.SourceId, warehouseId, v.Quantity).
			First(&warehouseItem).Error
		if err != nil && err != gorm.ErrRecordNotFound {return err }
		if err == gorm.ErrRecordNotFound { return utils.Error{Message: "На складе отсутствует необходимо число товара"}	}

		// Резервируем нужный объем
		if quantity > 0 {
			err := db.Exec("UPDATE warehouse_items SET (stock,reservation) = (stock - ?, reservation + ?) WHERE id = ?",
				quantity*v.Quantity, quantity*v.Quantity, warehouseItem.Id).Error
			if err != nil { return err }
		}
	}
	
	// 3. Переводим cartItem в статус "Зарезервировано" и указываем warehouse_id
	err := db.Exec("UPDATE cart_items SET reserved = true, warehouse_id = ? WHERE id = ?", warehouseId, cartItem.Id).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	return nil
}

// Снимает статус резерва и снимает объем из резерва со склада
func (cartItem *CartItem) cancelReserve() error {

	// Проверяем текущее состояние
	if cartItem.Id < 1 || !cartItem.Reserved { return nil }

	// Получаем product источник(и) / может быть is_kit
	var product Product
	if err := (Account{Id: cartItem.AccountId}).LoadEntity(&product, cartItem.ProductId, []string{"SourceItems"}); err != nil {return err}

	// 1. Идем по исходникам и находим wh_item на нужном складе с проверкой на доступный объем
	for _,v := range product.SourceItems {
		
		// Снимаем резерв для каждого элемента
		if cartItem.Quantity > 0 && cartItem.WarehouseId != nil {

			// Объем для резерва
			var warehouseItem WarehouseItem

			// Находим нужный Item
			err := db.Model(&WarehouseItem{}).Where("product_id = ? AND warehouse_id = ?", v.SourceId, *cartItem.WarehouseId).
				First(&warehouseItem).Error
			if err != nil && err != gorm.ErrRecordNotFound {return err }
			if err == gorm.ErrRecordNotFound { return utils.Error{Message: "На складе отсутствует необходимо число товара"}	}


			err = db.Exec("UPDATE warehouse_items SET (stock,reservation) = (stock + ?, reservation - ?) WHERE id = ?",
				cartItem.Quantity*v.Quantity, cartItem.Quantity*v.Quantity, warehouseItem.Id).Error
			if err != nil { return err }
		}
	}

	if cartItem.WarehouseId == nil && cartItem.Quantity > 0 {
		return utils.Error{Message: "Тех.ошибка снятия резерва с неуказанным складом."}
	}

	// 2. Снимаем статус 'Reserved' c CartItem в любом случае (т.к. Reserved = true )
	err := db.Exec("UPDATE cart_items SET reserved = false WHERE id = ?", cartItem.Id).Error
	if err != nil && err != gorm.ErrRecordNotFound { return err }

	return nil
}

// Списывает товар со склада, либо из резерва, либо из общего числа
func (cartItem *CartItem) wasted() error {

	if cartItem.Id < 1 || cartItem.ProductId < 1 { return utils.Error{Message: "Тех.ошибка cartItem.Id < 1"}}

	// Загружаем еще раз т.к. остатки могли обновиться
	if err := cartItem.load([]string{""}); err != nil {
		return err
	}

	if cartItem.WarehouseId == nil { return utils.Error{Message: "Не указан склад списания: warehouse id"}}

	var product Product
	if err := (Account{Id: cartItem.AccountId}).LoadEntity(&product, cartItem.ProductId, []string{"SourceItems"}); err != nil {return err}

	// 1. Идем по исходникам и находим wh_item на нужном складе с проверкой на доступный объем
	for _,v := range product.SourceItems {
		// fmt.Printf("Source item: id: %v || q: %v \n", v.SourceId, v.Quantity)

		// Объем для резерва
		var warehouseItem WarehouseItem

		if cartItem.Quantity > 0 {
			if cartItem.Reserved {

				err := db.Model(&WarehouseItem{}).Where("product_id = ? AND warehouse_id = ? AND reservation >= ?", v.SourceId, *cartItem.WarehouseId, cartItem.Quantity*v.Quantity).
					First(&warehouseItem).Error
				if err != nil {
					log.Println("err :: ", err)
					return utils.Error{Message: "На складе в резерве отсутствует указанный товар в нужном объеме"}
				}

				// 1. Списываем с Warehouse_item из резерва
				err = db.Exec("UPDATE warehouse_items SET reservation = reservation - ? WHERE id = ?",
					cartItem.Quantity*v.Quantity, warehouseItem.Id).Error
				if err != nil { return err }

			} else {

				err := db.Model(&WarehouseItem{}).Where("product_id = ? AND warehouse_id = ? AND stock >= ?", v.SourceId, *cartItem.WarehouseId, cartItem.Quantity*v.Quantity).
					First(&warehouseItem).Error
				if err != nil {
					log.Println("err :: ", err)
					return utils.Error{Message: "На складе отсутствует указанный товар"}
				}

				// Списываем из stock нужное число товара
				err = db.Exec("UPDATE warehouse_items SET stock = stock - ? WHERE id = ?",
					cartItem.Quantity*v.Quantity, warehouseItem.Id).Error
				if err != nil { return err }
			}
		}
	}

	// 3. Переводим cartItem в статус "Не Зарезервировано" && "Потрачено"
	err := db.Exec("UPDATE cart_items SET reserved = false,wasted = true WHERE id = ?", cartItem.Id).Error
	if err != nil && err != gorm.ErrRecordNotFound {return err}

	return nil
}

// Возврат товара на склад без резерва
func (cartItem *CartItem) wastedRollBack() error {

	if cartItem.Id < 1 || cartItem.ProductId < 1 { return utils.Error{Message: "Тех.ошибка cartItem.Id < 1"}}

	if cartItem.WarehouseId == nil { return  nil }
	// if cartItem.WarehouseId == nil { return utils.Error{Message: "Не указан склад списания: warehouse id"}}

	if cartItem.Quantity == 0 || !cartItem.Wasted {return nil }

	var product Product
	if err := (Account{Id: cartItem.AccountId}).LoadEntity(&product, cartItem.ProductId, []string{"SourceItems"}); err != nil {return err}

	// 1. Идем по исходникам и находим wh_item на нужном складе с проверкой на доступный объем
	for _,v := range product.SourceItems {
		// fmt.Printf("Source item: id: %v || q: %v \n", v.SourceId, v.Quantity)

		// Объем для резерва
		var warehouseItem WarehouseItem

		// Находим нужный Item
		err := db.Model(&WarehouseItem{}).Where("product_id = ? AND warehouse_id = ?", v.SourceId, *cartItem.WarehouseId).
			First(&warehouseItem).Error
		if err != nil {
			return utils.Error{Message: "На складе отсутствует указанный товар"}
		}

		// 1. Делаем возврат товара
		// Добавляем объем в stock
		err = db.Exec("UPDATE warehouse_items SET stock = stock + ? WHERE id = ?",
			cartItem.Quantity*v.Quantity, warehouseItem.Id).Error
		if err != nil { return err }
	}

	// 3. Переводим cartItem в статус "НЕ Потрачено"
	err := db.Exec("UPDATE cart_items SET wasted = false WHERE id = ?", cartItem.Id).Error
	if err != nil && err != gorm.ErrRecordNotFound {return err}

	return nil
}

// Списывает позицию со склада, если она была в резерве
func (cartItem *CartItem) SetWastedOffFromWarehouse() error {

	if cartItem.Id < 1 || cartItem.Wasted {
		return nil
	}

	// Списываем, если объем > 0 и он был либо в резерве либо из общего хранилища
	if cartItem.Quantity > 0 && cartItem.WarehouseId != nil {

		reserve := float64(0)
		stock := float64(0)

		var product Product
		if err := (Account{Id: cartItem.AccountId}).LoadEntity(&product, cartItem.ProductId, []string{"SourceItems"}); err != nil {return err}

		// 1. Идем по исходникам и находим wh_item на нужном складе с проверкой на доступный объем
		for _,v := range product.SourceItems {
			// fmt.Printf("Source item: id: %v || q: %v \n", v.SourceId, v.Quantity)

			// Объем для резерва
			var warehouseItem WarehouseItem

			// Находим нужный Item
			err := db.Model(&WarehouseItem{}).Where("product_id = ? AND warehouse_id = ?", v.SourceId, *cartItem.WarehouseId).
				First(&warehouseItem).Error
			if err != nil {
				return utils.Error{Message: "На складе отсутствует указанный товар"}
			}

			if cartItem.Reserved {
				reserve = cartItem.Quantity*v.Quantity
			} else {
				stock = cartItem.Quantity*v.Quantity
			}

			err = db.Exec("UPDATE warehouse_items SET (stock,reservation) = (stock - ?, reservation - ?) WHERE id = ?",
				stock, reserve, warehouseItem.Id).Error
			if err != nil { return err }
		}

	}

	// Переводим в статус Списано и снимаем статус зарезервировано
	err := db.Exec("UPDATE cart_items SET reserved = false, wasted = true WHERE id = ?", cartItem.Id).Error
	if err != nil && err != gorm.ErrRecordNotFound {return err}

	return nil
}

// Возвращает склады, где есть указанный товар в нужном обхеме или все компоненты (на одном складе)
// func (cartItem CartItem) GetAvailabilityWarehouseItems() []WarehouseItem {
func GetAvailabilityProductWarehouseItems(accountId, productId uint) []WarehouseItem {


	warehouseItems := make([]WarehouseItem,0)

	if accountId < 1 || productId < 1 { return warehouseItems}

	if productId < 1 { return warehouseItems}

	/*var product Product
	if err := db.Table("products").Preload("SourceItems").First(&product,productId).Error; err != nil {
		return warehouseItems
	}*/
	sourceItems := make([]ProductSource,0)
	if err := db.Model(&ProductSource{}).Where("product_id = ?", productId).Find(&sourceItems).Error; err != nil {
		return warehouseItems
	}
	/*if err := (Account { Id: accountId }).LoadEntity(&product, productId, []string{"SourceItems"}); err != nil {
		return warehouseItems
	}*/

	// Это доставка, она доступна (по дефолту)
	if productId < 1 { return warehouseItems }

	// Список item для резерва
	whItems := make([]WarehouseItem,0)

	for itemId := range sourceItems {

		// товар ли это > 0 ?
		if sourceItems[itemId].ProductId > 0 {

			// ProductId - id товара ДЛЯ которого собираем
			// SourceId - id товаров ИЗ которых собираем

			// Список позиций на складах, где хранится один из source' товара
			wIs := make([]WarehouseItem,0)

			// min cписок складов для исходника товара
			warehouses := make([]uint,0)

			// Загружаем список позиций на складах подходящих по объему
			err := db.Model(&WarehouseItem{}).
				Where("product_id = ? AND stock >= ? AND stock > 0", sourceItems[itemId].SourceId, sourceItems[itemId].Quantity).
				Preload("Warehouse").Find(&wIs).Error

			// Если ничего не найдено или список 0 => возвращаем  []
			if err != nil || len(wIs) == 0 { return warehouseItems }

			// Если это первый проход, то без проверок
			// Если нашли позицию(ии) добавляем id склада
			if itemId == 0 {
				for y := range wIs {
					warehouses = append(warehouses, wIs[y].WarehouseId)
					whItems = append(whItems,wIs[y]) // << добавляем warehouseItem для товара
					continue
				}
			} else {

				// если это не первый товар в исходниках

				// arr := make([]WarehouseItem,0)
				arr := whItems
				// идем по найденным позициям на складах и проверяем есть ли эти склады у нового товара в исходниках
				for _, vh := range arr {
					needWarehouseId := vh.WarehouseId

					existOfOne := false // если есть = true

					// Идем по найденным whItems
					for y, v := range wIs {

						if v.WarehouseId == needWarehouseId {
							existOfOne = true
							whItems = append(whItems,wIs[y]) // << добавляем warehouseItem для товара
						}
					}

					// Если не найден склад, то исключаем все позиции с wh = needWarehouseId
					if !existOfOne {
						whItems = RemoveWarehouseFromItems(whItems, needWarehouseId)
					}


				}
			}

		}
	}

	warehouses := make([]uint,0)
	for _, v := range whItems {
		// Добавляем ID склада
		if _, ok := FindUINT(warehouses,v.WarehouseId); !ok {
			warehouses = append(warehouses, v.WarehouseId)
		}
	}

	// Просто возвращаем ID складов, где найдено необходимо число товаров
	for _, v := range warehouses {
		warehouseItems = append(warehouseItems, WarehouseItem{WarehouseId: v})
	}

	return warehouseItems
}

func FindUINT(slice []uint, val uint) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func RemoveWarehouseFromItems(warehouseItems []WarehouseItem, warehouseId uint) []WarehouseItem {
	arr := make([]WarehouseItem,0)

	for index,v := range warehouseItems {
		// fmt.Println("in Id: ", warehouseItems[index].WarehouseId)
		// fmt.Println("wh Id: ", warehouseId)
		if v.WarehouseId != warehouseId {
			// warehouseItems = append(warehouseItems[:index], warehouseItems[index+1:]...)
			arr = append(arr, warehouseItems[index])
		}
	}
	return arr
}