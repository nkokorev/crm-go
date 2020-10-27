package models

import (
	"errors"
	"fmt"
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

	// null - ни в резерве, ни в списан. 
	WarehouseItemId		*uint `json:"warehouse_item_id" gorm:"type:int;"`
	WarehouseItem		*WarehouseItem	`json:"warehouse_item"`
	WarehouseItems		[]WarehouseItem	`json:"warehouse_items" gorm:"-"` // AfterFind

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
	cartItem.WarehouseItems = cartItem.GetAvailabilityWarehouseItems()

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

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Product","PaymentSubject","PaymentAmount","PaymentMode","WarehouseItem"})

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
		log.Println("Error update order: ", err)
	}
	if err := (Order{Id: cartItem.OrderId, AccountId: cartItem.AccountId}).UpdateCost(); err != nil {
		log.Println("Error update order: ", err)
	}

	err := cartItem.GetPreloadDb(false,false, preloads).First(cartItem, cartItem.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (cartItem *CartItem) delete () error {
	if err := cartItem.GetPreloadDb(true,false,nil).Where("id = ?", cartItem.Id).Delete(cartItem).Error; err != nil {
		fmt.Println("Доставка удалена")
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
}

// func (cartItem *CartItem) UpdateReserve (warehouseId uint, quantity float64, reserved bool) error {
func (cartItem *CartItem) UpdateReserve (data ReserveCartItem) error {

	if cartItem.Id < 1 { return utils.Error{Message: "Техническая ошибка: не указан id of CartItem"}}

	// Если резерва нет
	if cartItem.WarehouseItemId == nil && data.Reserved != nil{
		// check reserve
	}

	// Если нужно изменить wh_id
	if data.WarehouseId != nil {

		// Проверяем preload in CartItemCr
		if cartItem.WarehouseItem == nil {
			return utils.Error{Message: "Тех. ошибка: необходимо загрузить cartItem.WarehouseItem"}
		}

		// Проверяем, нужно ли реально менять склад
		if cartItem.WarehouseItem.WarehouseId != *data.WarehouseId {
			// todo: change reserve

			// 1. Проверяем был ли резерв и снимаем его, если да
			if cartItem.Reserved {

				// Если есть реальный warehouse_item_id - переводим резерв запас на складе
				if cartItem.WarehouseItemId != nil {

					err := db.Exec("UPDATE warehouse_items SET (stock,reservation) = (stock + ?, reservation - ?) WHERE id = ?",
						cartItem.Quantity,cartItem.Quantity, *cartItem.WarehouseItemId).Error
					if err != nil && err != gorm.ErrRecordNotFound { return err }
				}

				// Снимаем статус 'Reserved'
				err := db.Exec("UPDATE cart_items SET reserved = false WHERE id = ?", cartItem.Id).Error
				if err != nil && err != gorm.ErrRecordNotFound { return err }
				
			}

			// 2. Проверить есть ли место на новом складе

			// Проверяем, если необходимо число на новом складе и если есть - резервируем нужный объем
			quantity := float64(0)
			if data.Quantity != nil {
				quantity = *data.Quantity
			}
		

			var warehouseItem WarehouseItem
			err := db.Model(&WarehouseItem{}).Where("product_id = ? AND warehouse_id = ? AND stock >= ?", cartItem.ProductId, *data.WarehouseId, quantity).
				First(&warehouseItem).Error
			if err != nil && err != gorm.ErrRecordNotFound {return err }
			if err == gorm.ErrRecordNotFound { return utils.Error{Message: "На складе отсутствует необходимо число товара"}	}

			if data.Quantity != nil {
				fmt.Println("Обновляем число: ", *data.Quantity)
				err = db.Exec("UPDATE warehouse_items SET (stock,reservation) = (stock - ?, reservation + ?) " +
					"WHERE id = ?",
					*data.Quantity, *data.Quantity, warehouseItem.Id).Error
				if err != nil { return err }
			}


			// Переводи в статус "Зарезервировано" и указываем warehouse_id
			err = db.Exec("UPDATE cart_items SET reserved = true, warehouse_item_id = ? WHERE id = ?", warehouseItem.Id, cartItem.Id).Error
			if err != nil && err != gorm.ErrRecordNotFound {return err}

		}
	}

	return nil

}

// Возвращает склады, где есть указанный товар в нужном обхеме
func (cartItem CartItem) GetAvailabilityWarehouseItems() []WarehouseItem {

	warehouseItems := make([]WarehouseItem,0)

	if cartItem.Id < 1 { return warehouseItems}

	// Это доставка, она доступна (по дефолту)
	if cartItem.ProductId < 1 { return warehouseItems }

	err := db.Model(&WarehouseItem{}).
		Where("product_id = ? AND stock >= ?", cartItem.ProductId, cartItem.Quantity).
		Preload("Warehouse").Find(&warehouseItems).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return warehouseItems
	}

	return warehouseItems

}
