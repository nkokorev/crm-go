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
type ShipmentProduct struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	ShipmentId 	uint	`json:"shipment_id" gorm:"type:int;index;"`
	ProductId 	uint	`json:"product_id" gorm:"type:int;index;"`

	// Сколько ед. в одном товаре ()
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

func (ShipmentProduct) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ShipmentProduct{}); err != nil {
		log.Fatal(err)
	}
	err := db.Exec("ALTER TABLE shipment_products " +
		"ADD CONSTRAINT shipment_products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT shipment_products_shipment_id_fkey FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT shipment_products_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (shipmentProduct *ShipmentProduct) BeforeCreate(tx *gorm.DB) error {
	shipmentProduct.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&ShipmentProduct{}).Where("account_id = ? AND shipment_id = ?",  shipmentProduct.AccountId, shipmentProduct.ShipmentId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	shipmentProduct.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (shipmentProduct *ShipmentProduct) AfterFind(tx *gorm.DB) (err error) {

	return nil
}
func (shipmentProduct *ShipmentProduct) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(shipmentProduct)
	} else {
		_db = _db.Model(&ShipmentProduct{})
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
func (shipmentProduct ShipmentProduct) GetId() uint { return shipmentProduct.Id }
func (shipmentProduct *ShipmentProduct) setId(id uint) { shipmentProduct.Id = id }
func (shipmentProduct *ShipmentProduct) setPublicId(publicId uint) { shipmentProduct.PublicId = publicId }
func (shipmentProduct ShipmentProduct) GetAccountId() uint { return shipmentProduct.AccountId }
func (shipmentProduct *ShipmentProduct) setAccountId(id uint) { shipmentProduct.AccountId = id }
func (ShipmentProduct) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (shipmentProduct ShipmentProduct) create() (Entity, error)  {

	en := shipmentProduct

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
func (ShipmentProduct) get(id uint, preloads []string) (Entity, error) {

	var shipmentProduct ShipmentProduct

	err := shipmentProduct.GetPreloadDb(false,false,preloads).First(&shipmentProduct, id).Error
	if err != nil {
		return nil, err
	}
	return &shipmentProduct, nil
}
func (shipmentProduct *ShipmentProduct) load(preloads []string) error {

	err := shipmentProduct.GetPreloadDb(false,false,preloads).First(shipmentProduct, shipmentProduct.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (shipmentProduct *ShipmentProduct) loadByPublicId(preloads []string) error {
	
	if shipmentProduct.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить ShipmentProduct - не указан  Id"}
	}
	if err := shipmentProduct.GetPreloadDb(false,false, preloads).First(shipmentProduct, "account_id = ? AND public_id = ?", shipmentProduct.AccountId, shipmentProduct.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (ShipmentProduct) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return ShipmentProduct{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (ShipmentProduct) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	shipmentProducts := make([]ShipmentProduct,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&ShipmentProduct{}).GetPreloadDb(false,false,preloads).
			Joins("LEFT JOIN products ON products.id = shipment_products.product_id").
			Select("products.*, shipment_products.*").
			Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&shipmentProducts, "shipment_id ILIKE ?", search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ShipmentProduct{}).
			Where("shipment_id ILIKE ? ", search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&ShipmentProduct{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&shipmentProducts).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ShipmentProduct{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(shipmentProducts))
	for i := range shipmentProducts {
		entities[i] = &shipmentProducts[i]
	}

	return entities, total, nil
}
func (shipmentProduct *ShipmentProduct) update(input map[string]interface{}, preloads []string) error {

	delete(input,"product")
	delete(input,"shipment")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","shipment_id","product_id"}); err != nil {
		return err
	}
	// input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := shipmentProduct.GetPreloadDb(false,false,nil).Where(" id = ?", shipmentProduct.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := shipmentProduct.GetPreloadDb(false,false,preloads).First(shipmentProduct, shipmentProduct.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (shipmentProduct *ShipmentProduct) delete () error {
	return shipmentProduct.GetPreloadDb(true,false,nil).Where("id = ?", shipmentProduct.Id).Delete(shipmentProduct).Error
}
// ######### END CRUD Functions ############

// Перевод состояние позиции как загруженной на склад
func (shipmentProduct *ShipmentProduct) SetWarehousePosted () error {
	return shipmentProduct.update(map[string]interface{}{"warehouse_posted":true},nil)
}
