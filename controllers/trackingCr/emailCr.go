package trackingCr

import (
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"time"
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


func OpenEmailByPixelUser(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err == nil && account != nil {
		mtaHistoryHashId, ok := utilsCr.GetQuerySTRVarFromGET(r, "hi")
		if ok {

			mtaHistory, err := account.GetMTAHistoryByHashId(mtaHistoryHashId)
			if err == nil {
				_ = mtaHistory.UpdateOpenUser(utilsCr.GetIP(r))
			}
		}
	}

	// ####### - Обновляем контент открытий - #######


	// 1. Находим событие по по hashId, в следствии которого, клиент получал письмо


	const transPixel = "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3B"

	// Pixel
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	w.Header().Set("Expires", "Wed, 11 Nov 1998 11:11:11 GMT")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	w.Header().Set("Pragma", "no-cache")
	fmt.Fprintf(w, transPixel)
}
