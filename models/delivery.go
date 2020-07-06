package models

type Delivery interface {
	Entity
	GetName() string
	// AppendAssociationMethod(options Entity)
}
