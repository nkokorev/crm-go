package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

// Ui / API there!
func OrderCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Order
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// 1. Создаем нового клиента
	if input.CustomerId != nil {

	}
	// 2. Создаем что-то еще

	
	order, err := account.CreateEntity(&input.Order)
	if err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка во время создания заказа"))
		// u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заявки"}))
		return
	}

	resp := u.Message(true, "POST Order Created")
	resp["order"] = order
	u.Respond(w, resp)
}

func OrderGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}
	
	orderId, err := utilsCr.GetUINTVarFromRequest(r, "orderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	var order models.Order

	// 2. Узнаем, какой объект нужен
	publicIdOk:= utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	if publicIdOk {
		err = account.LoadEntityByPublicId(&order, orderId, preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список заказов"))
			return
		}
	} else {
		err = account.LoadEntity(&order, orderId, preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список заказов"))
			return
		}
	}

	resp := u.Message(true, "GET Order")
	resp["order"] = order
	u.Respond(w, resp)
}

func OrderGetListPagination(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
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

	var total int64 = 0
	orders := make([]models.Entity,0)
	
	orders, total, err = account.GetPaginationListEntity(&models.Order{}, offset, limit, sortBy, search, nil,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET Order Pagination List")
	resp["total"] = total
	resp["orders"] = orders
	u.Respond(w, resp)
}

func OrderUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	orderId, err := utilsCr.GetUINTVarFromRequest(r, "orderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var order models.Order
	err = account.LoadEntity(&order, orderId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&order, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Order Update")
	resp["order"] = order
	u.Respond(w, resp)
}

func OrderDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	orderId, err := utilsCr.GetUINTVarFromRequest(r, "orderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}


	var order models.Order
	err = account.LoadEntity(&order, orderId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}
	if err = account.DeleteEntity(&order); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении заказа"))
		return
	}

	resp := u.Message(true, "DELETE Order Successful")
	u.Respond(w, resp)
}

func OrderAppendProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	orderId, err := utilsCr.GetUINTVarFromRequest(r, "orderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id warehouseId"))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var order models.Order
	if err =account.LoadEntity(&order, orderId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = order.AppendProduct(product); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка добавления продукта в заказ"))
		return
	}

	if err = account.LoadEntity(&order, orderId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка загрузки данных заказа"}))
		return
	}

	resp := u.Message(true, "PATCH Order Append Product")
	resp["order"] = order
	u.Respond(w, resp)
}

func OrderRemoveProduct(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	orderId, err := utilsCr.GetUINTVarFromRequest(r, "orderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id warehouseId"))
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

	var order models.Order
	if err = account.LoadEntity(&order, orderId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}

	if err = order.RemoveProduct(product); err !=nil {
		u.Respond(w, u.MessageError(err, "Ошибка удаления товара из заказа"))
		return
	}

	// Обновляем данные карточки
	if err =account.LoadEntity(&order, orderId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}

	resp := u.Message(true, "PATCH Order Remove Products")
	resp["order"] = order
	u.Respond(w, resp)
}

func OrderSetUnknownCustomer(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	orderId, err := utilsCr.GetUINTVarFromRequest(r, "orderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var order models.Order
	err = account.LoadEntity(&order, orderId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	if err := order.SetUnknownCustomer(); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении данных заказчика"))
		return
	}

	_ = account.LoadEntity(&order, orderId, preloads)

	resp := u.Message(true, "PATCH Order Set Unknown Customer")
	resp["order"] = order
	u.Respond(w, resp)
}

func OrderUpdateReserve(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	orderId, err := utilsCr.GetUINTVarFromRequest(r, "orderId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var order models.Order
	err = account.LoadEntity(&order, orderId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}
	
	var input models.ReserveCartItem
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// обнуляем лишнее
	input.Quantity = nil
	input.WarehouseId = nil
	input.Wasted = nil

	if input.Reserved != nil {
		err = order.UpdateReserveCartItems(&input)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
			return
		}
	}

	err = account.LoadEntity(&order, orderId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	resp := u.Message(true, "PATCH Order Update")
	resp["order"] = order
	u.Respond(w, resp)
}