package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func CartItemCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// fixs
	u.Respond(w, u.MessageError(u.Error{Message:"Ошибка при создании: модуль еще не в работе"}))
	return

	// Get JSON-request
	var input struct{
		models.CartItem
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	cartItem, err := account.CreateEntity(&input.CartItem)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания объекта"}))
		return
	}

	resp := u.Message(true, "POST Cart Item Created")
	resp["cart_item"] = cartItem
	u.Respond(w, resp)
}

func CartItemGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	// ThisIs PublicID or inside
	cartItemId, err := utilsCr.GetUINTVarFromRequest(r, "cartItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var cartItem models.CartItem

	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&cartItem, cartItemId, preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&cartItem, cartItemId, preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	}

	resp := u.Message(true, "GET Cart Item")
	resp["cart_item"] = cartItem
	u.Respond(w, resp)
}

func CartItemGetListPagination(w http.ResponseWriter, r *http.Request) {

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

	// возвращаемые переменные
	var total int64 = 0
	cartItems := make([]models.Entity,0)

	// 2. Узнаем, какой список нужен
	all := utilsCr.GetQueryBoolVarFromGET(r, "all")

	if all {
		cartItems, total, err = account.GetListEntity(&models.CartItem{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	} else {
		cartItems, total, err = account.GetPaginationListEntity(&models.CartItem{}, offset, limit, sortBy, search, nil,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список"))
			return
		}
	}
	


	resp := u.Message(true, "GET Cart Item Pagination List")
	resp["total"] = total
	resp["cart_items"] = cartItems
	u.Respond(w, resp)
}

func CartItemUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	cartItemId, err := utilsCr.GetUINTVarFromRequest(r, "cartItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var cartItem models.CartItem
	err = account.LoadEntity(&cartItem, cartItemId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Статус меняется только через отдельную функцию
	delete(input,"amount")
	delete(input,"payment_subject_yandex")
	delete(input,"payment_mode_yandex")
	delete(input,"product")
	delete(input,"order")

	err = account.UpdateEntity(&cartItem, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Cart Item Update")
	resp["cart_item"] = cartItem
	u.Respond(w, resp)
}

func CartItemDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	cartItemId, err := utilsCr.GetUINTVarFromRequest(r, "cartItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var cartItem models.CartItem
	err = account.LoadEntity(&cartItem, cartItemId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}
	if err = account.DeleteEntity(&cartItem); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении cartItem"))
		return
	}
	

	resp := u.Message(true, "DELETE Cart Item Successful")
	u.Respond(w, resp)
}

// -- ### Reserve ### --
func CartItemUpdateReserve(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	cartItemId, err := utilsCr.GetUINTVarFromRequest(r, "cartItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var cartItem models.CartItem
	err = account.LoadEntity(&cartItem, cartItemId, []string{"WarehouseItem"}) // загружаем с WhItem т.к. в нем будет warehouse_id
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input models.ReserveCartItem
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if input.Quantity != nil && *input.Quantity < 0 { *input.Quantity = -1 * *input.Quantity}

	if err := cartItem.UpdateReserve(input); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	_ = account.LoadEntity(&cartItem, cartItemId, preloads)

	resp := u.Message(true, "Update Cart Item reserve")
	resp["cart_item"] = cartItem
	u.Respond(w, resp)
}
func CartItemGetWarehouseItems(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	cartItemId, err := utilsCr.GetUINTVarFromRequest(r, "cartItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	// preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var cartItem models.CartItem
	err = account.LoadEntity(&cartItem, cartItemId, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input models.ReserveCartItem
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	resp := u.Message(true, "Update Cart Item reserve")
	resp["warehouse_items"] = cartItem.GetAvailabilityWarehouseItems()
	u.Respond(w, resp)
}

func CartItemCreateReserve(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	cartItemId, err := utilsCr.GetUINTVarFromRequest(r, "cartItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var cartItem models.CartItem
	err = account.LoadEntity(&cartItem, cartItemId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Статус меняется только через отдельную функцию
	delete(input,"amount")
	delete(input,"payment_subject_yandex")
	delete(input,"payment_mode_yandex")
	delete(input,"product")
	delete(input,"order")

	err = account.UpdateEntity(&cartItem, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Cart Item Update")
	resp["cart_item"] = cartItem
	u.Respond(w, resp)
}
func CartItemRemoveReserve(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	cartItemId, err := utilsCr.GetUINTVarFromRequest(r, "cartItemId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var cartItem models.CartItem
	err = account.LoadEntity(&cartItem, cartItemId, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Статус меняется только через отдельную функцию
	delete(input,"amount")
	delete(input,"payment_subject_yandex")
	delete(input,"payment_mode_yandex")
	delete(input,"product")
	delete(input,"order")

	err = account.UpdateEntity(&cartItem, input, preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Cart Item Update")
	resp["cart_item"] = cartItem
	u.Respond(w, resp)
}


