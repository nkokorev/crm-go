package models

import (
	"github.com/nkokorev/crm-go/event"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
type Event struct {}


func (Event) UserCreated(account Account, user User) event.Event {
	return event.NewBasic("UserCreated", map[string]interface{}{"accountId":account.ID, "userId":user.ID})
}

func (Event) UserAppendedToAccount(account Account, user User, role Role) event.Event {
	return event.NewBasic("UserAppendedToAccount", map[string]interface{}{"accountId":account.ID, "userId":user.ID, "roleId":role.ID})
}


