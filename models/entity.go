package models

type Entity interface {

	getID() uint
	getAccountID() uint
	//getHashID() string
	setAccountID(uint)

	// CRUD model
	create() error
	get (string) error // read ;)
	delete() error
	update() error

}
