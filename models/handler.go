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

	handle.TargetName = handle.Handler.Name
	// 1. Получаем метод обработки по имени Target
	m := reflect.ValueOf(handle).MethodByName(handle.TargetName)
	if m.IsNil() {
		e.Abort(true)
		log.Println("Observer Handle is nill")
		return utils.Error{Message: fmt.Sprintf("Observer Handle is nill: %v", handle.TargetName)}
	}

	// 2. Преобразуем метод, чтобы его можно было вызвать от объекта Event
	target, ok := m.Interface().(func(e event.Event) error)
	if !ok {
		e.Abort(true)
		log.Println("Observer mCallable !ok")
		return utils.Error{Message: fmt.Sprintf("Observer mCallable !ok: %v", handle.TargetName)}
	}

	// 3. Вызываем Target-метод с объектом Event
	accountStr := e.Get("accountId")
	accountId, ok :=  accountStr.(uint)
	if !ok {
		log.Println("Not ok!")
	}

	// Выполняем, только если совпадает контекст аккаунта
	if handle.AccountID == accountId {
		if err := target(e); err != nil {
			e.Abort(true)
			return err
		}
	}


	return nil
}

// #############   Event Handlers   #############
func (handler EventListener) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, обытие: %v данные: %v\n",e.Name(), e.Data())
	// fmt.Println("Observer: ", handler) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}
func (handler EventListener) WebHookCall(e event.Event) error {
	fmt.Printf("Вызов вебхука, событие: %v Данные: %v\n", e.Name(), e.Data())
	// fmt.Println("Observer: ", handler) // контекст вебхука, какой именно и т.д.
	// e.Set("result", "OK")
	return nil
}
// #############   END Of Event Handlers   #############


