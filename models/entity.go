package models

import (
	"fmt"
	"reflect"
)

type Entity interface {

	getID() uint

	// CRUD model
	create() error
	get () error
	save() error
	delete() error


}

func Get(v Entity) error {

	// id := v.getID()

	r := reflect.TypeOf(v)

	fmt.Println(r, r.Elem(), r.NumMethod())

	println("We are use v Entity GET function")

	return nil
}
