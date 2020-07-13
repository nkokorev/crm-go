package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"reflect"
)

func (EventHandler) GetTargets () map[string]func(e event.Event) error {
	return map[string]func(e event.Event)error {
	"EmailQueueRun": EventHandler{}.EmailQueueRun,
	"WebHookCall": EventHandler{}.WebHookCall,
	}
}

// Нужна функция ReloadEventHandler(e)
func RegisterEventHandler() error {

	eventListeners, err := EventHandler{}.getFullList()
	if err != nil {
		return utils.Error{Message: "Не удалось загрузить EventHandlers!"}
	}

	for _,v := range eventListeners {
		event.On(v.EventName, v, v.Priority)
	}
	return nil
}

// функция обработчик для каждого!!!
// дописать проверку на всякие ошибки!!!
func (eh EventHandler) Handle(e event.Event) error {

	method := reflect.ValueOf(eh).MethodByName(eh.TargetName)
	fmt.Println(method)

	m := reflect.ValueOf(eh).MethodByName(eh.TargetName)
	mCallable := m.Interface().(func(e event.Event) error)

	if err := mCallable(e); err != nil {
		e.Abort(true)
	}

	return nil
}

func (eh EventHandler) EmailQueueRun(e event.Event) error {
	fmt.Printf("Запуск серии писем, данные: %v\n", e.Data())
	fmt.Println("EventHandler: ", eh) // контекст серии писем, какой именно и т.д.
	return nil
}
func (eh EventHandler) WebHookCall(e event.Event) error {
	fmt.Printf("Вызов вебхука, данные: %v\n", e.Data())
	fmt.Println("EventHandler: ", eh) // контекст вебхука, какой именно и т.д.
	return nil
}


func runEvents() {
	// Register event listener


	// ... ...

	// Trigger event
	// Note: The second listener has a higher priority, so it will be executed first.
	/*err, _e := event.Fire("userCreated", event.M{"arg0": "val0", "arg1": "val1"})
	if err != nil {
		log.Println(err)
	}
	fmt.Println("_e: ", _e)*/
	// event.AddEvent(e)

	// err, _e := event.Fire(e.Name(), e.Data())

	/*e := event.NewBasic("userCreated", event.M{"id":2, "createdAt":time.Now().String()})
	err := event.FireEvent(e)
	if err != nil {
		log.Println(err)
	}*/
	// fmt.Println("_e: ", _e.IsAborted())

}
