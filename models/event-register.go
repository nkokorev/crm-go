package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"reflect"
)

// Возвращает список всех возможных обработчиков
func GetSystemHandlerList () []map[string]string{
	return []map[string]string {
		{"name": "EmailQueueRun", "description": "Запуск автоматической серии email писем с указанным ID."},
		{"name":"WebHookCall","description":"Вызов WebHook с указанным ID."},
	}
}

// Нужна функция ReloadEventHandler(e)
func (Observer) Registration() error {

	eventListeners, err := Observer{}.getFullList()
	if err != nil {
		return utils.Error{Message: "Не удалось загрузить EventHandlers!"}
	}

	for _,v := range eventListeners {
		event.On(v.EventName, v, v.Priority)
	}
	
	return nil
}

func (Observer) ReloadEventHandler() error {
	em := event.DefaultEM
	em.Clear()

	return Observer{}.Registration()
}



// функция обработчик для каждого события
func (eh Observer) Handle(e event.Event) error {

	// 1. Получаем метод обработки по имени Target
	m := reflect.ValueOf(eh).MethodByName(eh.TargetName)
	if m.IsNil() {
		e.Abort(true)
		return utils.Error{Message: fmt.Sprintf("Observer Handle is nill: %v", eh.TargetName)}
	}

	// 2. Преобразуем метод, чтобы его можно было вызвать от объекта Event
	target, ok := m.Interface().(func(e event.Event) error)
	if !ok {
		e.Abort(true)
		return utils.Error{Message: fmt.Sprintf("Observer mCallable !ok: %v", eh.TargetName)}
	}

	// 3. Вызываем Target-метод с объектом Event
	if err := target(e); err != nil {
		e.Abort(true)
		return err
	}

	return nil
}

// ########################################################


// #############   Event Handlers   #############
func (eh Observer) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, данные: %v\n", e.Data())
	// fmt.Println("Observer: ", eh) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}
func (eh Observer) WebHookCall(e event.Event) error {
	fmt.Printf("Вызов вебхука, данные: %v\n", e.Data())
	// fmt.Println("Observer: ", eh) // контекст вебхука, какой именно и т.д.
	// e.Set("result", "OK")
	return nil
}
// #############   END Of Event Handlers   #############
