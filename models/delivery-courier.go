package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type DeliveryCourier struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint	`json:"-" gorm:"index,not null"` // аккаунт-владелец ключа
	ShopID		uint 	`json:"shopId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится
	Code 		string	`json:"code" gorm:"type:varchar(16);default:'courier';"` // Для идентификации во фронтенде

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // активен ли способ доставки
	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Курьерская доставка", "Почта России", "Самовывоз"
	Price 		float64 `json:"price" gorm:"type:numeric;default:0"` // стоимость доставки

	MaxWeight 	float64 `json:"maxWeight" gorm:"type:int;default:40"` // максимальная масса в кг

	AddressRequired	bool	`json:"addressRequired" gorm:"type:bool;default:true"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postalCodeRequired" gorm:"type:bool;default:false"` // Требуется ли индекс в адресе доставки

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (DeliveryCourier) PgSqlCreate() {
	db.CreateTable(&DeliveryCourier{})
	db.Model(&DeliveryCourier{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

// ############# Entity interface #############
func (deliveryCourier DeliveryCourier) getId() uint { return deliveryCourier.ID }
func (deliveryCourier *DeliveryCourier) setId(id uint) { deliveryCourier.ID = id }
func (deliveryCourier DeliveryCourier) GetAccountId() uint { return deliveryCourier.AccountID }
func (deliveryCourier *DeliveryCourier) setAccountId(id uint) { deliveryCourier.AccountID = id }
func (deliveryCourier *DeliveryCourier) setShopId(shopId uint) { deliveryCourier.ShopID = shopId }

func (deliveryCourier DeliveryCourier) GetCode() string {
	return deliveryCourier.Code
}
// ############# Entity interface #############
func (deliveryCourier *DeliveryCourier) BeforeCreate(scope *gorm.Scope) error {
	deliveryCourier.ID = 0
	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (deliveryCourier DeliveryCourier) create() (Entity, error)  {
	var newItem Entity = &deliveryCourier
	
	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (DeliveryCourier) get(id uint) (Entity, error) {

	var deliveryCourier DeliveryCourier

	err := db.First(&deliveryCourier, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryCourier, nil
}

func (deliveryCourier *DeliveryCourier) load() error {

	err := db.First(deliveryCourier).Error
	if err != nil {
		return err
	}
	return nil
}

func (DeliveryCourier) getPaginationList(accountId uint, offset, limit int, order string, search *string) ([]Entity, error) {

	delivers := make([]DeliveryCourier,0)

	err := db.Model(&DeliveryCourier{}).Limit(limit).Offset(offset).Order(order).Find(&delivers, "account_id = ?", accountId).Error
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

func (deliveryCourier *DeliveryCourier) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(deliveryCourier).Omit("id", "account_id").Update(input).Error
}

func (deliveryCourier DeliveryCourier) delete () error {
	return db.Model(DeliveryCourier{}).Where("id = ?", deliveryCourier.ID).Delete(deliveryCourier).Error
}

// ########## End of CRUD Entity interface ###########

func (deliveryCourier DeliveryCourier) GetName () string {
	return "Доставка курьером"
}

func (deliveryCourier DeliveryCourier) CalculateDelivery(deliveryData DeliveryData) (*DeliveryData, error) {

	deliveryData.TotalCost = deliveryCourier.Price
	return &deliveryData, nil
}

func (deliveryCourier DeliveryCourier) checkMaxWeight(deliveryData DeliveryData) error {
	// проверяем максимальную массу:
	if deliveryData.Weight > deliveryCourier.MaxWeight {
		return utils.Error{Message: fmt.Sprintf("Превышен максимальный вес посылки в %vкг.", deliveryCourier.MaxWeight)}
	}

	return nil
}

