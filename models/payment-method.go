package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/utils"
)

// "payment_cash", "payment_yandex"
type PaymentMethod interface {
	Entity

	GetType() string
	GetCode() string

	// Функция запускающая процесс создания платежа под Order (Заказ)
	CreatePaymentByOrder(order Order) (*Payment, error)
	GetWebSiteId() uint
}

func (account Account) GetPaymentMethods() ([]PaymentMethod, error) {

	// Находим все необходимые методы
	var paymentCashes []PaymentCash
	paymentCashesEntity, _, err := PaymentCash{}.getList(account.Id, "id")
	if err != nil {
	   return nil, err
	}
	for i := range(paymentCashesEntity) {
		_p, ok :=  paymentCashesEntity[i].(*PaymentCash)
		if !ok { return nil, utils.Error{Message: "Ошибка при получении списка PaymentCash"} }
		paymentCashes = append(paymentCashes,*_p)
	}

	var paymentYandexes []PaymentYandex
	paymentYandexEntity, _, err := PaymentYandex{}.getList(account.Id, "id")
	if err != nil {
		return nil, err
	}
	for i := range(paymentYandexEntity) {
		_p, ok :=  paymentYandexEntity[i].(*PaymentYandex)
		if !ok { return nil, utils.Error{Message: "Ошибка при получении списка PaymentCash"} }

		paymentYandexes = append(paymentYandexes,*_p)
	}

	methods := make([]PaymentMethod, len(paymentCashes)+len(paymentYandexes))
	for i,_ := range paymentCashes {
		methods[i] = &paymentCashes[i]
	}
	for i,_ := range paymentYandexes {
		methods[i+len(paymentCashes)] = &paymentYandexes[i]
	}

	return methods, nil
}

func (account Account) GetPaymentMethodByCode(code string, methodId uint) (PaymentMethod, error) {

	if code == "" || methodId < 1{
		return nil, utils.Error{Message: "Не верно указаны данные типа оплаты"}
	}
	// 1. Получаем все варианты доставки (обычно их мало). Можно через switch, но лень потом исправлять баг с новыми типом доставки
	methods, err := account.GetPaymentMethods()
	if err != nil { return nil, err}


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

func (account Account) GetPaymentMethodByType(codeType string, methodId uint) (PaymentMethod, error) {

	fmt.Println("GetPaymentMethodByCode")
	fmt.Println(codeType, methodId)
	
	if codeType == "" || methodId < 1{
		return nil, utils.Error{Message: "Не верно указаны данные типа оплаты"}
	}
	// 1. Получаем все варианты доставки (обычно их мало). Можно через switch, но лень потом исправлять баг с новыми типом доставки
	methods, err := account.GetPaymentMethods()
	if err != nil { return nil, err}


	// Ищем наш вариант доставки
	var method PaymentMethod
	for _,v := range methods {
		if v.GetType() == codeType && v.GetId() == methodId {
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