package models

import (
	"github.com/nkokorev/crm-go/event"
	"reflect"
)

// Возвращает список доступных событий в системе
func GetSystemEventList() []string { // el := EventList{}
	list := make([]string,0)

	t := reflect.TypeOf(&Event{})
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		list = append(list, m.Name)
	}

	return list
}

// Структура со списком всех event's.
// !!! Не добавлять любые другие функции под этим интерфейсом !!!
type Event struct {}


func (Event) UserCreated(account Account, user User) event.Event {
	return event.NewBasic("UserCreated", map[string]interface{}{"accountId":account.ID, "userId":user.ID})
}

func (Event) UserAppendedToAccount(account Account, user User, role Role) event.Event {
	return event.NewBasic("UserAppendedToAccount", map[string]interface{}{"accountId":account.ID, "userId":user.ID, "roleId":role.ID})
}


