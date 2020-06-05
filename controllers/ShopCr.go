package controllers

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

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

	shopId, err := GetUINTVarFromRequest(r, "id")
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

	shopId, err := GetUINTVarFromRequest(r, "id")
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
