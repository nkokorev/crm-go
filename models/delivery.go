package models

type Delivery interface {
	Entity
	GetCode() string
	GetName() string
	GetVatCode() VatCode

	CalculateDelivery(deliveryData DeliveryData, weight float64) (float64, error) // weight в кг
	checkMaxWeight(weight float64) error // проверяет макс вес

	setShopId(uint)
	AppendPaymentOptions(paymentOptions []PaymentOption) error
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