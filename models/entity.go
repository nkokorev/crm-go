package models

import (
	"errors"
	"fmt"
	"reflect"
)

type Entity interface {

	getId() uint
	setId(id uint)
	GetAccountId() uint

	setAccountId(id uint)

	getEntityName() string // Article, EmailTemplate

	// CRUD model
	create() (Entity, error)
	// update(input map[string]interface{}) (*Entity, error)
	// get (id uint) (interface{}, error)
	get (id uint) (Entity, error)
	GetPaginationList(accountId uint, offset, limit int, order string, search *string) ([]Entity, error)
	load () error
	//update(input map[string]interface{}) error
	//delete() error

	// GetNullArray() []interface{}


}

func Get(v Entity) error {

	// id := v.getID()

	r := reflect.TypeOf(v)

	fmt.Println(r, r.Elem(), r.NumMethod())

	println("We are use v Entity GET function")

	return nil
}

func getEventNameCreated(entity Entity) string {
	return "Create" + entity.getEntityName() + "Created"
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
		return nil, errors.New("Модель принадлежит другому аккаунту")
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
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return nil
}

func (account Account) GetPaginationListEntity(model Entity, offset, limit int, order string, search *string) ([]Entity, error) {
	return model.GetPaginationList(account.ID, offset, limit, order, search)
}

/*func (account Account) UpdateEntity(entity Entity, input map[string]interface{}) error {

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
		return errors.New("Модель принадлежит другому аккаунту")
	}

	return nil
}*/

/*func (account Account) DeleteEntity(id uint, model Entity) error {

	entity, err := model.get(id)
	if err != nil {
		return nil, err
	}

	if entity.GetAccountId() != account.ID {
		return nil, errors.New("Модель принадлежит другому аккаунту")
	}

	return entity, nil
}*/

/*func (account Account) ApiKeyGet(id uint) (*ApiKey, error) {
	apiKey, err := ApiKey{}.get(id)
	if err != nil {
		return nil, err
	}

	if apiKey.AccountID != account.ID {
		return nil, errors.New("ApiKey не принадлежит аккаунту")
	}

	return apiKey, nil
}

func (account Account) ApiKeyGetByToken(token string) (*ApiKey, error) {
	apiKey, err := GetApiKeyByToken(token)
	if err != nil {
		return nil, err
	}

	if apiKey.AccountID != account.ID {
		return nil, errors.New("ApiKey не принадлежит аккаунту")
	}

	return apiKey, nil
}

func (account Account) ApiKeysList() ([]ApiKey, error) {

	keyList, err := ApiKey{}.getList(account.ID)
	if err != nil {
		return nil, errors.New("Не удалось получить список")
	}

	return keyList, nil
}

func (account Account) ApiKeyUpdate(id uint, input interface{}) (*ApiKey, error) {
	apiKey, err := account.ApiKeyGet(id)
	if err != nil {
		return nil, err
	}

	if account.ID != apiKey.AccountID {
		return nil, utils.Error{Message: "Ключ принадлежит другому аккаунту"}
	}

	err = apiKey.update(input)

	return apiKey, err

}

func (account Account) ApiKeyDelete(id uint) error {

	apiKey, err := account.ApiKeyGet(id)
	if err != nil {
		return err
	}

	return apiKey.delete()
}*/