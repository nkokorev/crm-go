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
	get (id uint, preloads []string) (Entity, error)
	load (preloads []string) error
	loadByPublicId (preloads []string) error
	getList(accountId uint, order string, preloads []string) ([]Entity, int64, error)
	getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error)
	update(input map[string]interface{}, preloads []string) error
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

func (account Account) GetEntity(model Entity, id uint, preloads []string) (Entity, error) {

	entity, err := model.get(id,preloads)
	if err != nil {
		return nil, err
	}

	// Тут надо бы показать, что она системная
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return nil, errors.New("Модель принадлежит другому аккаунту")
	}

	return entity, nil
}

func (account Account) LoadEntity(entity Entity, primaryKey uint, preloads []string) error {

	// На всякий случай
	entity.setId(primaryKey)
	
	// Загружаем по ссылке
	err := entity.load(preloads)
	if err != nil {
		return err
	}

	// Проверяем принадлежность к аккаунту
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return nil
}

func (account Account) LoadEntityByPublicId(entity Entity, publicId uint, preloads []string) error {

	// На всякий случай
	entity.setPublicId(publicId)
	entity.setAccountId(account.Id)

	// Загружаем по ссылке
	err := entity.loadByPublicId(preloads)
	if err != nil {
		return err
	}

	// Проверяем принадлежность к аккаунту
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return nil
}

func (account Account) GetListEntity(model Entity, order string, preload []string) ([]Entity, int64, error) {
	return model.getList(account.Id, order,preload)
}

func (account Account) GetPaginationListEntity(model Entity, offset, limit int, order string, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {
	return model.getPaginationList(account.Id, offset, limit, order, search, filter,preloads)
}

func (account Account) UpdateEntity(entity Entity, input map[string]interface{}, preloads []string) error {
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return entity.update(input, preloads)
}

func (account Account) DeleteEntity(entity Entity) error {
	if entity.GetAccountId() != account.Id && !entity.SystemEntity() {
		return errors.New("Модель принадлежит другому аккаунту")
	}
	
	return entity.delete()
}
