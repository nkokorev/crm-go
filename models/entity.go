package models

import (
	"fmt"
	"reflect"
)

type Entity interface {

	getId() uint
	getAccountId() uint

	setAccountId(id uint)

	getEntityName() string

	// CRUD model
	create() (*Entity, error)
	//get () error
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

/*func (account Account) CreateEntity(input Entity) (*Entity, error) {

	input.setAccountId(account.ID)

	entity, err := input.create()
	if err != nil {
		return nil, err
	}

	// todo: костыль вместо евента
	go account.CallWebHookIfExist(getEventNameCreated(*entity), *entity)

	return entity, nil
}*/
