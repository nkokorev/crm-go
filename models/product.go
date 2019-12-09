package models

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

type Product struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`

	// Article string `json:"article"` // артикул товара из иных соображений (часто публичный)
	SKU string `json:"sku"` // складской индектификатор

	Name string `json:"name"`

	Account Account `json:"-"`
	// Offers []ProductOffer `json:"offers"`
}

func (p *Product) Create() error {

	if p.Name == "" {
		return utils.Error{Message:"Ошибки при создании продукта", Errors: map[string]interface{} {"name":"Имя продукта обязательно к заполнению"} }
	}

	if p.ExistSKU() {
		return utils.Error{Message: fmt.Sprintf("Продукт с SKU = [%v] уже существует",p.SKU) }
	}

	// создаем продукт
	if err := db.Create(p).Error; err != nil {
		return err
	}

	return nil
}

func (p *Product) Get () error {

	// чекаем на всякий случай ID аккаунта
	if p.AccountID < 1 {
		return errors.New("Необходимо указать Account ID")
	}

	err := db.Model(Product{}).Where("id = ? AND account_id = ?", p.ID, p.AccountID).First(p).Error;

	if err != nil && err == gorm.ErrRecordNotFound {
		return utils.Error{Message:fmt.Sprintf("Указанный продукт с id = %v не найден.", p.ID)}
	}
	return err
}

func (Product) GetAll(account_id uint) (products []Product, err error) {
	err = db.Model(Product{}).Order("id asc").Where("account_id = ?", account_id).Find(&products).Error
	return
}

func (p *Product) Save() error {

	// чекаем на всякий случай ID аккаунта, в контексте которого происходит выполнение
	if p.AccountID < 1 {
		return utils.Error{Message:"Непредвиденная ошибка синхронизации с аккаунтом"}
	}

	// проверяем, что нет совпадающих значений, исключая текущее значение
	if  p.SKU != "" && !db.Unscoped().First(&Product{},"account_id = ? AND sku = ? AND id != ?", p.AccountID, p.SKU, p.ID).RecordNotFound() {
		return utils.Error{Message:"Ошибки при обновлении данных товара", Errors: map[string]interface{} {"sku":"Повторяющиеся значение параметра"} }
	}

	// обновляем данные
	err :=  db.Model(Product{}).Where("id = ? AND account_id = ?", p.ID, p.AccountID).Omit("id", "account_id").
		Save(p).Find(p, "id = ?", p.ID).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		return errors.New(fmt.Sprintf("Ошибка при сохранении: склада не найден id  = %v", p.ID))
	}

	return err
}

func (p *Product) Update(input interface{}) error {

	// чекаем на всякий случай ID аккаунта, в контексте которого происходит выполнение
	if p.AccountID < 1 {
		return utils.Error{Message:"Непредвиденная ошибка синхронизации с аккаунтом"}
	}

	// проверяем, что нет совпадающих значений, исключая текущее значение stock
	newSKU := input.(map[string]interface{})["sku"]
	if  newSKU != nil && !db.Unscoped().First(&Product{},"account_id = ? AND sku = ? AND id != ?", p.AccountID, newSKU, p.ID).RecordNotFound() {
		return utils.Error{Message:"Ошибки при обновлении данных продукта", Errors: map[string]interface{} {"sku":"Повторяющиеся значение параметра"} }
	}

	// обновляем данные
	err :=  db.Model(Product{}).Where("id = ? AND account_id = ?", p.ID, p.AccountID).Omit("id", "account_id").
		Update(input).Find(p, "id = ?", p.ID).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		return utils.Error{Message:fmt.Sprintf("Невозможно обновить продукт, указанный id = %v не найден.", p.ID)}
	}
	return err
}

func (p *Product) Delete() error {
	// чекаем на всякий случай ID аккаунта, в контексте которого происходит выполнение
	if p.AccountID < 1 {
		return utils.Error{Message:"Непредвиденная ошибка синхронизации с аккаунтом"}
	}

	if p.ID < 1 {
		return utils.Error{Message:"Неуказан ID удаляемого товара"}
	}

	// удаляем данные. Если объект не будет найден - ошибки не будет.
	return db.Model(Product{}).Where("id = ? AND account_id = ?", p.ID, p.AccountID).Delete(p).Error
}

func (p Product) ExistSKU() bool {
	return !db.Unscoped().First(&Product{},"account_id = ? AND sku = ?", p.AccountID, p.SKU).RecordNotFound()
}

func (p Product) Exist() bool {
	return !db.Unscoped().First(&Product{},"id = ? AND account_id = ?", p.ID, p.AccountID).RecordNotFound()
}