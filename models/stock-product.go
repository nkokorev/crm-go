package models

// Список товаров на складе
type StockProduct struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`
	ProductID uint `json:"productID"`


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