package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"reflect"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
// Event - не загружен!
func (eventListener *EventListener) Handle(e Event) error {

	TargetName := eventListener.Handler.Code

	if TargetName == "" {
		log.Printf("EventListener Handle Name is nill: %v\n", TargetName)
		return utils.Error{Message: fmt.Sprintf("EventListener Handle Name is nill %v\n", TargetName)}
	}

	// Проверяем принадлежность
	accountStr := e.Get("account_id")
	accountId, ok :=  accountStr.(uint)
	if !ok {
		log.Println("Event Handler: accountStr.(uint):: ", eventListener.Name)
		return nil
	}
	if eventListener.AccountId != accountId {
		// log.Println("Event listener event Id: ", eventListener.EventId, " Handle ID: ", eventListener.HandlerId)
		// log.Println("Event Handler: eventListener.AccountId != accountId:: ", eventListener.Name, " eventListener.AccountId: ", eventListener.AccountId, " accountID: ", accountId)
		return nil
	}

	// 1. Получаем метод обработки по имени Target
	v := reflect.ValueOf(eventListener)
	if v.IsNil() {
		log.Println("Observer ValueOf handle is nill 1")
		return utils.Error{Message: fmt.Sprintf("Observer Handle is nill: %v", TargetName)}
	}

	// Тест на существование
	el := v.Type()
	_, ok = el.MethodByName(TargetName)
	if !ok {
		log.Println("Observer ValueOf handle is nill 2")
		return utils.Error{Message: fmt.Sprintf("Observer Handle Type is nill: %v", TargetName)}
	}

	// Получаем метод
	m := v.MethodByName(TargetName)
	if m.IsNil() {
		log.Println("Observer Handle is nill 3")
		return utils.Error{Message: fmt.Sprintf("Observer MethodByName not found: %v", TargetName)}
	}

	// 2. Преобразуем метод, чтобы его можно было вызвать от объекта Event
	// target, ok := m.Interface().(func(e Event) error)
	target, ok := m.Interface().(func(e Event) error)
	if !ok {
		log.Println("Observer mCallable !ok 4")
		return utils.Error{Message: fmt.Sprintf("Observer mCallable !ok: %v", TargetName)}
	}

	// 3. Вызываем Target-метод с объектом Event
	if err := target(e); err != nil {
		log.Println(err)
		return err
	}

	// log.Println("target успешно выполнен!")

	return nil
}

func (eventListener *EventListener) OLDHandle(e Event) error {

	TargetName := eventListener.Handler.Code

	if TargetName == "" {
		// log.Printf("EventListener Handle Name is nill: %v\n", TargetName)
		return utils.Error{Message: fmt.Sprintf("EventListener Handle Name is nill %v\n", TargetName)}
	}

	// 1. Получаем метод обработки по имени Target
	v := reflect.ValueOf(eventListener)
	if v.IsNil() {
		// log.Println("Observer ValueOf handle is nill 1")
		return utils.Error{Message: fmt.Sprintf("Observer Handle is nill: %v", TargetName)}
	}

	// Тест на существование
	el := v.Type()
	_, ok := el.MethodByName(TargetName)
	if !ok {
		// log.Println("Observer ValueOf handle is nill 2")
		return utils.Error{Message: fmt.Sprintf("Observer Handle Type is nill: %v", TargetName)}
	}

	// Получаем метод
	m := v.MethodByName(TargetName)
	if m.IsNil() {
		// log.Println("Observer Handle is nill 3")
		return utils.Error{Message: fmt.Sprintf("Observer MethodByName not found: %v", TargetName)}
	}

	// 2. Преобразуем метод, чтобы его можно было вызвать от объекта Event
	// target, ok := m.Interface().(func(e Event) error)
	target, ok := m.Interface().(func(e Event) error)
	if !ok {
		log.Println(m.Interface())
		log.Println("Observer mCallable !ok 4")
		return utils.Error{Message: fmt.Sprintf("Observer mCallable !ok: %v", TargetName)}
	}

	// 3. Вызываем Target-метод с объектом Event
	accountStr := e.Get("account_id")
	accountId, ok :=  accountStr.(uint)
	if !ok || eventListener.AccountId != accountId {
		return nil
	}

	if err := target(e); err != nil {
		// что-то пошло не так...
		// log.Println(err)
		// return err
	}

	// fmt.Println("target успешно выполнен!")

	return nil
}

// #############   Event Handlers   #############
func (eventListener EventListener) EmailNotificationRun(e Event) error {

	fmt.Printf("#### Запуск уведомления письмом, обытие: %v данные: %v\n",e.GetName(), e.Data())
	// fmt.Println("Observer entity id: ", handler.EntityId) // контекст серии писем, какой именно и т.д.

	accountStr := e.Get("account_id")
	accountId, ok :=  accountStr.(uint)
	if !ok {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить Email Notification id = %v, не найден accountId.", eventListener.EntityId)}
	}

	account, err := GetAccount(accountId)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить Email Notification id = %v, не найден account by id: %v.", eventListener.EntityId, accountId)}
	}

	var en EmailNotification
	if err := account.LoadEntity(&en, eventListener.EntityId,nil); err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить Email Notification id = %v, уведомление не найдено!", eventListener.EntityId)}
	}

	if !en.IsActive() {
		return utils.Error{Message: fmt.Sprintf("Уведомление id = %v не может быть отправлено т.к. находится в статусе - 'Отключено'", eventListener.EntityId)}
	}

	// Загружаем данные в тело события по их id (accountId => Account)
	eventListener.uploadEntitiesData(&e)

	return en.Execute(e.Data())
}

func (eventListener EventListener) EmailQueueRun(e Event) error {
	fmt.Printf("Запуск серии писем, обытие: %v данные: %v\n",e.GetName(), e.Data())

	accountStr := e.Get("account_id")
	accountId, ok :=  accountStr.(uint)
	if !ok {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить EmailQueue id = %v, не найден accountId.", eventListener.EntityId)}
	}

	account, err := GetAccount(accountId)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить EmailQueue id = %v, не найден account by id: %v.", eventListener.EntityId, accountId)}
	}

	var emailQueue EmailQueue
	if err := account.LoadEntity(&emailQueue, eventListener.EntityId,nil); err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить EmailQueue id = %v, не загружается объект: %v.", eventListener.EntityId, err)}
	}

	// Загружаем данные в теле
	eventListener.uploadEntitiesData(&e)

	// Получаем userId
	if userId, ok := e.Get("user_id").(uint); ok {
		// Проверяем, что он аккаунте
		if account.ExistAccountUser(userId) {
			if err := emailQueue.AppendUser(userId); err != nil {
				return err
			}
		}
	}
	


	// fmt.Println("Observer: ", handler) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}

func (eventListener EventListener) WebHookCall(e Event) error {

	// fmt.Printf("Вызов вебхука, событие: %v Данные: %v, EventId %v\n", e.Name, e.Data(), eventListener.EventId)
	accountStr := e.Get("account_id")
	accountId, ok :=  accountStr.(uint)
	if !ok {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить WebHook id = %v, не найден accountId.", eventListener.EntityId)}
	}

	account, err := GetAccount(accountId)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить WebHook id = %v, не найден account by id: %v.", eventListener.EntityId, accountId)}
	}

	var webHook WebHook
	if err := account.LoadEntity(&webHook, eventListener.EntityId,nil); err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить WebHook id = %v, не загружается webHook.", eventListener.EntityId)}
	}

	// Загружаем данные в теле
	eventListener.uploadEntitiesData(&e)

	// return webHook.Execute(e)
	return webHook.Execute(e.Data())
}

// #############   END Of Event Handlers   #############


// Загружает данные по переданным id
func (eventListener EventListener) uploadEntitiesData(event *Event) {
	// data :=(*e).Data()
	e := *event

	// 1. Get Account
	accountId, ok :=  e.Get("account_id").(uint)
	if !ok { return }

	account, err := GetAccount(accountId)
	if err != nil || account == nil {
		return
	}
	e.Add("Account", account.GetDepersonalizedData())

	if userId, ok := e.Get("user_id").(uint); ok {
		user, err := account.GetUser(userId)
		if err == nil {
			e.Add("user_id", userId)
			e.Add("User", user.GetDepersonalizedData())
		}
	}

	if productId, ok := e.Get("product_id").(uint); ok {
		var product Product
		err := account.LoadEntity(&product, productId,nil)
	   if err == nil {
		   e.Add("product_id", productId)
		   e.Add("Product", product)
	   }
	}

	if productCardId, ok := e.Get("product_card_id").(uint); ok {
		var productCard ProductCard
		err := account.LoadEntity(&productCard, productCardId,nil)
		if err == nil {
			e.Add("product_card_id", productCardId)
			e.Add("ProductCard", productCard)
		}
	}

	if orderId, ok := e.Get("order_id").(uint); ok {
		var order Order
		err := account.LoadEntity(&order, orderId,nil)
		if err == nil {
			e.Add("order_id", orderId)
			e.Add("Order", order)

			// Добавляем заказчика
			if order.CustomerId != nil {
				customer, err := account.GetUser(*order.CustomerId)
				if err == nil {
					e.Add("customer_id", order.CustomerId)
					e.Add("Customer", customer)
				}
			}

			if *order.ManagerId > 0 {
				manager, err := account.GetUser(*order.ManagerId)
				if err == nil {
					e.Add("manager_id", order.ManagerId)
					e.Add("Manager", manager)
				}
			}
		}
	}

	if deliveryOrderId, ok := e.Get("delivery_order_id").(uint); ok {
		var deliveryOrder DeliveryOrder
		err := account.LoadEntity(&deliveryOrder, deliveryOrderId,nil)
		if err == nil {
			e.Add("delivery_order_id", deliveryOrderId)
			e.Add("DeliveryOrder", deliveryOrder)
		}
	}

	if paymentId, ok := e.Get("payment_id").(uint); ok {
		var payment DeliveryOrder
		err := account.LoadEntity(&payment, paymentId,nil)
		if err == nil {
			e.Add("payment_id", paymentId)
			e.Add("Payment", payment)
		}
	}
}