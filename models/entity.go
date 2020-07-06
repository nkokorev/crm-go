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
	load () error
	//update(input map[string]interface{}) error
	//delete() error


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

/*func (account Account) GetEntity(id uint, model Entity) (Entity, error) {

	entity, err := model.get(id)
	if err != nil {
		return nil, err
	}

	if entity.GetAccountId() != account.ID {
		return nil, errors.New("Модель принадлежит другому аккаунту")
	}

	return entity, nil
}*/
// func (account Account) GetEntity(id uint, entity Entity) error {
func (account Account) GetEntity(entity Entity, keys ...uint) error {

	if len(keys) > 0 {
		entity.setId(keys[0])
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