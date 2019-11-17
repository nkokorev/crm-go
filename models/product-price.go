package models

// структура, которая хранит данные о ценах, об их изменениях (временные цены с какого-то до какого-то момента). Для услуг и товаров
type EntityPrice struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"` // внутренний ключ, не должен экспортироваться
	HashID 		string 	`json:"id" gorm:"type:varchar(10);unique_index;not null;"` // публичный уникальный ключ категории
	AccountID 	uint 	`json:"-" gorm:"default:null;"` // аккаунт, к которой принадлежит

	Price		float64	`json:"price" gorm:"default:(0,0))"` // отпускная цена
	PurchasePrice		float64	`json:"purchase_price" gorm:"default:(0,0))"` // закупочная цена
	//Cost		float64	`json:"cost" gorm:"default:(0,0))"`

	// ### ниже надо доработать
	From 	uint

	// Стоимость у каждого сайта может быть своя (ололошеньки)
	StoreWebSite	[]StoreWebSite `gorm:"many2many:product_price_store_website"` // если null, то стоимость одна на все сайты

	// Стоимость у категории пользователей (Customer) может быть своя (совсем ололошеньки)
	CustomerGroup 	[]CustomerGroup `gorm:"many2many:product_price_customer_group"` // если null, то один на всех
}
