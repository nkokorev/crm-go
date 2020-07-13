package models

import (
	"errors"
	"fmt"
	"github.com/nkokorev/crm-go/event"
)

func init() {
	fmt.Println("Регистрируем слушателей")

	event.On("userCreated", event.ListenerFunc(func(e event.Event) error {
		// fmt.Printf("handle event 1: %s\n", e.Name())
		return nil
	}), event.Normal)

	// Register multiple listeners
	event.On("userAddedToAccount", event.ListenerFunc(func(e event.Event) error {

		value, ok := e.Data()["accountId"];
		if !ok {
			e.Abort(true)
			return errors.New("ID аккаунта не указано")
		}
		accountId, ok := value.(uint)
		if !ok {
			e.Abort(true)
			return errors.New("ID аккаунта неверного формата")
		}

		eventActions, err := EventActions{}.getEnabledByName(accountId, e.Name())
		if err != nil {
			return err
		}

		// fmt.Println("CallEventAction ", e)
		CallEventAction(eventActions, e)

		return nil
	}), event.Normal)
}

func CallEventAction(eventActions []EventActions, e event.Event)  {

	for _,v := range eventActions {

		switch v.TargetName {
		case "emailQueueRun":
			fmt.Printf("Запуск серии писем №%v , данные: %v\n", v.TargetId, e.Data())
		case "webHookCall":
			fmt.Printf("Вызов вебхука №%v , данные: %v\n", v.TargetId, e.Data())

		}

	}
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
