package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func WarehouseCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Warehouse
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	warehouse, err := account.CreateEntity(&input.Warehouse)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST Warehouse Created")
	resp["warehouse"] = warehouse
	u.Respond(w, resp)
}

func WarehouseGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	warehouseId, err := utilsCr.GetUINTVarFromRequest(r, "warehouseId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке warehouse Id"))
		return
	}

	var warehouse models.Warehouse
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&warehouse, warehouseId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&warehouse, warehouseId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось загрузить магазин"))
			return
		}
	}

	resp := u.Message(true, "GET Warehouse")
	resp["warehouse"] = warehouse
	u.Respond(w, resp)
}

func WarehouseListPaginationGet(w http.ResponseWriter, r *http.Request) {

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
	/*
	webSiteId, _filterWebSite := utilsCr.GetQueryUINTVarFromGET(r, "webSiteId")
	if _filterWebSite {
		filter["web_site_id"] = webSiteId
	}*/

	var total int64 = 0
	warehouses := make([]models.Entity,0)

	if all && len(filter) < 1{
		warehouses, total, err = account.GetListEntity(&models.Warehouse{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		warehouses, total, err = account.GetPaginationListEntity(&models.Warehouse{}, offset, limit, sortBy, search, filter,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	}
	
	resp := u.Message(true, "GET product Cards PaginationList")
	resp["warehouses"] = warehouses
	resp["total"] = total
	u.Respond(w, resp)
}

func WarehouseUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	warehouseId, err := utilsCr.GetUINTVarFromRequest(r, "warehouseId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	var warehouse models.Warehouse

	err = account.LoadEntity(&warehouse, warehouseId,nil)

	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&warehouse, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Warehouse Update")
	resp["warehouse"] = warehouse
	u.Respond(w, resp)
}

func WarehouseDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	warehouseId, err := utilsCr.GetUINTVarFromRequest(r, "warehouseId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var warehouse models.Warehouse
	err = account.LoadEntity(&warehouse, warehouseId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&warehouse); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении магазина"))
		return
	}

	resp := u.Message(true, "DELETE Warehouse Successful")
	u.Respond(w, resp)
}
func WarehouseAppendProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	warehouseId, err := utilsCr.GetUINTVarFromRequest(r, "warehouseId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var warehouse models.Warehouse
	if err =account.LoadEntity(&warehouse, warehouseId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	// Декодируем параметры
	var input struct{
		AmountUnit float64 `json:"amount_unit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if err = warehouse.AppendProduct(product, input.AmountUnit); err !=nil {
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из карточки товара"))
		return
	}

	if err = account.LoadEntity(&warehouse, warehouseId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка загрузки карточки товара"}))
		return
	}

	resp := u.Message(true, "PATCH Warehouse Append Product")
	resp["warehouse"] = warehouse
	u.Respond(w, resp)
}

func WarehouseRemoveProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	warehouseId, err := utilsCr.GetUINTVarFromRequest(r, "warehouseId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
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

	var warehouse models.Warehouse
	if err =account.LoadEntity(&warehouse, warehouseId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}
	
	if err = warehouse.RemoveProduct(product); err !=nil {
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из карточки товара"))
		return
	}

	// Обновляем данные карточки
	if err =account.LoadEntity(&warehouse, warehouseId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}

	resp := u.Message(true, "PATCH Warehouse Remove Products")
	resp["warehouse"] = warehouse
	u.Respond(w, resp)
}


