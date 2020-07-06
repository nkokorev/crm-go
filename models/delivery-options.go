package models

import (
	"github.com/jinzhu/gorm"
	"log"
)

type DeliveryOption struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint	`json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа
	ShopID		uint 	`json:"shopId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится

	// ownerID   	uint	//`json:"-"`   // ?? gorm:"association_foreignkey:ID"
	// ownerType	string	//`json:"ownerType" gorm:"type:varchar(80);column:owner_type"`

	OwnerID   	uint	//`json:"-"`   // ?? gorm:"association_foreignkey:ID"
	OwnerType	string	//`json:"ownerType" gorm:"type:varchar(80);column:owner_type"`

	DeliveryMethod Delivery `gorm:"-"`

	// DeliveryTypeId	uint `json:"deliveryTypeId" gorm:"type:int;index;default:NULL;"` // тип доставки
	// DeliveryTypeId	uint `json:"deliveryTypeId" gorm:"type:int;index;default:NULL;"` // тип доставки

	Name string `json:"name" gorm:"type:varchar(255);"` // "Доставка курьером", "Доставка Почтой России", "Самовывоз"
	Enabled bool `json:"enabled" gorm:"type:bool;default:true"` // активен ли вариант доставки для магазина
}

func (DeliveryOption) PgSqlCreate() {
	db.CreateTable(&DeliveryOption{})
	
	db.Model(&DeliveryOption{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&DeliveryOption{}).AddForeignKey("shop_id", "shops(id)", "CASCADE", "CASCADE")
}

// ############# Entity interface #############
func (deliveryOption DeliveryOption) getId() uint { return deliveryOption.ID }
func (deliveryOption *DeliveryOption) setId(id uint) { deliveryOption.ID = id }
func (deliveryOption DeliveryOption) GetAccountId() uint { return deliveryOption.AccountID }
func (deliveryOption *DeliveryOption) setAccountId(id uint) { deliveryOption.AccountID = id }
// ############# Entity interface #############

// ###### GORM Functional #######
func (DeliveryOption) TableName() string { return "delivery_options" }
func (deliveryOption *DeliveryOption) BeforeCreate(scope *gorm.Scope) error {
	deliveryOption.ID = 0
	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (deliveryOption DeliveryOption) create() (Entity, error)  {
	var newItem Entity = &deliveryOption
	
	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	if deliveryOption.DeliveryMethod != nil {
		err := db.Model(&DeliveryRussianPost{ID: 1}).Association("DeliveryOption").Append(&newItem).Error
		if err != nil {
			log.Fatal(err)
		}

		/*if err := deliveryOption.AppendAssociationMethod(deliveryOption.DeliveryMethod); err != nil {
			return nil, err
		}*/
	}

	return newItem, nil
}

func (DeliveryOption) get(id uint) (Entity, error) {

	var deliveryOption DeliveryOption

	err := db.First(&deliveryOption, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryOption, nil
}

func (deliveryOption *DeliveryOption) load() error {

	err := db.First(deliveryOption).Error
	if err != nil {
		return err
	}
	return nil
}

func (DeliveryOption) getPaginationList(accountId uint, offset, limit int, order string, search *string) ([]Entity, error) {

	delivers := make([]DeliveryOption,0)

	err := db.Model(&DeliveryOption{}).Limit(limit).Offset(offset).Order(order).Find(&delivers, "account_id = ?", accountId).Error
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

func (deliveryOption *DeliveryOption) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(deliveryOption).Omit("id", "account_id").Update(input).Error
}

func (deliveryOption DeliveryOption) delete () error {
	return db.Model(DeliveryOption{}).Where("id = ?", deliveryOption.ID).Delete(deliveryOption).Error
}

// ########## End of CRUD Entity interface ###########

