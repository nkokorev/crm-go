package models

import "github.com/jinzhu/gorm"

// Хелп таблица ManyToMany PaymentMethods <> Delivery
type Payment2Delivery struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"-" gorm:"index;not null"`
	WebSiteId	uint 	`json:"webSiteId" gorm:"type:int;index;default:NULL;"`

	PaymentId		uint	`json:"paymentId" gorm:"type:int;not null;"`
	PaymentType	string	`json:"paymentType" gorm:"varchar(32);not null;"`

	DeliveryId	uint	`json:"deliveryId" gorm:"type:int;not null;"`
	DeliveryType	string	`json:"deliveryType" gorm:"varchar(32);not null;"`
}

func (Payment2Delivery) PgSqlCreate() {
	db.AutoMigrate(&Payment2Delivery{})
	db.Model(&Payment2Delivery{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&Payment2Delivery{}).AddForeignKey("web_site_id", "web_sites(id)", "CASCADE", "CASCADE")
}
func (p2d *Payment2Delivery) BeforeCreate(scope *gorm.Scope) error {
	p2d.Id = 0
	return nil
}
func (Payment2Delivery) TableName() string {
	return "payment_to_delivery"
}

// Добавляет, если есть - ничего не делает
func  (webSite WebSite) AppendPayment2Delivery(paymentId uint, paymentType string, deliveryId uint, deliveryType string) error {
	if err := db.Model(&Payment2Delivery{}).FirstOrCreate(Payment2Delivery{
		AccountId: webSite.AccountId,
		WebSiteId: webSite.Id,
		PaymentId:  paymentId,
		PaymentType: paymentType,
		DeliveryId: deliveryId,
		DeliveryType: deliveryType,
	}).Error; err != nil {
		return err
	}

	return nil
}

func  (p2d Payment2Delivery) Append(paymentMethod PaymentMethod, delivery Delivery) {

}

func  (Payment2Delivery) RemoveByIds(paymentMethodId uint, paymentMethodType string, paymentDeliveryId uint, paymentDeliveryType string) {

}
