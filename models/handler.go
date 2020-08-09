package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"reflect"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!

func (handle *EventListener) Handle(e event.Event) error {

	TargetName := handle.Handler.Code

	if TargetName == "" {
		// log.Printf("EventListener Handle Name is nill: %v\n", TargetName)
		return utils.Error{Message: fmt.Sprintf("EventListener Handle Name is nill %v\n", TargetName)}
	}

	// 1. Получаем метод обработки по имени Target
	v := reflect.ValueOf(handle)
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
	target, ok := m.Interface().(func(e event.Event) error)
	if !ok {
		log.Println("Observer mCallable !ok 4")
		return utils.Error{Message: fmt.Sprintf("Observer mCallable !ok: %v", TargetName)}
	}

	// 3. Вызываем Target-метод с объектом Event
	accountStr := e.Get("accountId")
	accountId, ok :=  accountStr.(uint)
	if !ok || handle.AccountId != accountId {
		return nil
	}

	if err := target(e); err != nil {
		log.Println(err)
	}

	// fmt.Println("target успешно выполнен!")

	return nil
}

// #############   Event Handlers   #############


func (handler EventListener) EmailNotificationRun(e event.Event) error {

	fmt.Printf("#### Запуск уведомления письмом, обытие: %v данные: %v\n",e.Name(), e.Data())
	// fmt.Println("Observer entity id: ", handler.EntityId) // контекст серии писем, какой именно и т.д.

	accountStr := e.Get("accountId")
	accountId, ok :=  accountStr.(uint)
	if !ok {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить Email Notification id = %v, не найден accountId.", handler.EntityId)}
	}

	account, err := GetAccount(accountId)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить Email Notification id = %v, не найден account by id: %v.", handler.EntityId, accountId)}
	}

	var en EmailNotification
	if err := account.LoadEntity(&en, handler.EntityId); err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить Email Notification id = %v, уведомление не найдено!", handler.EntityId)}
	}

	if !en.Enabled {
		return utils.Error{Message: fmt.Sprintf("Уведомление id = %v не может быть отправлено т.к. находится в статусе - 'Отключено'", handler.EntityId)}
	}

	// Загружаем данные в теле
	handler.uploadEntitiesData(&e)

	return en.Execute(e.Data())
}

func (handler EventListener) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, обытие: %v данные: %v\n",e.Name(), e.Data())

	accountStr := e.Get("accountId")
	accountId, ok :=  accountStr.(uint)
	if !ok {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить EmailQueue id = %v, не найден accountId.", handler.EntityId)}
	}

	account, err := GetAccount(accountId)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить EmailQueue id = %v, не найден account by id: %v.", handler.EntityId, accountId)}
	}

	var emailQueue EmailQueue
	if err := account.LoadEntity(&emailQueue, handler.EntityId); err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить EmailQueue id = %v, не загружается объект: %v.", handler.EntityId, err)}
	}

	// Загружаем данные в теле
	handler.uploadEntitiesData(&e)

	// Получаем userId
	if userId, ok := e.Get("userId").(uint); ok {
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

func (handler EventListener) WebHookCall(e event.Event) error {

	// fmt.Printf("Вызов вебхука, событие: %v Данные: %v, entityId %v\n", e.Name(), e.Data(), handler.EntityId)

	accountStr := e.Get("accountId")
	accountId, ok :=  accountStr.(uint)
	if !ok {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить WebHook id = %v, не найден accountId.", handler.EntityId)}
	}

	account, err := GetAccount(accountId)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить WebHook id = %v, не найден account by id: %v.", handler.EntityId, accountId)}
	}

	var webHook WebHook
	if err := account.LoadEntity(&webHook, handler.EntityId); err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить WebHook id = %v, не загружается webHook.", handler.EntityId)}
	}

	// Загружаем данные в теле
	handler.uploadEntitiesData(&e)

	// return en.Execute(e.Data())
	
	// return webHook.Execute(e)
	return webHook.Execute(e.Data())
}
// #############   END Of Event Handlers   #############


// Загружает данные по переданным id
func (handle EventListener) uploadEntitiesData(event *event.Event) {
	// data :=(*e).Data()
	e := *event

	// 1. Get Account
	accountId, ok :=  e.Get("accountId").(uint);
	if !ok { return }

	account, err := GetAccount(accountId)
	if err != nil || account == nil {
		return
	}
	e.Add("Account", account.GetDepersonalizedData())

	if userId, ok := e.Get("userId").(uint); ok {
		user, err := account.GetUser(userId)
		if err == nil {
			e.Add("userId", userId)
			e.Add("User", user.GetDepersonalizedData())
		}
	}

	if productId, ok := e.Get("productId").(uint); ok {
	   product, err := account.GetProduct(productId)
	   if err == nil {
		   e.Add("productId", productId)
		   e.Add("Product", *product)
	   }
	}

	if productCardId, ok := e.Get("productCardId").(uint); ok {
		productCard, err := account.GetProductCard(productCardId)
		if err == nil {
			e.Add("productCardId", productCardId)
			e.Add("ProductCard", *productCard)
		}
	}

	if orderId, ok := e.Get("orderId").(uint); ok {
		var order Order
		err := account.LoadEntity(&order, orderId)
		if err == nil {
			e.Add("orderId", orderId)
			e.Add("Order", order)

			// Добавляем заказчика
			if order.CustomerId > 0 {
				customer, err := account.GetUser(order.CustomerId)
				if err == nil {
					e.Add("customerId", order.CustomerId)
					e.Add("Customer", customer)
				}
			}

			if order.ManagerId > 0 {
				manager, err := account.GetUser(order.ManagerId)
				if err == nil {
					e.Add("managerId", order.ManagerId)
					e.Add("Manager", manager)
				}
			}
		}
	}

	if deliveryOrderId, ok := e.Get("deliveryOrderId").(uint); ok {
		var deliveryOrder DeliveryOrder
		err := account.LoadEntity(&deliveryOrder, deliveryOrderId)
		if err == nil {
			e.Add("deliveryOrderId", deliveryOrderId)
			e.Add("DeliveryOrder", deliveryOrder)
		}
	}

	if paymentId, ok := e.Get("paymentId").(uint); ok {
		var payment DeliveryOrder
		err := account.LoadEntity(&payment, paymentId)
		if err == nil {
			e.Add("paymentId", paymentId)
			e.Add("Payment", payment)
		}
	}
	
}
