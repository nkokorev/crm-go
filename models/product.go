package models

import (
	"errors"
	"fmt"
	"github.com/nkokorev/crm-go/database/base"
	"os"

	//e "github.com/nkokorev/crm-go/errors"
	//t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
)

// Entity compare
type Product struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string 	`json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Name 		string 	`json:"name" gorm:"not null;"`
	Account    Account 	`json:"account" gorm:"not null;"`
	AccountID 	uint 	`json:"-"`
}

// Все связывающие функции внутренние и вызываются в контексте аккаунта. Безопасные функции экспортируются и их можно вызвать напрямую.

// вспомогательная функция для получения ID
func (p Product) getID () uint { return p.ID }

func (p Product) getHashID () string { return p.HashID }

// создает продукт, устанавливая хеш ID
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

// обновляет данные продукта с защитой служебных полей
func (product *Product) update() (err error) {

	// указываем какие поля обновлять не надо
	if err := base.GetDB().Model(&product).Omit("id", "hash_id", "account_id").Updates(product).Error; err != nil {
		return err
	}

	return nil
}

// ищет продукт по hashID. Возвращает ошибку, если продукт не найден или еще что-то пошло не так
func (product *Product) getByHashID(hash_id string) error {

	if err := base.GetDB().First(product,"hash_id = ?", hash_id).Error;err != nil {
		return err
	}
	return nil
}

// вспомогательная функция для получения ID
func (p Product) getAccountID () (id uint) { return p.AccountID }

// вспомогательная функция для получения ID
func (p *Product) setAccountID (id uint) { p.AccountID = id }

// ## Helpers func



func init() {

	return
	if os.Getenv("ENV_VAR") == "test" {
		return
	}

	db := base.GetDB()

	db.DropTableIfExists(&Product{})
	db.AutoMigrate(&Product{})

	db.Table("products").AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")

	account := Account{}
	db.Find(&account,2)

	products := []Product{
		{Name:"Дянь хун", Account:account},
		{Name:"Дянь мун", Account:account},
		{Name:"Дянь сан", Account:account},
	}

	for i,v := range products{
		err := v.create()
		if err != nil { fmt.Printf("Охтыж боже мой... не удалось: %v %v, err: %v \r\n", i, v.Name, err);
			return
		}
	}

	test_product := Product{}
	db.LogMode(true)

	db.Preload("Account").Find(&test_product,1)

	fmt.Println("Мы нашли продукт: ", test_product)
	fmt.Println("У аккаунта: ", test_product.Account.Name)

}