package models

import (
	"github.com/jinzhu/gorm"
)

type Delivery struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint	`json:"accountId" gorm:"index,not null"` // аккаунт-владелец ключа
	ShopID		uint 	`json:"shopId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится

	Name string `json:"name" gorm:"type:varchar(255);"` // "Доставка курьером", "Доставка Почтой России", "Самовывоз"
	Enabled bool `json:"enabled" gorm:"type:bool;default:true"` // активна ли форма доставки
}

func (Delivery) PgSqlCreate() {
	db.CreateTable(&Delivery{})
	
	db.Model(&Delivery{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&Delivery{}).AddForeignKey("shop_id", "shops(id)", "CASCADE", "CASCADE")
}

func (delivery Delivery) getId() uint {
	return delivery.ID
}
func (delivery *Delivery) setId(id uint) {
	delivery.ID = id
}
func (delivery Delivery) GetAccountId() uint {
	return delivery.AccountID
}
func (delivery *Delivery) setAccountId(id uint) {
	delivery.AccountID = id
}
func (Delivery) getEntityName() string {
	return "Delivery"
}
func (Delivery) TableName() string {
	return "deliveries"
}
func (delivery *Delivery) BeforeCreate(scope *gorm.Scope) error {
	delivery.ID = 0
	return nil
}

/*func (Delivery) GetNullArray() []interface{} {
	return make([]Delivery,0)
}*/

// ############# Entity interface #############

func (delivery Delivery) create() (Entity, error)  {
	var newItem Entity = &delivery
	
	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (Delivery) get(id uint) (Entity, error) {

	var delivery Delivery

	err := db.First(&delivery, id).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (delivery *Delivery) load() error {

	err := db.First(delivery).Error
	if err != nil {
		return err
	}
	return nil
}

func (Delivery) GetPaginationList(accountId uint, offset, limit int, order string, search *string) ([]Entity, error) {

	delivers := make([]Delivery,0)

	err := db.Model(&Delivery{}).Limit(limit).Offset(offset).Order(order).Find(&delivers, "account_id = ?", accountId).Error
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

func (delivery *Delivery) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(delivery).Omit("id", "account_id").Update(input).Error
}
func (delivery Delivery) delete () error {
	return db.Model(Delivery{}).Where("id = ?", delivery.ID).Delete(delivery).Error
}

/*

func (apiKey ApiKey) delete () error {
	return db.Model(ApiKey{}).Where("id = ?", apiKey.ID).Delete(apiKey).Error
}

func (apiKey *ApiKey) update(input interface{}) error {
	// return db.Model(apiKey).Omit("token", "account_id", "created_at", "updated_at").Select("Name", "Enabled").Updates(&input).Error
	return db.Model(apiKey).Select("Name", "Enabled").Updates(structs.Map(input)).Error

}*/

// ######## !!!! Все что выше покрыто тестами на прямую или косвено
