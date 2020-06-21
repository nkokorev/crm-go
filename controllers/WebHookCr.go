package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func WebHookCreate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	// Get JSON-request
	var input struct{
		models.ProductGroup
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	group, err := shop.CreateProductGroup(input.ProductGroup)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания группы"}))
		return
	}

	resp := u.Message(true, "POST ProductGroup Created")
	resp["group"] = *group
	u.Respond(w, resp)
}

func WebHookGet(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error

	// 1. Получаем рабочий аккаунт в зависимости от источника (автома. сверка с {hashId}.)
	if isApiRequest(r) {
		account, err = GetWorkAccount(w,r)
		if err != nil || account == nil {
			return
		}
	} else {
		account, err = GetWorkAccountCheckHashId(w,r)
		if err != nil || account == nil {
			return
		}
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	productCardId, err := GetUINTVarFromRequest(r, "productCardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID product card"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	productCard, err := shop.GetProductCard(productCardId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	resp := u.Message(true, "GET Product Card")
	resp["productCard"] = productCard
	u.Respond(w, resp)
}

func WebHookListPaginationGet(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	// 1. Получаем рабочий аккаунт в зависимости от источника (автома. сверка с {hashId}.)
	if isApiRequest(r) {
		account, err = GetWorkAccount(w,r)
		if err != nil || account == nil {
			return
		}
	} else {
		account, err = GetWorkAccountCheckHashId(w,r)
		if err != nil || account == nil {
			return
		}
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	// 2. Узнаем, какой список нужен
	all, allOk := GetQuerySTRVarFromGET(r, "all")

	limit, ok := GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 100
	}
	offset, ok := GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	search, ok := GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	productGroups := make([]models.ProductGroup,0)
	total := 0

	if all == "true" && allOk {
		productGroups, err = shop.GetProductGroups()
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
			return
		}
	} else {
		productGroups, total, err = shop.GetProductGroupsPaginationList(offset, limit, search)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
			return
		}
	}


	resp := u.Message(true, "GET Product Group List")
	resp["productGroups"] = productGroups
	resp["total"] = total
	u.Respond(w, resp)
}

func WebHookUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	groupId, err := GetUINTVarFromRequest(r, "groupId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID группы"))
		return
	}

	// var input interface{}
	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// group, err := shop.UpdateProductGroup(groupId, &input.ProductGroup)
	group, err := shop.UpdateProductGroup(groupId, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH ProductGroup Update")
	resp["group"] = group
	u.Respond(w, resp)
}

func WebHookDelete(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	shop, err := account.GetShop(shopId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	groupId, err := GetUINTVarFromRequest(r, "groupId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}


	if err = shop.DeleteProductGroup(groupId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении товарной группы"))
		return
	}

	resp := u.Message(true, "DELETE ProductGroup Successful")
	u.Respond(w, resp)
}
