package models

type Delivery interface {
	Entity
	GetCode() string
}

type DeliveryRequest struct {

	// Корзина
	Cart map[string]struct{
		ProductId 	uint `json:"productId"` // id product
		Count 		uint `json:"count"`      // число позиций
	} `json:"cart"`

	// Метод доставки
	DeliveryMethod struct{
		ID 		uint `json:"id"` 		// id доставки в ее таблице
		Code 	string `json:"code"`	// code по которому можно понять что за таблица
		ShopId 	uint `json:"shopId"`    // на всякий случай
	} `json:"deliveryMethod"`

	// Данные для расчета доставки
	Data struct{
		Address	string `json:"address"` 		// id доставки в ее таблице
	} `json:"data"`
}
