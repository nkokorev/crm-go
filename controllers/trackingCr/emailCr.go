package trackingCr

import (
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

//  Query:
//  ?u={userHashId}   <<< Минимальный пакет данных
// 	?u={userHashId}&?i={mtaHistoryId}&hi={hashId}
// 	?u={userHashId}&hi={hashId}
func UnsubscribeUser(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	userHashId, ok := utilsCr.GetQuerySTRVarFromGET(r, "u")
	if !ok {
		u.Respond(w, u.MessageError(err, "Необходимо указать пользователя"))
		return
	}

	// Получаем пользователя
	user, err := account.GetUserByHashId(userHashId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти пользователя"))
		return
	}

	// Если пользователь уже отписан
	if !user.Subscribed {
		u.Respond(w, u.MessageError(err, "Пользователь уже отписан от всех рассылок"))
		return
	}

	if err := user.Unsubscribing(); err != nil {
		u.Respond(w, u.MessageWithErrors("Ошибка отписки пользователя", nil))
		return
	}

	// ####### - Обновляем контент отписки - #######


	// 1. Находим событие по по hashId, в следствии которого, клиент получал письмо
	mtaHistoryHashId, ok := utilsCr.GetQuerySTRVarFromGET(r, "hi")
	if ok {
		mtaHistory, err := account.GetMTAHistoryByHashId(mtaHistoryHashId)
		if err == nil {
			// если hashId Отправки совпадает, то отписываем. Это как проверочный код.
			if mtaHistory.HashId == mtaHistoryHashId {
				_ = mtaHistory.UpdateSetUnsubscribeUser(utilsCr.GetIP(r))
			}
		}
	}

	resp := u.Message(true, "User unsubscribed!")
	u.Respond(w, resp)
}
