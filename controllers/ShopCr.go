package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func ShopCreate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Shop
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	shop, err := account.CreateShop(input.Shop)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания ключа"}))
		return
	}

	resp := u.Message(true, "POST Shop Created")
	resp["shop"] = *shop
	u.Respond(w, resp)
}

func ShopListGet(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
		return
	}
	
	shops, err := account.GetShops()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}



	resp := u.Message(true, "GET Shop List")
	resp["shops"] = shops
	u.Respond(w, resp)
}

func ShopUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	// Get JSON-request
	var input struct{
		models.Shop
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	shop, err := account.UpdateShop(shopId, &input.Shop)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Shop Update")
	resp["shop"] = shop
	u.Respond(w, resp)
}

func ShopDelete(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	shopId, err := GetUINTVarFromRequest(r, "shopId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	if err = account.DeleteShop(shopId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении магазина"))
		return
	}

	resp := u.Message(true, "DELETE Shop Successful")
	u.Respond(w, resp)
}

/////////////////////////////////////

func ProductGroupCreate(w http.ResponseWriter, r *http.Request) {

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

func ProductGroupListGet(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
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

	groups, err := shop.GetProductGroups()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}



	resp := u.Message(true, "GET Product Group List")
	resp["groups"] = groups
	u.Respond(w, resp)
}

func ProductGroupUpdate(w http.ResponseWriter, r *http.Request) {

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

func ProductGroupDelete(w http.ResponseWriter, r *http.Request) {
	
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


/////////////////////////////////////

func ProductCardCreate(w http.ResponseWriter, r *http.Request) {

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

func ProductCardListGet(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
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

	groups, err := shop.GetProductGroups()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}



	resp := u.Message(true, "GET Product Group List")
	resp["groups"] = groups
	u.Respond(w, resp)
}

func ProductCardUpdate(w http.ResponseWriter, r *http.Request) {

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

func ProductCardDelete(w http.ResponseWriter, r *http.Request) {

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