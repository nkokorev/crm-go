package models

type Delivery interface {
	Entity
	GetCode() string
	CalculateDelivery(deliveryData DeliveryData, weight uint) (float64, error)
}

type DeliveryRequest struct {

	// Список товаров в корзине
	Cart map[string]struct{
		ProductId 	uint `json:"productId"` // id product
		Count 		uint `json:"count"`      // число позиций
	} `json:"cart"`

	// Метод доставки
	DeliveryMethod struct{
		ID 		uint 	`json:"id"` 	// id доставки в ее таблице
		Code 	string 	`json:"code"`	// code по которому можно понять что за таблица
		ShopId 	uint 	`json:"shopId"` // на всякий случай
	} `json:"deliveryMethod"`

	// Данные для расчета доставки
	DeliveryData DeliveryData `json:"deliveryData"`

}

type DeliveryData struct {
	Address		string `json:"address"` 		// id доставки в ее таблице
	PostalCode	string 	`json:"postalCode"` 	// Почтовый индекс для расчета
	Comment		string `json:"comment"` 		// коммент к доставке
	ProductWeightKey string `json:"productWeightKey"` //  grossWeight
}