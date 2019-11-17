package models

import (
	"errors"
	"github.com/nkokorev/crm-go/database/base"
	//e "github.com/nkokorev/crm-go/errors"
	//t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
)

// Support Account Entity
type Product struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Account    Account 	`json:"-" gorm:"not null;"`
	AccountID 	uint 	`json:"-" gorm:"index,not null;"` // к какому аккаунту он принадлежит

	Name 		string 	`json:"name" gorm:"not null;"`
	Sku 		string 	`json:"sku" gorm:"not null;"` // Stock Keeping Unit

	Price		float64	`json:"price" gorm:"default:0.0"` // стоимость товара
	//ProductPrice		EntityPrice	`json:"product_price"` // загружаем в структуру продукта
	//ProductPriceID		uint	`json:"-" gorm:"default:null"` // стоимость товара

	// ### Атрибуты
	EavAttrSet	EavAttrSet	`json:"eav_attr_set" gorm:"default:null;"` // какой набор атрибутов будет у продукта
	EavAttrSetID	uint	`json:"-" gorm:"default:null;"` // какой набор атрибутов будет у продукта
	EavAttrs	[]EavAttr	`json:"-" gorm:"many2many:eav_attr_products"` // какие атрибуты дополнительно будут у продукта | ManyToMany
	Categories 	[]Category	`json:"-" gorm:"many2many:category_products"` // к какой категории относится продукт | ManyToMany


	//EavAttrGroup		[]EavAttrGroup 	`gorm:"many2many:eav_attr_group_product;"`
}

// Все связывающие функции внутренние и вызываются в контексте аккаунта. Безопасные функции экспортируются и их можно вызвать напрямую.

// вспомогательная функция для получения ID
func (p Product) getID () uint { return p.ID }

// создает продукт в БД, устанавливая хеш ID
func (product *Product) create() (err error) {

	// проверяем на повторное создание (иначе будет апдейт)
	if product.ID > 0 {
		return errors.New("Cant create dublicate product")
	}

	// устанавливаем хеш продукта
	product.HashID, err = u.CreateHashID(product)
	if err != nil {
		return err
	}

	// создаем объект
	if err := base.GetDB().Create(product).Error; err != nil {
		return err
	}

	return nil
}

// полностью удаляет продукт из БД
func (product *Product) delete() (err error) {

	// проверяем чтобы объект имел реальный ID
	if product.ID < 1 {
		return errors.New("cant delete productL: ID product not found")
	}

	// создаем объект
	if err := base.GetDB().Delete(product).Error; err != nil {
		return err
	}

	return nil
}

// обновляет данные продукта в БД с защитой служебных полей
func (product *Product) update() (err error) {

	// указываем какие поля обновлять не надо
	if err := base.GetDB().Model(&product).Omit("id", "hash_id", "account_id").Updates(product).Error; err != nil {
		return err
	}

	return nil
}

// ищет продукт по hashID в БД. Возвращает ошибку, если продукт не найден или еще что-то пошло не так
func (product *Product) get(hash_id string) error {

	if err := base.GetDB().First(product,"hash_id = ?", hash_id).Error;err != nil {
		return err
	}
	return nil
}

// вспомогательная функция для получения ID
func (p Product) getAccountID () (id uint) { return p.AccountID }

// вспомогательная функция для установки аккаунта
func (p *Product) setAccountID (id uint) { p.AccountID = id }
