package models

import (
	"github.com/nkokorev/crm-go/utils"
)

type Delivery interface {
	Entity
	GetCode() string
	GetType() string
	GetName() string
	GetVatCode() (*VatCode, error)
	GetPaymentSubject() PaymentSubject
	GetWebSiteId() uint

	CalculateDelivery(DeliveryData, float64) (float64, error) // weight в кг
	checkMaxWeight(float64) error // проверяет макс вес

	setWebSiteId(uint)
	// AppendPaymentOptions([]PaymentOption) error
	// RemovePaymentOptions([]PaymentOption) error


	// new 31.07.2020
	AppendPaymentMethods([]PaymentMethod) error
	RemovePaymentMethods([]PaymentMethod) error
	ExistPaymentMethod(method PaymentMethod) bool

	// Создавать ли ордер на доставку или это моментальная выдача товара
	// NeedToCreateDeliveryOrder() bool

	// getListByShop(accountId, websiteId uint) (interface{}, error)

	CreateDeliveryOrder(DeliveryData, PaymentAmount, Order) (Entity, error)
}

type DeliveryRequest struct {

	// Список товаров в корзине
	Cart []CartData `json:"cart"`

	// Данные для доставки
	DeliveryData	DeliveryData `json:"delivery_data"`
}

type DeliveryData struct {
	Id 		uint 	`json:"id"` 	// id доставки в ее таблице
	Code 	string 	`json:"code"`

	Address		string 	`json:"address"` 		// адрес доставки
	PostalCode	string 	`json:"postal_code"`
}

type CartData struct {
	Id 	uint	`json:"id"`	// id product
	Quantity	uint	`json:"quantity"`	// число позиций
}

func (account Account) GetDeliveryMethods() []Delivery {
	// Находим все необходимые методы
	var posts []DeliveryRussianPost
	if err := db.Model(&DeliveryRussianPost{}).Preload("PaymentSubject").Preload("VatCode").
		Find(&posts, "account_id = ?", account.Id).Error; err != nil {
		return nil
	}

	var couriers []DeliveryCourier
	if err := db.Model(&DeliveryCourier{}).Preload("PaymentSubject").Preload("VatCode").
		Find(&couriers, "account_id = ?", account.Id).Error; err != nil {
		return nil
	}

	var pickups []DeliveryPickup
	if err := db.Model(&DeliveryPickup{}).Preload("PaymentSubject").Preload("VatCode").
		Find(&pickups, "account_id = ?", account.Id).Error; err != nil {
		return nil
	}

	deliveries := make([]Delivery, len(posts)+len(pickups)+len(couriers))
	for i,_ := range posts {
		deliveries[i] = &posts[i]
	}
	for i,_ := range couriers {
		deliveries[i+len(posts)] = &couriers[i]
	}
	for i,_ := range pickups {
		deliveries[i+len(posts)+len(couriers)] = &pickups[i]
	}

	return deliveries
}

func (account Account) GetDeliveryByCode(code string, methodId uint) (Delivery, error){

	// 1. Получаем все варианты доставки (обычно их мало). Можно через switch, но лень потом исправлять баг с новыми типом доставки
	deliveries := account.GetDeliveryMethods()


	// Ищем наш вариант доставки
	var delivery Delivery
	for _,v := range deliveries {
		if v.GetCode() == code && v.GetId() == methodId {
			delivery = v
			break
		}
	}

	// Проверяем, удалось ли найти выбранный вариант доставки
	if delivery == nil {
		return nil, utils.Error{Message: "Не верно указан тип доставки"}
	}

	return delivery, nil
}

// Для получения методов оплаты 
func GetPaymentMethodsByDelivery(delivery Delivery) ([]PaymentMethod, error){
	// Get ALL Payment Methods
	paymentCashes, err := PaymentCash{}.GetListByWebSiteAndDelivery(delivery)
	if err != nil { return nil, err }
	paymentYandexes, err := PaymentYandex{}.GetListByWebSiteAndDelivery(delivery)
	if err != nil { return nil, err }

	methods := make([]PaymentMethod, len(paymentYandexes) + len(paymentCashes))
	for i := range paymentCashes {
		methods[i] = &paymentCashes[i]
	}
	for i := range paymentYandexes {
		methods[i + len(paymentCashes)] = &paymentYandexes[i]
	}
	return methods, nil
}