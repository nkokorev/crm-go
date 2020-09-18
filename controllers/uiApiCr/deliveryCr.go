package uiApiCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"net/http"
)

// Возвращает список доступных доставок
func DeliveryGetListByShop(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	// 1. Получаем рабочий аккаунт в зависимости от источника (автома. сверка с {hashId}.)

	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке webSiteId"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	resp := u.Message(true, "GET WebSite Deliveries")
	resp["deliveryMethods"] = webSite.GetDeliveryMethods()
	u.Respond(w, resp)
}

func DeliveryCalculateDeliveryCost(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке webSiteId"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}
	
	var input models.DeliveryRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// check - проверяем данные.

	totalCost, weight, err := webSite.CalculateDelivery(input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка расчета стоимости доставки"))
		// u.Respond(w, u.MessageError(u.Error{Message:"Ошибка расчета стоимости доставки", Errors: map[string]interface{}{"delivery":err.Error()}}))
		return
	}

	resp := u.Message(true, "GET Calculate Delivery")
	resp["weight"] = weight
	resp["total_cost"] = totalCost
	u.Respond(w, resp)
}

func DeliveryCodeList(w http.ResponseWriter, r *http.Request) {

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
	err = account.LoadEntity(&webSite, webSiteId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	resp := u.Message(true, "GET Deliveries List Options By WebSite")
	resp["delivery_code_list"] = webSite.DeliveryCodeList()
	u.Respond(w, resp)
}
