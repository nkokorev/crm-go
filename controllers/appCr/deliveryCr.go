package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"

)

func DeliveryCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	delivery, err := webSite.CreateDelivery(input)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST Delivery Create")
	resp["delivery"] = delivery
	u.Respond(w, resp)
}

func DeliveryGetListByShop(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	deliveries := webSite.GetDeliveryMethods()

	resp := u.Message(true, "GET Deliveries List By WebSite")
	resp["deliveries"] = deliveries
	u.Respond(w, resp)
}

func DeliveryGetList(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error

	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	/*offset, ok := utilsCr.GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 100
	}
	sortBy, ok := utilsCr.GetQuerySTRVarFromGET(r, "sortBy")
	if !ok {
		sortBy = ""
	}
	sortDesc := utilsCr.GetQueryBoolVarFromGET(r, "sortDesc")
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}*/

	deliveries := webSite.GetDeliveryMethods()

	resp := u.Message(true, "GET Deliveries List Pagination")
	// resp["webHooks"] = webHooks
	resp["deliveries"] = deliveries
	u.Respond(w, resp)
}

func DeliveryUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Работаем через метод магазина, т.к. доставки разные и это интерфейс
	delivery, err := webSite.UpdateDelivery(input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH WebHook Update")
	resp["delivery"] = delivery
	u.Respond(w, resp)
}

func DeliveryDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Работаем через метод магазина, т.к. доставки разные и это интерфейс
	if err = webSite.DeleteDelivery(input); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении метода доставки"))
		return
	}

	resp := u.Message(true, "DELETE Delivery")
	u.Respond(w, resp)
}


