package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

// Физический склад с набором методов
type Stock struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`

	Code string `json:"code" gorm:"default:NULL"`
	Name string `json:"name"`
	Address string `json:"address"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"-" sql:"index"`

	Products []StockProduct `json:"-"`

}

func (stock *Stock) create () error {

	if stock.Name == "" {
		return utils.Error{Message:"Ошибки при создании склада", Errors: map[string]interface{} {"name":"Имя склада обязательно к заполнению"} }
	}

	if stock.ExistCode() {
		return utils.Error{Message:"Ошибки при создании склада", Errors: map[string]interface{} {"code":"Повторяющиеся значение параметра"} }
	}
	
	return db.Omit("id","created_at", "deleted_at").Create(stock).Error
}


func (Stock) GetAll(account_id uint, stocks *[]Stock) error {
	//stocks = make([]Stock, 0)
	return db.Model(Stock{}).Order("id asc").Where("account_id = ?", account_id).Find(stocks).Error
}

func (stock *Stock) Get(account_id, stock_id uint) error {
	err := db.Model(Stock{}).Where("id = ? AND account_id = ?", stock_id, account_id).First(stock).Error;

	if err != nil && err == gorm.ErrRecordNotFound {
		return utils.Error{Message:fmt.Sprintf("Указанный склад с id = %v не найден.", stock_id)}
	}
	return err
}

type Example struct {
	Code string
}

func (stock *Stock) Update(input interface{}) error {

	// чекаем на всякий случай ID аккаунта, в контексте которого происходит выполнение
	if stock.AccountID < 1 {
		return utils.Error{Message:"Непредвиденная ошибка синхронизации с аккаунтом"}
	}

	// проверяем, что нет совпадающих значений, исключая текущее значение stock
	if !db.Unscoped().First(&Stock{},"account_id = ? AND code = ? AND id != ?", stock.AccountID, input.(map[string]interface{})["code"], stock.ID).RecordNotFound() {
		return utils.Error{Message:"Ошибки при обновлении данных склада", Errors: map[string]interface{} {"code":"Повторяющиеся значение параметра"} }
	}

	// обновляем данные
	err :=  db.Model(Stock{}).Where("id = ? AND account_id = ?", stock.ID, stock.AccountID).Omit("id", "account_id", "created_at", "deleted_at", "updated_at").
		Update(input).Find(stock, "id = ?", stock.ID).Error

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

func (Stock) Exist(account_id,stock_id uint ) bool {

	return !db.Unscoped().First(&Stock{},"stock_id = ? AND account_id = ?", stock_id, account_id).RecordNotFound()
}