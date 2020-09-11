package models

import (
	"database/sql"
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

	// Поставщик (может быть не известен)
	CompanyId 	*uint	`json:"company_id" gorm:"type:int;index;"`

	// Склад назначения
	WarehouseId *uint	`json:"warehouse_id" gorm:"type:int;index;"`

	// Сумма поставки - высчитывается
	Amount	 	float64 `json:"_amount" gorm:"-"`

	// Единиц товаров - высчитывается
	ProductUnits	uint `json:"_product_units" gorm:"-"`

	// Завершена ли поставка. Если true - нельзя добавить или убрать из нее товар.
	Completed	bool	`json:"completed" gorm:"type:bool;default:false"`

	// Фактический список товаров в поставке + объем + закупочная цены
	ShipmentProduct 	[]ShipmentProduct 	`json:"shipment_product"`

	// Список товаров в поставке
	Products 	[]Product 	`json:"products" gorm:"many2many:shipment_products"`

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

	err = db.SetupJoinTable(&Shipment{}, "Products", &ShipmentProduct{})
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

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Product","Warehouse"})

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
			Where(filter).Find(&shipments, "label ILIKE ?", search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Shipment{}).
			Where("account_id = ? AND label ILIKE ?", accountId, search).
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

	delete(input,"image")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","priority","web_site_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

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
func (shipment *Shipment) CompleteDelivery() error {
	return nil
}