package models

type PaymentMethod interface {
	Entity

	// создает платеж
	CreatePayment(order Order) (*Payment, error)
}
