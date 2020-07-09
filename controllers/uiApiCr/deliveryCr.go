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

	shopId, err := utilsCr.GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке shopId"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	resp := u.Message(true, "GET Shop Deliveries")
	resp["deliveryMethods"] = shop.GetDeliveryMethods()
	u.Respond(w, resp)
}

func DeliveryCalculateDeliveryCost(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	shopId, err := utilsCr.GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке shopId"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}
	
	var input models.DeliveryRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	deliveryData, err := shop.CalculateDelivery(input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка расчета стоимости доставки"))
		return
	}

	resp := u.Message(true, "GET Calculate Delivery")
	resp["deliveryData"] = deliveryData
	u.Respond(w, resp)
}
