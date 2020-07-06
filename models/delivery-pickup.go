package models

import (
	"github.com/jinzhu/gorm"
)

type DeliveryPickup struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint	`json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа
	ShopID		uint 	`json:"shopId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится

	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Курьерская доставка", "Почта России", "Самовывоз"
	Price 		float64 `json:"price" gorm:"type:numeric;default:0"` // стоимость доставки
	ApiKey 		float64 `json:"apiKey" gorm:"type:varchar(255);"` // стоимость доставки

	Shop Shop
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

func (DeliveryPickup) getPaginationList(accountId uint, offset, limit int, order string, search *string) ([]Entity, error) {

	delivers := make([]DeliveryPickup,0)

	err := db.Model(&DeliveryPickup{}).Limit(limit).Offset(offset).Order(order).Find(&delivers, "account_id = ?", accountId).Error
	if err != nil {
		return nil, err
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(delivers))
	for i, v := range delivers {
		entities[i] = &v
	}

	return entities, nil
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


