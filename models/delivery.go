package models

type Delivery interface {
	Entity
	GetCode() string

	CalculateDelivery(deliveryData DeliveryData, weight float64) (float64, error) // weight в кг
	checkMaxWeight(weight float64) error // проверяет макс вес

	setShopId(uint)
}

type DeliveryRequest struct {

	// Список товаров в корзине
	/*Cart map[string]struct{
		ProductId 	uint `json:"productId"` // id product
		Quantity 		uint `json:"quantity"`      // число позиций
	} `json:"cart"`*/
	Cart []CartData `json:"cart"`

	DeliveryData	DeliveryData `json:"deliveryData"`

	// Метод доставки
	/*DeliveryMethod struct {
		Id 		uint 	`json:"id"` 	// id доставки в ее таблице
		Code 	string 	`json:"code"`	// code по которому можно понять что за таблица
		WebSiteId 	uint 	`json:"webSiteId"` // на всякий случай
	} `json:"deliveryMethod"`

	DeliveryData DeliveryData `json:"deliveryData"`*/

}
// Данные для расчета доставки
/*type DeliveryData struct {
	
	Address		string 	`json:"address"` 		// адрес доставки
	PostalCode	string 	`json:"postalCode"` 	// Почтовый индекс доставки для расчета

	TotalCost 	float64 `json:"totalCost"` // общая стоимость доставки в рублях ! (расчетная величина)
	Weight 		float64 `json:"weight"` // расчетная величина внутри CRM - итоговый вес посылки БРУТТО в кг ! (как правило расчетная величина)

	NeedToCalculateWeight 	bool  `json:"needToCalculateWeight"`	// необходимость расчета веса посылки
	ProductWeightKey 		string `json:"productWeightKey"` //  ключ для расчета веса продуктов в их атрибутах grossWeight
}*/

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