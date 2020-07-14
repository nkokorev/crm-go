package models

import (
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
)

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

// Возвращает список обработчиков, которые могут быть назначены на event
/*func (EventHandler) EventList () map[string]interface{} {
	return map[string]interface{}{
		"userCreated": map[string]string{"description":"Создан новый пользователь в рамках текущего аккаунта"},
		"userAddedToAccount": map[string]string{"description":"Пользователь добавлен в аккаунт"},
		"userRemovedFromAccount": map[string]string{"description":"Исключение пользователя из аккаунта"},
	}
}

func (EventHandler) HandleList () map[string]interface{} {
	return map[string]interface{}{
		"EmailQueueRun": map[string]string{"description":"Запуск автоматической серии email писем с указанным ID."},
		"WebHookCall": map[string]string{"description":"Вызов WebHook с указанным ID."},
	}
}*/



/*func (EventHandler) GetTargets () map[string]func(e event.Event) error {
	return map[string]func(e event.Event)error {
		"EmailQueueRun": EventHandler{}.EmailQueueRun,
		"WebHookCall": EventHandler{}.WebHookCall,
	}
}*/

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
