package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"reflect"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
/*type EventListener struct {
	TargetName string // Имя функции, которую вызывают локально
}*/

func (handle EventListener) Handle(e event.Event) error {

	// utils.TimeTrack(time.Now())

	TargetName := handle.Handler.Code

	if TargetName == "" {
		log.Printf("EventListener Handle Name is nill: %v\n", TargetName)
		return utils.Error{Message: fmt.Sprintf("EventListener Handle Name is nill %v\n", TargetName)}
	}

	// 1. Получаем метод обработки по имени Target
	m := reflect.ValueOf(handle).MethodByName(TargetName)
	if m.IsNil() {
		log.Println("Observer Handle is nill")
		return utils.Error{Message: fmt.Sprintf("Observer Handle is nill: %v", TargetName)}
	}

	// 2. Преобразуем метод, чтобы его можно было вызвать от объекта Event
	target, ok := m.Interface().(func(e event.Event) error)
	if !ok {
		log.Println("Observer mCallable !ok")
		return utils.Error{Message: fmt.Sprintf("Observer mCallable !ok: %v", TargetName)}
	}

	// 3. Вызываем Target-метод с объектом Event
	accountStr := e.Get("accountId")
	accountId, ok :=  accountStr.(uint)
	if !ok || handle.AccountID != accountId {
		return nil
	}

	return target(e)
}

// #############   Event Handlers   #############
func (handler EventListener) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, обытие: %v данные: %v\n",e.Name(), e.Data())
	// fmt.Println("Observer: ", handler) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}
func (handler EventListener) WebHookCall(e event.Event) error {
	fmt.Printf("Вызов вебхука, событие: %v Данные: %v, entityId %v\n", e.Name(), e.Data(), handler.EntityId)



	return nil
}
// #############   END Of Event Handlers   #############


