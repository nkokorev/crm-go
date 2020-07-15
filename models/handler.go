package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"reflect"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
type Handler struct {
	TargetName string // Имя функции, которую вызывают локально
}

func (handle Handler) Handle(e event.Event) error {

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
	if err := target(e); err != nil {
		e.Abort(true)
		return err
	}

	return nil
}

// #############   Event Handlers   #############
func (handler Handler) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, обытие: %v данные: %v\n",e.Name(), e.Data())
	// fmt.Println("Observer: ", handler) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}
func (handler Handler) WebHookCall(e event.Event) error {
	fmt.Printf("Вызов вебхука, событие: %v Данные: %v\n", e.Name(), e.Data())
	// fmt.Println("Observer: ", handler) // контекст вебхука, какой именно и т.д.
	// e.Set("result", "OK")
	return nil
}
// #############   END Of Event Handlers   #############


