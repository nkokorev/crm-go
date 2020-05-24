package controllers

import (
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func CheckApi(w http.ResponseWriter, r *http.Request) {
	resp := u.Message(true, "Welcome to REST JSON API of RatusCRM")
	resp["help"] = "You can read more about that: https://dev.ratuscrm.com/#api"
	u.Respond(w, resp)
	return
}

func CheckAppUiApi(w http.ResponseWriter, r *http.Request) {
	resp := u.Message(true, "This is UI-API of APP RatusCRM")
	resp["help"] = "Most likely, you were looking for Public UI-API. Read more: https://dev.ratuscrm.com/#ui-api"
	u.Respond(w, resp)
	return
}
func CheckUiApi(w http.ResponseWriter, r *http.Request) {
	resp := u.Message(true, "Welcome to Public UI-API of RatusCRM")
	resp["help"] = "You can read more about that: https://dev.ratuscrm.com/#ui-api"
	u.Respond(w, resp)
	return
}