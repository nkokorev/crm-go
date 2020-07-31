package models

import "github.com/nkokorev/crm-go/utils"

type Delivery interface {
	Entity
	GetCode() string
	GetName() string
	GetVatCode() VatCode

	CalculateDelivery(DeliveryData, float64) (float64, error) // weight в кг
	checkMaxWeight(float64) error // проверяет макс вес

	setShopId(uint)
	AppendPaymentOptions([]PaymentOption) error
	RemovePaymentOptions([]PaymentOption) error
	ExistPaymentOption(PaymentOption) bool

	// new 31.07.2020
	/*AppendPaymentMethods([]PaymentMethod) error
	RemovePaymentMethods([]PaymentMethod) error
	ExistPaymentMethod(PaymentMethod) bool*/

	CreateDeliveryOrder(DeliveryData, PaymentAmount, Order) (Entity, error)
}

type DeliveryRequest struct {

	// Список товаров в корзине
	Cart []CartData `json:"cart"`

	// Данные для доставки
	DeliveryData	DeliveryData `json:"deliveryData"`
}

type DeliveryData struct {
	Id 		uint 	`json:"id"` 	// id доставки в ее таблице
	Code 	string 	`json:"code"`

	Address		string 	`json:"address"` 		// адрес доставки
	PostalCode	string 	`json:"postalCode"`
}

type CartData struct {
	ProductId 	uint	`json:"productId"`	// id product
	Quantity	uint	`json:"quantity"`	// число позиций
}

func (account Account) GetDeliveryMethods() []Delivery {
	// Находим все необходимые методы
	var posts []DeliveryRussianPost
	if err := db.Model(&DeliveryRussianPost{}).Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").
		Find(&posts, "account_id = ?", account.Id).Error; err != nil {
		return nil
	}

	var couriers []DeliveryCourier
	if err := db.Model(&DeliveryCourier{}).Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").
		Find(&couriers, "account_id = ?", account.Id).Error; err != nil {
		return nil
	}

	var pickups []DeliveryPickup
	if err := db.Model(&DeliveryPickup{}).Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").
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