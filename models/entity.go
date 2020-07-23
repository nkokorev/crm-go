package models

import (
	"errors"
	"fmt"
	"reflect"
)

type Entity interface {

	GetID() uint
	setID(id uint)
	GetAccountID() uint
	setAccountID(id uint)

	// CRUD model
	create() (Entity, error)
	get (id uint) (Entity, error)
	load () error
	getList(accountID uint, order string) ([]Entity, uint, error)
	getPaginationList(accountID uint, offset, limit int, sortBy, search string) ([]Entity, uint, error)
	update(input map[string]interface{}) error
	delete() error

	// AppendAssociationMethod(options Entity)
	SystemEntity() bool

}

func Get(v Entity) error {

	// id := v.getID()

	r := reflect.TypeOf(v)

	fmt.Println(r, r.Elem(), r.NumMethod())

	println("We are use v Entity GET function")

	return nil
}

func (account Account) CreateEntity(input Entity) (Entity, error) {
	input.setAccountID(account.ID)
	return input.create()
}

func (account Account) GetEntity(model Entity, id uint) (Entity, error) {

	entity, err := model.get(id)
	if err != nil {
		return nil, err
	}

	// Тут надо бы показать, что она системная
	if entity.GetAccountID() != account.ID && !entity.SystemEntity() {
		return nil, errors.New("Модель принадлежит другому аккаунту")
	}

	return entity, nil
}

func (account Account) LoadEntity(entity Entity, primaryKey ...uint) error {

	if len(primaryKey) > 0 {
		entity.setID(primaryKey[0])
	}
	
	// Загружаем по ссылке
	err := entity.load()
	if err != nil {
		return err
	}

	// Проверяем принадлежность к аккаунту
	if entity.GetAccountID() != account.ID && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return nil
}

func (account Account) GetListEntity(model Entity, order string) ([]Entity, uint, error) {
	return model.getList(account.ID, order)
}

func (account Account) GetPaginationListEntity(model Entity, offset, limit int, order string, search string) ([]Entity, uint, error) {
	return model.getPaginationList(account.ID, offset, limit, order, search)
}

func (account Account) UpdateEntity(entity Entity, input map[string]interface{}) error {
	if entity.GetAccountID() != account.ID && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return entity.update(input)
}

func (account Account) DeleteEntity(entity Entity) error {
	if entity.GetAccountID() != account.ID && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}
	
	return entity.delete()
}
