package models

// Список товаров на складе
type StockProduct struct {
	Id uint	`json:"id"`
	AccountId uint `json:"-"`
	ProductId uint `json:"productId"`


	Product Product `json:"-"`
}


func (stock *Stock) AddProduct (product Product, count interface{}) error {
	return nil
}

func (stock *Stock) RemoveProduct (product Product, count interface{}) error {
	return nil
}

func (stock *Stock) ReserveProduct (product Product, count interface{}) error {
	return nil
}