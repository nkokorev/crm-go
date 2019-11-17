package models

import (
	"errors"
	"github.com/nkokorev/crm-go/database/base"
	u "github.com/nkokorev/crm-go/utils"

)

// Support Account Entity
type Category struct {
	ID        	uint 	`gorm:"primary_key;unique_index;" json:"-"` // внутренний ключ, не должен экспортироваться
	HashID 		string 	`json:"id" gorm:"type:varchar(10);unique_index;not null;"` // публичный уникальный ключ категории
	ParentID	uint	`json:"parent_id" gorm:"type:varchar(10);index;default:null;"` // is null == root

	Name 		string 	`json:"name" gorm:"not null;"` // Имя категории "Чай", "Зеленый чай", "Сапоги", "Скейтборды" и т.д.
	//Root		bool	`json:"root" gorm:"default:false;"` // корневая ли директория
	Products	[]Product `json:"products"` // hasMany ...

	AccountID 	uint 	`json:"-" gorm:"default:null;"` // аккаунт, к которой принадлежит категория (можно указывать только у родительской...)
	StoreID		uint	`json:"-" gorm:"default:null;"` 	// привязка root-category к магазину
}

// ### Account Entity model

// вспомогательная функция для получения ID
func (c Category) getID () uint { return c.ID }

// вспомогательная функция для получения ID
func (c Category) getAccountID () (id uint) { return c.AccountID }

// вспомогательная функция для установки аккаунта
func (c *Category) setAccountID (id uint) { c.AccountID = id }

// создает продукт в БД, устанавливая хеш ID
func (c *Category) create() (err error) {

	// проверяем на повторное создание (иначе будет апдейт)
	if c.ID > 0 {
		return errors.New("Can't create dublicate category")
	}

	// устанавливаем хеш продукта
	c.HashID, err = u.CreateHashID(c)
	if err != nil {
		return err
	}

	// создаем объект
	if err := base.GetDB().Create(c).Error; err != nil {
		return err
	}

	return nil
}

// полностью удаляет продукт из БД
func (c *Category) delete() (err error) {

	// проверяем чтобы объект имел реальный ID
	if c.ID < 1 {
		return errors.New("Can't delete category: ID category not found")
	}

	// создаем объект
	if err := base.GetDB().Delete(c).Error; err != nil {
		return err
	}

	return nil
}

// обновляет данные продукта в БД с защитой служебных полей
func (c *Category) update() (err error) {

	// указываем какие поля обновлять не надо
	if err := base.GetDB().Model(&c).Omit("id", "hash_id", "account_id").Updates(c).Error; err != nil {
		return err
	}

	return nil
}

// ищет продукт по hashID в БД. Возвращает ошибку, если продукт не найден или еще что-то пошло не так
func (c *Category) get(hash_id string) error {

	if err := base.GetDB().First(c,"hash_id = ?", hash_id).Error;err != nil {
		return err
	}
	return nil
}


// ###

// Проверяет, является ли категория корневой
func (c Category) isRoot() bool {
	return c.ParentID > 1
}

