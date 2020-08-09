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
// 	?u={userHashId}
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

	user, err := account.GetUserByHashId(userHashId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти пользователя 2"))
		return
	}

	// Если пользователь уже отписан
	if !user.Subscribed {
		u.Respond(w, u.MessageError(err, "Пользователь уже отписан от всех рассылок"))
		return
	}

	// получаем контекст отписки
	// &w=<>
/*	userHashId, ok := utilsCr.GetQuerySTRVarFromGET(r, "w")
	if !ok {
		u.Respond(w, u.MessageError(err, "Необходимо указать пользователя"))
		return
	}*/

	// Тут функция отписки пользователя.
	update := map[string]interface{} {
		"subscribed":false,
		"unsubscribedAt":time.Now().UTC(),

		// неизвестная причина, т.к. человек пришел по ссылке. Потом можно сделать историю...
		"unsubscribedReason" : "unsubscribe_url",
	}
	
	defer account.UpdateUser(user.Id, update)

	

	fmt.Println("user: ", user.Email)


	/*limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 25
	}
	if limit > 100 { limit = 100 }
	offset, ok := utilsCr.GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	sortDesc := utilsCr.GetQueryBoolVarFromGET(r, "sortDesc") // обратный или нет порядок
	sortBy, ok := utilsCr.GetQuerySTRVarFromGET(r, "sortBy")
	if !ok {
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}

	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	var total uint = 0
	orders := make([]models.Entity,0)

	orders, total, err = account.GetPaginationListEntity(&models.Order{}, offset, limit, sortBy, search, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}*/

	resp := u.Message(true, "User unsubscribed!")
	u.Respond(w, resp)
}
