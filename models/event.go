package models

import (
	"github.com/nkokorev/crm-go/event"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
// При добавлении нового события - внести в список EventItem!!!
type Event struct {}


func (Event) UserCreated(accountId uint, userId uint) event.Event {
	return event.NewBasic("UserCreated", map[string]interface{}{"accountId":accountId, "userId":userId})
}

func (Event) UserAppendedToAccount(accountId, userId, roleId uint) event.Event {
	return event.NewBasic("UserAppendedToAccount", map[string]interface{}{"accountId":accountId, "userId":userId, "roleId":roleId})
}


