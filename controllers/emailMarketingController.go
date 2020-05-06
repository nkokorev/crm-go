package controllers

import (
	"fmt"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func GetDomains(w http.ResponseWriter, r *http.Request) {
	
	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	fmt.Println("Domains controller: ", account.Name)

	/*domains, err := account.GetDomains()
	if err != nil || domains == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"domains":"Не удалось получить список доменов"}}))
		return
	}*/

	resp := u.Message(true, "GET account domains")
	// resp["domains"] = domains
	u.Respond(w, resp)
}
