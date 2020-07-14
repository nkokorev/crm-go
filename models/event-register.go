package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"reflect"
)

// Возвращает список всех возможных обработчиков
func (EventHandler) GetSystemHandleList () map[string]interface{} {
	return map[string]interface{}{
		"EmailQueueRun": map[string]string{"description":"Запуск автоматической серии email писем с указанным ID."},
		"WebHookCall": map[string]string{"description":"Вызов WebHook с указанным ID."},
	}
}

// Нужна функция ReloadEventHandler(e)
func (EventHandler) RegisterEventHandler() error {

	eventListeners, err := EventHandler{}.getFullList()
	if err != nil {
		return utils.Error{Message: "Не удалось загрузить EventHandlers!"}
	}

	for _,v := range eventListeners {
		event.On(v.EventName, v, v.Priority)
	}
	
	return nil
}

func (EventHandler) ReloadEventHandler() error {
	em := event.DefaultEM
	em.Clear()

	return EventHandler{}.RegisterEventHandler()
}



// функция обработчик для каждого события
func (eh EventHandler) Handle(e event.Event) error {

	// 1. Получаем метод обработки по имени Target
	m := reflect.ValueOf(eh).MethodByName(eh.TargetName)
	if m.IsNil() {
		e.Abort(true)
		return utils.Error{Message: fmt.Sprintf("EventHandler Handle is nill: %v", eh.TargetName)}
	}

	// 2. Преобразуем метод, чтобы его можно было вызвать от объекта Event
	target, ok := m.Interface().(func(e event.Event) error)
	if !ok {
		e.Abort(true)
		return utils.Error{Message: fmt.Sprintf("EventHandler mCallable !ok: %v", eh.TargetName)}
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
func (eh EventHandler) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, данные: %v\n", e.Data())
	// fmt.Println("EventHandler: ", eh) // контекст серии писем, какой именно и т.д.
	// e.Set("result", "OK") // возможность записать в событие какие-то данные для других обработчиков..
	return nil
}
func (eh EventHandler) WebHookCall(e event.Event) error {
	fmt.Printf("Вызов вебхука, данные: %v\n", e.Data())
	// fmt.Println("EventHandler: ", eh) // контекст вебхука, какой именно и т.д.
	// e.Set("result", "OK")
	return nil
}
// #############   END Of Event Handlers   #############
