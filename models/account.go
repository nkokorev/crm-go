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

	/*rows, err := base.GetPool().NamedQuery("INSERT INTO users (username,email,password) VALUES (:username, :email, :password) RETURNING *",u)
	if err != nil {return err}
	defer rows.Close()

	// сканируем результат RETURNING
	if rows.Next() {
		return rows.StructScan(&u)
	}
	return nil*/
}

// осуществляет поиск по a.ID
func (a *Account) Get () error {
	return GetPool().First(a.ID).Error
}

// сохраняет ВСЕ необходимые поля
func (a *Account) Save () error {
	return GetPool().Save(a).Error
}

// # Delete
func (a *Account) Delete () error {
	return GetDB().Model(a).Where("id = ?", a.ID).Delete(a).Error
}
