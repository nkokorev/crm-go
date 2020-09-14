package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Данные поставки
type ShipmentItem struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	ShipmentId 	uint	`json:"shipment_id" gorm:"type:int;index;"`
	ProductId 	uint	`json:"product_id" gorm:"type:int;index;"`

	// Сколько ед.товара планируется / пришло по факту
	VolumeOrder	float64 `json:"volume_order" gorm:"type:numeric;"`
	VolumeFact	float64 `json:"volume_fact" gorm:"type:numeric;"`

	// Оприходование позиции на склад в размере VolumeFact 
	WarehousePosted	bool	`json:"warehouse_posted" gorm:"type:bool;default:false"`

	// Закупочная цена
	PaymentAmount	float64 `json:"payment_amount" gorm:"type:numeric;"`

	Product 	Product 	`json:"product"`
	Shipment 	Shipment 	`json:"shipment"`

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

func (ShipmentItem) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ShipmentItem{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE shipment_items " +
		"ADD CONSTRAINT shipment_items_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT shipment_items_shipment_id_fkey FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT shipment_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (shipmentItem *ShipmentItem) BeforeCreate(tx *gorm.DB) error {
	shipmentItem.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&ShipmentItem{}).Where("account_id = ? AND shipment_id = ?",  shipmentItem.AccountId, shipmentItem.ShipmentId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	shipmentItem.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (shipmentItem *ShipmentItem) AfterFind(tx *gorm.DB) (err error) {

	return nil
}
func (shipmentItem *ShipmentItem) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(shipmentItem)
	} else {
		_db = _db.Model(&ShipmentItem{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Product","Shipment","PaymentAmount"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}
}

// ############# Entity interface #############
func (shipmentItem ShipmentItem) GetId() uint { return shipmentItem.Id }
func (shipmentItem *ShipmentItem) setId(id uint) { shipmentItem.Id = id }
func (shipmentItem *ShipmentItem) setPublicId(publicId uint) { shipmentItem.PublicId = publicId }
func (shipmentItem ShipmentItem) GetAccountId() uint { return shipmentItem.AccountId }
func (shipmentItem *ShipmentItem) setAccountId(id uint) { shipmentItem.AccountId = id }
func (ShipmentItem) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (shipmentItem ShipmentItem) create() (Entity, error)  {

	en := shipmentItem

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
func (ShipmentItem) get(id uint, preloads []string) (Entity, error) {

	var shipmentItem ShipmentItem

	err := shipmentItem.GetPreloadDb(false,false,preloads).First(&shipmentItem, id).Error
	if err != nil {
		return nil, err
	}
	return &shipmentItem, nil
}
func (shipmentItem *ShipmentItem) load(preloads []string) error {

	err := shipmentItem.GetPreloadDb(false,false,preloads).First(shipmentItem, shipmentItem.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (shipmentItem *ShipmentItem) loadByPublicId(preloads []string) error {
	
	if shipmentItem.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить ShipmentItem - не указан  Id"}
	}
	if err := shipmentItem.GetPreloadDb(false,false, preloads).First(shipmentItem, "account_id = ? AND public_id = ?", shipmentItem.AccountId, shipmentItem.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (ShipmentItem) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return ShipmentItem{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (ShipmentItem) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	shipmentItems := make([]ShipmentItem,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&ShipmentItem{}).GetPreloadDb(false,false,preloads).
			Joins("LEFT JOIN products ON products.id = shipment_items.product_id").
			Select("products.*, shipment_items.*").
			Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&shipmentItems, "shipment_id ILIKE ?", search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ShipmentItem{}).
			Where("shipment_id ILIKE ? ", search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&ShipmentItem{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&shipmentItems).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ShipmentItem{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(shipmentItems))
	for i := range shipmentItems {
		entities[i] = &shipmentItems[i]
	}

	return entities, total, nil
}
func (shipmentItem *ShipmentItem) update(input map[string]interface{}, preloads []string) error {

	delete(input,"product")
	delete(input,"shipment")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","shipment_id","product_id"}); err != nil {
		return err
	}
	// input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := shipmentItem.GetPreloadDb(false,false,nil).Where(" id = ?", shipmentItem.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := shipmentItem.GetPreloadDb(false,false,preloads).First(shipmentItem, shipmentItem.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (shipmentItem *ShipmentItem) delete () error {
	return shipmentItem.GetPreloadDb(true,false,nil).Where("id = ?", shipmentItem.Id).Delete(shipmentItem).Error
}
// ######### END CRUD Functions ############

// Перевод состояние позиции как загруженной на склад
func (shipmentItem *ShipmentItem) SetWarehousePosted () error {
	return shipmentItem.update(map[string]interface{}{"warehouse_posted":true},nil)
}
