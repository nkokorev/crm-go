package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func ShipmentCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Shipment
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	shipment, err := account.CreateEntity(&input.Shipment)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST Shipment Created")
	resp["shipment"] = shipment
	u.Respond(w, resp)
}
func ShipmentGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	shipmentId, err := utilsCr.GetUINTVarFromRequest(r, "shipmentId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке shipment Id"))
		return
	}

	var shipment models.Shipment
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&shipment, shipmentId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&shipment, shipmentId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось загрузить данные поставки"))
			return
		}
	}

	resp := u.Message(true, "GET Shipment")
	resp["shipment"] = shipment
	u.Respond(w, resp)
}
func ShipmentListPaginationGet(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	// 1. Получаем рабочий аккаунт в зависимости от источника (автома. сверка с {hashId}.)

	account, err = utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 25
	}
	offset, ok := utilsCr.GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	sortDesc := utilsCr.GetQueryBoolVarFromGET(r, "sortDesc") // обратный или нет порядок
	sortBy, ok := utilsCr.GetQuerySTRVarFromGET(r, "sortBy")
	if !ok {
		sortBy = "id"
	}
	if sortDesc {
		sortBy += " desc"
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	// 2. Узнаем, какой список нужен
	all := utilsCr.GetQueryBoolVarFromGET(r, "all")

	// Узнаем нужен ли фильтр
	filter := map[string]interface{}{}

	var total int64 = 0
	shipments := make([]models.Entity,0)

	if all && len(filter) < 1{
		shipments, total, err = account.GetListEntity(&models.Shipment{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		shipments, total, err = account.GetPaginationListEntity(&models.Shipment{}, offset, limit, sortBy, search, filter, preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	}
	
	resp := u.Message(true, "GET product Cards PaginationList")
	resp["shipments"] = shipments
	resp["total"] = total
	u.Respond(w, resp)
}
func ShipmentUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shipmentId, err := utilsCr.GetUINTVarFromRequest(r, "shipmentId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	var shipment models.Shipment

	err = account.LoadEntity(&shipment, shipmentId,nil)

	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&shipment, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Shipment Update")
	resp["shipment"] = shipment
	u.Respond(w, resp)
}

func ShipmentDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shipmentId, err := utilsCr.GetUINTVarFromRequest(r, "shipmentId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке shipment Id"))
		return
	}

	var shipment models.Shipment
	err = account.LoadEntity(&shipment, shipmentId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&shipment); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении поставки"))
		return
	}

	resp := u.Message(true, "DELETE Shipment Successful")
	u.Respond(w, resp)
}
func ShipmentAppendProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shipmentId, err := utilsCr.GetUINTVarFromRequest(r, "shipmentId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке shipmentId"))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var shipment models.Shipment
	if err =account.LoadEntity(&shipment, shipmentId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке данных инвентаризации"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	// Декодируем параметры (нужно ли?)
	var input struct{
		VolumeOrder		float64 `json:"volume_order" gorm:"type:numeric;"` // 
		PaymentAmount	float64 `json:"payment_amount" gorm:"type:numeric;"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if err = shipment.AppendProduct(product, input.VolumeOrder, input.PaymentAmount); err !=nil {
		u.Respond(w, u.MessageError(err, "Ошибка добавления продукта в поставку"))
		return
	}

	if err = account.LoadEntity(&shipment, shipmentId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка загрузки данных поставки"}))
		return
	}

	resp := u.Message(true, "PATCH Shipment Append Product")
	resp["shipment"] = shipment
	u.Respond(w, resp)
}
func ShipmentRemoveProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shipmentId, err := utilsCr.GetUINTVarFromRequest(r, "shipmentId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id shipmentId"))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}
	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var shipment models.Shipment
	if err =account.LoadEntity(&shipment, shipmentId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке данных поставки"}))
		return
	}
	
	if err = shipment.RemoveProduct(product); err !=nil {
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из поставки"))
		return
	}

	// Обновляем данные карточки
	if err =account.LoadEntity(&shipment, shipmentId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке данных поставки"}))
		return
	}

	resp := u.Message(true, "PATCH Shipment Remove Products")
	resp["shipment"] = shipment
	u.Respond(w, resp)
}


