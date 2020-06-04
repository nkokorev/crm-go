package models

import "time"

type Order struct {
	ID uint	`json:"id" gorm:"primary_key;"` // внутренний id

	OrderKey uint `json:"publicId" gorm:"AUTO_INCREMENT"` // публичный порядковый ID офера в контексте аккаунта

	AccountID uint `json:"-" gorm:"index;not null;"`
	UserID uint `json:"userId" gorm:"index;default:null;"` // по-умолчанию ни к чему не привязываем

	User   User    `json:"user"`                                 // может быть нулевым
	// Offers []Offer `json:"offers" gorm:"many2many:order_offers"` // могут быть оферы с одним товаром
	//Products []Product `json:"products" gorm:"many2many:order_products"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

func (Order) PgSqlCreate() {
	db.CreateTable(&Order{})

	db.Exec("ALTER TABLE orders \n--     ALTER COLUMN parent_id SET DEFAULT NULL,\n    ADD CONSTRAINT orders_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    ADD CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE;\n")

}

func (order Order) create() (*Order,error)  {

	var outOrder Order
	var err error

	// account.GetLastOrderID()
	order.OrderKey = 1

	if err := db.Create(&outOrder).Error; err != nil {
		return nil, err
	}

	return &outOrder, err
}

func (order Order) delete () error {
	return db.Model(&Order{}).Unscoped().Where("id = ?", order.ID).Delete(order).Error
}
