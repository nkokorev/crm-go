package models

import (

	"github.com/nkokorev/crm-go/database/base"
	"time"
)

type Account struct {
	ID uint `json:"id"`
	Name string `json:"name"`
	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
	DeletedAt 	*time.Time `json:"-"`
}


// создает аккаунт, несмотря на a.ID
func (a *Account) Create () error {

	pool := base.GetPool()

	// создаем объект с необходимыми полями
	res, err := pool.Exec("insert into accounts (name) VALUES (?)", a.Name);
	if err != nil {
		return err
	}

	// получаем id вставки, при удачном создании
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// присваиваем нужный ID аккаунту
	a.ID = uint(id)

	// читаем из БД сохраненный объект и возвращаем ошибку
	return a.Get()
}

// осуществляет поиск по a.ID
func (a *Account) Get () error {

	return base.GetPool().QueryRow("select id,name,created_at,updated_at,deleted_at from accounts where id = ? AND deleted_at = 0;", a.ID).
		Scan(&a.ID, &a.Name, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt);

}

// сохраняет все необходимые поля
func (a *Account) save () error {
	res, err := base.GetPool().Exec("update accounts set name = ?, deleted_at = ? where id = ? ", a.Name, a.DeletedAt, a.ID);
	if err != nil {
		return err
	}

	// получаем id вставки, при удачном создании
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	// присваиваем нужный ID аккаунту
	a.ID = uint(id)

	return a.Get()
}
func (a *Account) delete () error {
	return nil
}

// ### пробуем упростить

func (a *Account) getID() uint {
	return a.ID
}



























