package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func WarehouseItemCreate(w http.ResponseWriter, r *http.Request) {

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

func WarehouseItemGet(w http.ResponseWriter, r *http.Request) {

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

func WarehouseItemListPaginationGet(w http.ResponseWriter, r *http.Request) {

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
	warehouseId, _filterWarehouse := utilsCr.GetQueryUINTVarFromGET(r, "warehouseId")
	if _filterWarehouse {
		filter["warehouse_id"] = warehouseId
	}
	productId, _filterProduct := utilsCr.GetQueryUINTVarFromGET(r, "productId")
	if _filterProduct {
		filter["product_id"] = productId
	}

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

	resp := u.Message(true, "GET Warehouse Item PaginationList")
	resp["warehouses"] = warehouses
	resp["total"] = total
	u.Respond(w, resp)
}

func WarehouseItemUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	warehouseItemId, err := utilsCr.GetUINTVarFromRequest(r, "warehouseItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	var warehouseItem models.WarehouseItem

	err = account.LoadEntity(&warehouseItem, warehouseItemId,nil)

	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if valStr, ok := input["stock"]; ok {
		f64, ok1 := valStr.(float64)
		if ok1 {
			if f64 < 0 {
				input["stock"] = -1*f64
			}
		}
	}
	if valStr, ok2 := input["reservation"]; ok2 {
		f64, ok3 := valStr.(float64)
		if ok3 {
			if f64 < 0 {
				input["reservation"] = -1*f64
			}
		}
	}

	err = account.UpdateEntity(&warehouseItem, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Warehouse Item Update")
	resp["warehouse_item"] = warehouseItem
	u.Respond(w, resp)
}

func WarehouseItemDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	warehouseItemId, err := utilsCr.GetUINTVarFromRequest(r, "warehouseItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var warehouseItem models.WarehouseItem
	err = account.LoadEntity(&warehouseItem, warehouseItemId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&warehouseItem); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении объекта склада"))
		return
	}

	resp := u.Message(true, "DELETE warehouse Item Successful")
	u.Respond(w, resp)
}



