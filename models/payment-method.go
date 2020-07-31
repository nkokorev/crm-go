package models

import "github.com/nkokorev/crm-go/utils"

// "payment_cash", "payment_yandex"
type PaymentMethod interface {
	Entity

	GetCode() string
	// создает платеж
	CreatePayment(order Order) (*Payment, error)
	GetWebSiteId() uint
}

func (account Account) GetPaymentMethods() []PaymentMethod {

	// Находим все необходимые методы
	var paymentCashes []PaymentCash
	if err := db.Model(&DeliveryRussianPost{}).Preload("WebSite").
		Find(&paymentCashes, "account_id = ?", account.Id).Error; err != nil {
		return nil
	}

	var paymentYandexes []PaymentYandex
	if err := db.Model(&DeliveryCourier{}).Preload("WebSite").
		Find(&paymentYandexes, "account_id = ?", account.Id).Error; err != nil {
		return nil
	}

	methods := make([]PaymentMethod, len(paymentCashes)+len(paymentYandexes))
	for i,_ := range paymentCashes {
		methods[i] = &paymentCashes[i]
	}
	for i,_ := range paymentYandexes {
		methods[i+len(paymentCashes)] = &paymentYandexes[i]
	}

	return methods
}

func (account Account) GetPaymentMethod(code string, methodId uint) (PaymentMethod, error){

	// 1. Получаем все варианты доставки (обычно их мало). Можно через switch, но лень потом исправлять баг с новыми типом доставки
	methods := account.GetPaymentMethods()


	// Ищем наш вариант доставки
	var method PaymentMethod
	for _,v := range methods {
		if v.GetCode() == code && v.GetId() == methodId {
			method = v
			break
		}
	}

	// Проверяем, удалось ли найти выбранный вариант оплаты
	if method == nil {
		return nil, utils.Error{Message: "Не верно указан тип оплаты"}
	}

	return method, nil
}