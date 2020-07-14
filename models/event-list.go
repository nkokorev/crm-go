package models

import (
	"fmt"
	"reflect"
)

// Возвращает список доступных событий в системе
func GetSystemEventList() []string {
	// el := EventList{}
	list := make([]string,0)

	t := reflect.TypeOf(&Event{})
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		fmt.Println(m.Name)
		list = append(list, m.Name)
	}

	return list
}

// Структура со списком всех event's
type Event struct {}


func (Event) UserCreated(account Account, user User) (string, map[string]interface{}){
	return "UserAppendedToAccount", map[string]interface{}{"accountId":account.ID, "userId":user.ID}
}

func (Event) UserAppendedToAccount(account Account, user User, role Role) (string, map[string]interface{}){
	return "UserAppendedToAccount", map[string]interface{}{"accountId":account.ID, "userId":user.ID, "roleId":role.ID}
}



