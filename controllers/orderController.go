package controllers

import (
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func GetOrders(w http.ResponseWriter, r *http.Request) {

	// Аккаунт, в котором происходит авторизация: issuerAccount
	/*err, issuerAccount := GetIssuerAccount(w,r)
	if err != nil || issuerAccount == nil {
		return
	}*/

	err, account := GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	err, orders := account.GetOrders()
	if err != nil || orders == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"orders":"Не удалось получить список заказов"}}))
		return
	}

	resp := u.Message(true, "GET account orders")
	resp["orders"] = orders
	u.Respond(w, resp)
}
