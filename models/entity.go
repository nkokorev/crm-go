package models

import (
	"errors"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"reflect"
)

type Entity interface {

	getId() uint
	setId(id uint)
	GetAccountId() uint
	setAccountId(id uint)

	// CRUD model
	create() (Entity, error)
	get (id uint) (Entity, error)
	load () error
	getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error)
	update(input map[string]interface{}) error
	delete() error

	// AppendAssociationMethod(options Entity)
	systemEntity() bool

}

func Get(v Entity) error {

	// id := v.getID()

	r := reflect.TypeOf(v)

	fmt.Println(r, r.Elem(), r.NumMethod())

	println("We are use v Entity GET function")

	return nil
}

func (account Account) CreateEntity(input Entity) (Entity, error) {
	input.setAccountId(account.ID)
	return input.create()
}

func (account Account) GetEntity(model Entity, id uint) (Entity, error) {

	entity, err := model.get(id)
	if err != nil {
		return nil, err
	}

	if entity.GetAccountId() != account.ID {
		if !entity.systemEntity() {
			return nil, errors.New("Модель принадлежит другому аккаунту")
		}

	}

	return entity, nil
}

func (account Account) LoadEntity(entity Entity, primaryKey ...uint) error {

	if len(primaryKey) > 0 {
		entity.setId(primaryKey[0])
	}
	
	// Загружаем по ссылке
	err := entity.load()
	if err != nil {
		return err
	}

	// Проверяем принадлежность к аккаунту
	if entity.GetAccountId() != account.ID {
		if !entity.systemEntity() {
			return errors.New("Модель принадлежит другому аккаунту")
		}

	}

	return nil
}

func (account Account) GetPaginationListEntity(model Entity, offset, limit int, order string, search string) ([]Entity, uint, error) {
	return model.getPaginationList(account.ID, offset, limit, order, search)
}

func (account Account) UpdateEntity(entity Entity, input map[string]interface{}) error {
	if entity.GetAccountId() != account.ID {
		return utils.Error{Message: "Объект принадлежит другому аккаунту"}
	}

	return entity.update(input)
}

func (account Account) DeleteEntity(entity Entity) error {
	if entity.GetAccountId() != account.ID {
		return utils.Error{Message: "Объект принадлежит другому аккаунту"}
	}
	
	return entity.delete()
}
