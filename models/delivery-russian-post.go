package models

import (
	"github.com/jinzhu/gorm"
)

type DeliveryRussianPost struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint	`json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа
	ShopID		uint 	`json:"shopId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится

	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Курьерская доставка", "Почта России", "Самовывоз"
	Price 		float64 `json:"price" gorm:"type:numeric;default:0"` // стоимость доставки
	ApiKey 		float64 `json:"apiKey" gorm:"type:varchar(255);"` // стоимость доставки

	Shop Shop
	// Shop		uint 	`gorm:"polymorphic:Owner;"` // магазин, к которому относится
}

func (DeliveryRussianPost) PgSqlCreate() {
	db.CreateTable(&DeliveryRussianPost{})
	
	db.Model(&DeliveryRussianPost{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}

// ############# Entity interface #############
func (deliveryRussianPost DeliveryRussianPost) getId() uint { return deliveryRussianPost.ID }
func (deliveryRussianPost *DeliveryRussianPost) setId(id uint) { deliveryRussianPost.ID = id }
func (deliveryRussianPost DeliveryRussianPost) GetAccountId() uint { return deliveryRussianPost.AccountID }
func (deliveryRussianPost *DeliveryRussianPost) setAccountId(id uint) { deliveryRussianPost.AccountID = id }
// ############# Entity interface #############


// ###### GORM Functional #######
func (DeliveryRussianPost) TableName() string { return "delivery_russian_post" }
func (deliveryRussianPost *DeliveryRussianPost) BeforeCreate(scope *gorm.Scope) error {
	deliveryRussianPost.ID = 0
	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (deliveryRussianPost DeliveryRussianPost) create() (Entity, error)  {
	var newItem Entity = &deliveryRussianPost
	
	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (DeliveryRussianPost) get(id uint) (Entity, error) {

	var deliveryRussianPost DeliveryRussianPost

	err := db.First(&deliveryRussianPost, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryRussianPost, nil
}

func (deliveryRussianPost *DeliveryRussianPost) load() error {

	err := db.First(deliveryRussianPost).Error
	if err != nil {
		return err
	}
	return nil
}

func (DeliveryRussianPost) getPaginationList(accountId uint, offset, limit int, order string, search *string) ([]Entity, error) {

	delivers := make([]DeliveryRussianPost,0)

	err := db.Model(&DeliveryRussianPost{}).Limit(limit).Offset(offset).Order(order).Find(&delivers, "account_id = ?", accountId).Error
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

func (deliveryRussianPost *DeliveryRussianPost) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(deliveryRussianPost).Omit("id", "account_id").Update(input).Error
}

func (deliveryRussianPost DeliveryRussianPost) delete () error {
	return db.Model(DeliveryRussianPost{}).Where("id = ?", deliveryRussianPost.ID).Delete(deliveryRussianPost).Error
}

// ########## End of CRUD Entity interface ###########

func (deliveryRussianPost DeliveryRussianPost) GetName () string {
	return "Почта России"
}

