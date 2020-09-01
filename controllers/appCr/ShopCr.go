package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

/////////////////////////////////////

func ProductCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Product
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	product, err := account.CreateProduct(input.Product)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка во время создания продукта"))
		return
	}

	resp := u.Message(true, "POST Product Created")
	resp["product"] = *product
	u.Respond(w, resp)
}

func ProductGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id товара"))
		return
	}

	var product models.Product
	// 2. Узнаем, какой список нужен
	publicIdOk:= utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicIdOk {
		_product, err := account.GetProductByPublicId(productId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
			return
		}
		product = *_product
	} else {
		_product, err := account.GetProduct(productId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
			return
		}
		product = *_product
	}


	resp := u.Message(true, "GET Product")
	resp["product"] = product
	u.Respond(w, resp)
}


func ProductListPaginationGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	// 2. Узнаем, какой список нужен
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

	products, total, err := account.GetProductListPagination(account.Id, offset, limit, sortBy, search, nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	resp := u.Message(true, "GET Product List Pagination")
	resp["total"] = total
	resp["products"] = products
	u.Respond(w, resp)
}

func ProductUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id товара"))
		return
	}

	var input map[string]interface{}
	/*var input = struct {
		models.Product	`json:"product"`
		Attributes map[string]interface{} `json:"attributes"`
	}{}*/

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	//card, err := account.UpdateProduct(productId, input.Product)
	card, err := account.UpdateProduct(productId, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Product Update")
	resp["product"] = card
	u.Respond(w, resp)
}

func ProductDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id товара"))
		return
	}


	if err = account.DeleteProduct(productId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении карточки товара"))
		return
	}

	resp := u.Message(true, "DELETE Product Successful")
	u.Respond(w, resp)
}

////////////////////////////////////

func ProductAttributeList(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	resp := u.Message(true, "GET Product List Pagination")
	// resp["total"] = total
	// resp["products"] = products
	u.Respond(w, resp)
}