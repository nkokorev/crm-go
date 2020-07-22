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

	// utils.TimeTrack(time.Now())

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
	accountStr := e.Get("accountID")
	accountID, ok :=  accountStr.(uint)
	if !ok || handle.AccountID != accountID {
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
	// fmt.Println("Observer entity id: ", handler.EntityID) // контекст серии писем, какой именно и т.д.

	accountStr := e.Get("accountID")
	accountID, ok :=  accountStr.(uint)
	if !ok {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить Email Notification id = %v, не найден accountID.", handler.EntityID)}
	}

	account, err := GetAccount(accountID)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить Email Notification id = %v, не найден account by id: %v.", handler.EntityID, accountID)}
	}

	var en EmailNotification
	if err := account.LoadEntity(&en, handler.EntityID); err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить Email Notification id = %v, уведомление не найдено!", handler.EntityID)}
	}

	if !en.Enabled {
		return utils.Error{Message: fmt.Sprintf("Уведомление id = %v не может быть отправлено т.к. находится в статусе - 'Отключено'", handler.EntityID)}
	}

	// Загружаем данные в теле
	handler.uploadEntitiesData(&e)

	return en.Execute(e.Data())
}

func (handler EventListener) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, обытие: %v данные: %v\n",e.Name(), e.Data())
	// fmt.Println("Observer: ", handler) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}

func (handler EventListener) WebHookCall(e event.Event) error {

	fmt.Printf("Вызов вебхука, событие: %v Данные: %v, entityID %v\n", e.Name(), e.Data(), handler.EntityID)

	accountStr := e.Get("accountID")
	accountID, ok :=  accountStr.(uint)
	if !ok {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить WebHook id = %v, не найден accountID.", handler.EntityID)}
	}

	account, err := GetAccount(accountID)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить WebHook id = %v, не найден account by id: %v.", handler.EntityID, accountID)}
	}

	var webHook WebHook
	if err := account.LoadEntity(&webHook, handler.EntityID); err != nil {
		return utils.Error{Message: fmt.Sprintf("Невозможно выполнить WebHook id = %v, не загружается webHook.", handler.EntityID)}
	}

	return webHook.Execute(e)
}
// #############   END Of Event Handlers   #############


// Загружает данные по переданным id
func (handle EventListener) uploadEntitiesData(event *event.Event) {
	// data :=(*e).Data()
	e := *event

	// 1. Get Account
	accountID, ok :=  e.Get("accountID").(uint);
	if !ok { return }

	account, err := GetAccount(accountID)
	if err != nil || account == nil {
		return
	}
	e.Add("Account", *account)

	if userID, ok := e.Get("userID").(uint); ok {
		user, err := account.GetUser(userID)
		if err == nil {
			e.Add("User", *user)
		}
	}

	if productID, ok := e.Get("productID").(uint); ok {
	   product, err := account.GetProduct(productID)
	   if err == nil {
		   e.Add("Product", *product)
	   }
	}

	if productCardID, ok := e.Get("productCardID").(uint); ok {
		productCard, err := account.GetProductCard(productCardID)
		if err == nil {
			e.Add("ProductCard", *productCard)
		}
	}
	
}
