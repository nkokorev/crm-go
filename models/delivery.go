package models

type Delivery interface {
	Entity
	GetCode() string
}
