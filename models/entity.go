package models

import (
	"errors"
	"fmt"
	"reflect"
)

type Entity interface {

	GetId() uint
	setId(id uint)
	setPublicId(id uint)
	GetAccountId() uint
	setAccountId(id uint)

	// CRUD model
	create() (Entity, error)
	get (id uint) (Entity, error)
	load () error
	loadByPublicId () error
	getList(accountId uint, order string) ([]Entity, uint, error)
	getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error)
	update(input map[string]interface{}) error
	delete() error

	// AppendAssociationMethod(options Entity)
	SystemEntity() bool

}

func Get(v Entity) error {

	// id := v.getId()

	r := reflect.TypeOf(v)

	fmt.Println(r, r.Elem(), r.NumMethod())

	println("We are use v Entity GET function")

	return nil
}

func (account Account) CreateEntity(input Entity) (Entity, error) {
	input.setAccountId(account.Id)
	return input.create()
}

func (account Account) GetEntity(model Entity, id uint) (Entity, error) {

	entity, err := model.get(id)
	if err != nil {
		return nil, err
	}

	// Тут надо бы показать, что она системная
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return nil, errors.New("Модель принадлежит другому аккаунту")
	}

	return entity, nil
}

func (account Account) LoadEntity(entity Entity, primaryKey uint) error {

	// На всякий случай
	entity.setId(primaryKey)
	
	// Загружаем по ссылке
	err := entity.load()
	if err != nil {
		return err
	}

	// Проверяем принадлежность к аккаунту
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return nil
}

func (account Account) LoadEntityByPublicId(entity Entity, publicId uint) error {

	// На всякий случай
	entity.setPublicId(publicId)

	// Загружаем по ссылке
	err := entity.loadByPublicId()
	if err != nil {
		return err
	}

	// Проверяем принадлежность к аккаунту
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return nil
}

func (account Account) GetListEntity(model Entity, order string) ([]Entity, uint, error) {
	return model.getList(account.Id, order)
}

func (account Account) GetPaginationListEntity(model Entity, offset, limit int, order string, search string) ([]Entity, uint, error) {
	return model.getPaginationList(account.Id, offset, limit, order, search)
}

func (account Account) UpdateEntity(entity Entity, input map[string]interface{}) error {
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return entity.update(input)
}

func (account Account) DeleteEntity(entity Entity) error {
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}
	
	return entity.delete()
}
