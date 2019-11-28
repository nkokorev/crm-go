package models

import (
	"time"
)

type Account struct {
	ID uint `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	CreatedAt 	time.Time `json:"created_at" db:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt 	*time.Time `json:"-" db:"deleted_at"`
}

// создает аккаунт, несмотря на a.ID
func (a *Account) Create () error {
	return GetPool().Create(a).Error
}

// осуществляет поиск по a.ID
func (a *Account) Get () error {
	return GetPool().First(a.ID).Error
}

// сохраняет ВСЕ необходимые поля, кроме id, deleted_at и возвращает в Account обновленные данные
func (a *Account) Save () error {
	return db.Model(Account{}).Omit("id", "deleted_at").Save(a).Find(a, "id = ?", a.ID).Error
}

// обновляет данные аккаунта кроме id, deleted_at и возвращает в Account обновленные данные
func (a *Account) Update (input interface{}) error {
	return db.Model(Account{}).Where("id = ?", a.ID).Omit("id", "deleted_at").Update(input).Find(a, "id = ?", a.ID).Error
}

// # Delete
func (a *Account) Delete () error {
	return db.Model(Account{}).Where("id = ?", a.ID).Delete(a).Error
}
