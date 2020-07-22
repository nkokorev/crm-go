package models

type Delivery interface {
	Entity
	GetCode() string

	CalculateDelivery(deliveryData DeliveryData) (*DeliveryData, error) // weight в кг
	checkMaxWeight(deliveryData DeliveryData) error // проверяет макс вес

	setShopID(uint)
}

type DeliveryRequest struct {

	// Список товаров в корзине
	Cart map[string]struct{
		ProductID 	uint `json:"productID"` // id product
		Count 		uint `json:"count"`      // число позиций
	} `json:"cart"`

	// Метод доставки
	DeliveryMethod struct{
		ID 		uint 	`json:"id"` 	// id доставки в ее таблице
		Code 	string 	`json:"code"`	// code по которому можно понять что за таблица
		WebSiteID 	uint 	`json:"webSiteID"` // на всякий случай
	} `json:"deliveryMethod"`

	// Данные для расчета доставки
	DeliveryData DeliveryData `json:"deliveryData"`

}
// Данные для расчета доставки
type DeliveryData struct {
	Address		string `json:"address"` 		// адрес доставки
	PostalCode	string 	`json:"postalCode"` 	// Почтовый индекс доставки для расчета
	Comment		string `json:"comment"` 		// комментарий к доставке

	TotalCost 	float64 `json:"totalCost"` // общая стоимость доставки в рублях ! (расчетная величина)
	Weight 		float64 `json:"weight"` // итоговый вес посылки БРУТТО в кг ! (как правило расчетная величина)

	NeedToCalculateWeight bool  `json:"needToCalculateWeight"`	// необходимость расчета веса посылки
	ProductWeightKey string `json:"productWeightKey"` //  ключ для расчета веса продуктов в их атрибутах grossWeight
}