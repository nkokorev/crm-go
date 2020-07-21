package models

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
)

// Физический склад с набором методов
type Stock struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`

	Code string `json:"code" gorm:"default:NULL"`
	Name string `json:"name"`
	Address string `json:"address"`

	Products []StockProduct `json:"-"`

}

func (stock *Stock) Create () error {

	// чекаем на всякий случай ID аккаунта
	if stock.AccountID < 1 {
		return errors.New("Необходимо указать Account ID")
	}

	if stock.Name == "" {
		return utils.Error{Message:"Ошибки при создании склада", Errors: map[string]interface{} {"name":"Имя склада обязательно к заполнению"} }
	}

	if stock.ExistCode() {
		return utils.Error{Message:"Ошибки при создании склада", Errors: map[string]interface{} {"code":"Повторяющиеся значение параметра"} }
	}
	
	return db.Omit("id").Create(stock).Error
}

func (stock *Stock) Get() error {

	// чекаем на всякий случай ID аккаунта
	if stock.AccountID < 1 {
		return errors.New("Необходимо указать accountID")
	}

	err := db.Model(Stock{}).Where("id = ? AND account_id = ?", stock.ID, stock.AccountID).First(stock).Error;

	if err != nil && err == gorm.ErrRecordNotFound {
		return utils.Error{Message:fmt.Sprintf("Указанный склад с id = %v не найден.", stock.ID)}
	}
	return err
}

func (Stock) GetAll(account_id uint) (stocks []Stock, err error) {
	err = db.Model(Stock{}).Order("id asc").Where("account_id = ?", account_id).Find(&stocks).Error
	return
	//return db.Model(Stock{}).Order("id asc").Where("account_id = ?", account_id).Find(stocks).Error
}

func (stock *Stock) Save() error {

	// чекаем на всякий случай ID аккаунта, в контексте которого происходит выполнение
	if stock.AccountID < 1 {
		return utils.Error{Message:"Непредвиденная ошибка синхронизации с аккаунтом"}
	}

	// проверяем, что нет совпадающих значений, исключая текущее значение stock
	if  stock.Code != "" && !db.Unscoped().First(&Stock{},"account_id = ? AND code = ? AND id != ?", stock.AccountID, stock.Code, stock.ID).RecordNotFound() {
		return utils.Error{Message:"Ошибки при обновлении данных склада", Errors: map[string]interface{} {"code":"Повторяющиеся значение параметра"} }
	}

	// обновляем данные
	err :=  db.Model(Stock{}).Where("id = ? AND account_id = ?", stock.ID, stock.AccountID).Omit("id", "account_id").
		Save(stock).Find(stock, "id = ?", stock.ID).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		return errors.New(fmt.Sprintf("Ошибка при сохранении: склада не найден id  = %v", stock.ID))
	}

	return err
}

func (stock *Stock) Update(input interface{}) error {

	// чекаем на всякий случай ID аккаунта, в контексте которого происходит выполнение
	if stock.AccountID < 1 {
		return utils.Error{Message:"Непредвиденная ошибка синхронизации с аккаунтом"}
	}

	// проверяем, что нет совпадающих значений, исключая текущее значение stock
	newCode := input.(map[string]interface{})["code"]
	if  newCode != nil && !db.Unscoped().First(&Stock{},"account_id = ? AND code = ? AND id != ?", stock.AccountID, newCode, stock.ID).RecordNotFound() {
		return utils.Error{Message:"Ошибки при обновлении данных склада", Errors: map[string]interface{} {"code":"Повторяющиеся значение параметра"} }
	}

	// обновляем данные
	err :=  db.Model(Stock{}).Where("id = ? AND account_id = ?", stock.ID, stock.AccountID).Omit("id", "account_id").
		Updates(input).Find(stock, "id = ?", stock.ID).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		return utils.Error{Message:fmt.Sprintf("Невозможно обновить склад, указанный id = %v не найден.", stock.ID)}
	}
	return err
}

func (stock *Stock) Delete() error {

	// чекаем на всякий случай ID аккаунта, в контексте которого происходит выполнение
	if stock.AccountID < 1 {
		return utils.Error{Message:"Непредвиденная ошибка синхронизации с аккаунтом"}
	}

	if stock.ID < 1 {
		return utils.Error{Message:"Неуказан ID удаляемого склада"}
	}

	// удаляем данные. Если объект не будет найден - ошибки не будет.
	return db.Model(Stock{}).Where("id = ? AND account_id = ?", stock.ID, stock.AccountID).Delete(stock).Error
}

func (stock Stock) ExistCode() bool {
	return !db.Unscoped().First(&Stock{},"account_id = ? AND code = ?", stock.AccountID, stock.Code).RecordNotFound()
}

func (stock Stock) Exist() bool {
	return !db.Unscoped().First(&Stock{},"id = ? AND account_id = ?", stock.ID, stock.AccountID).RecordNotFound()
}