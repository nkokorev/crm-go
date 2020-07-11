package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type DeliveryPickup struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint	`json:"-" gorm:"index,not null"` // аккаунт-владелец ключа
	ShopID		uint 	`json:"shopId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится
	Code 		string	`json:"code" gorm:"type:varchar(16);default:'pickup';"` // Для идентификации во фронтенде

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // активен ли способ доставки
	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Самовывоз со основного склада"
	Price 		float64 `json:"price" gorm:"type:numeric;default:0"` // стоимость доставки при самовывозе (может быть и не 0, если через сервис)

	AddressRequired	bool	`json:"addressRequired" gorm:"type:bool;default:false"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postalCodeRequired" gorm:"type:bool;default:false"` // Требуется ли индекс в адресе доставки

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (DeliveryPickup) PgSqlCreate() {
	db.CreateTable(&DeliveryPickup{})
	
	db.Model(&DeliveryPickup{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}

// ############# Entity interface #############
func (deliveryPickup DeliveryPickup) getId() uint { return deliveryPickup.ID }
func (deliveryPickup *DeliveryPickup) setId(id uint) { deliveryPickup.ID = id }
func (deliveryPickup DeliveryPickup) GetAccountId() uint { return deliveryPickup.AccountID }
func (deliveryPickup *DeliveryPickup) setAccountId(id uint) { deliveryPickup.AccountID = id }
func (deliveryPickup *DeliveryPickup) setShopId(shopId uint) { deliveryPickup.ShopID = shopId }
func (DeliveryPickup) systemEntity() bool { return false }

func (deliveryPickup DeliveryPickup) GetCode() string {
	return deliveryPickup.Code
}
// ############# Entity interface #############


// ###### GORM Functional #######
// func (DeliveryPickup) TableName() string { return "delivery_pickups" }
func (deliveryPickup *DeliveryPickup) BeforeCreate(scope *gorm.Scope) error {
	deliveryPickup.ID = 0
	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (deliveryPickup DeliveryPickup) create() (Entity, error)  {
	var newItem Entity = &deliveryPickup
	
	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (DeliveryPickup) get(id uint) (Entity, error) {

	var deliveryPickup DeliveryPickup

	err := db.First(&deliveryPickup, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryPickup, nil
}

func (deliveryPickup *DeliveryPickup) load() error {

	err := db.First(deliveryPickup).Error
	if err != nil {
		return err
	}
	return nil
}

func (DeliveryPickup) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	delivers := make([]DeliveryPickup,0)
	var total uint

	err := db.Model(&DeliveryPickup{}).Limit(limit).Offset(offset).Order(sortBy).Find(&delivers, "account_id = ?", accountId).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Model(&DeliveryPickup{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(delivers))
	for i, v := range delivers {
		entities[i] = &v
	}

	return entities, total, nil
}

func (deliveryPickup *DeliveryPickup) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(deliveryPickup).Omit("id", "account_id").Update(input).Error
}

func (deliveryPickup DeliveryPickup) delete () error {
	return db.Model(DeliveryPickup{}).Where("id = ?", deliveryPickup.ID).Delete(deliveryPickup).Error
}

// ########## End of CRUD Entity interface ###########

func (deliveryPickup DeliveryPickup) GetName () string {
	return "Почта России"
}

func (deliveryPickup DeliveryPickup) CalculateDelivery(deliveryData DeliveryData) (*DeliveryData, error) {
	deliveryData.TotalCost = 0
	return &deliveryData, nil
}

func (deliveryPickup DeliveryPickup) checkMaxWeight(deliveryData DeliveryData) error {
	return nil
}

