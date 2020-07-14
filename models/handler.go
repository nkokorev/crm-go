package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"reflect"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
type Handler struct {
	TargetName string
}

func (handle Handler) Handle(e event.Event) error {

	// 1. Получаем метод обработки по имени Target
	m := reflect.ValueOf(handle).MethodByName(handle.TargetName)
	if m.IsNil() {
		e.Abort(true)
		return utils.Error{Message: fmt.Sprintf("Observer Handle is nill: %v", handle.TargetName)}
	}

	// 2. Преобразуем метод, чтобы его можно было вызвать от объекта Event
	target, ok := m.Interface().(func(e event.Event) error)
	if !ok {
		e.Abort(true)
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
func (observer Handler) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, данные: %v\n", e.Data())
	// fmt.Println("Observer: ", observer) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}
func (observer Handler) WebHookCall(e event.Event) error {
	fmt.Printf("Вызов вебхука, данные: %v\n", e.Data())
	// fmt.Println("Observer: ", observer) // контекст вебхука, какой именно и т.д.
	// e.Set("result", "OK")
	return nil
}
// #############   END Of Event Handlers   #############


