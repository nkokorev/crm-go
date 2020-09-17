package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func CompanyCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Company
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	productCard, err := account.CreateEntity(&input.Company)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST Company Created")
	resp["company"] = productCard
	u.Respond(w, resp)
}

func CompanyGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	companyId, err := utilsCr.GetUINTVarFromRequest(r, "companyId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productCard Id"))
		return
	}

	var productCard models.Company
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&productCard, companyId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&productCard, companyId,preloads)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(err, "Не удалось загрузить магазин"))
			return
		}
	}

	resp := u.Message(true, "GET Company")
	resp["company"] = productCard
	u.Respond(w, resp)
}

func CompanyListPaginationGet(w http.ResponseWriter, r *http.Request) {

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
	companies := make([]models.Entity,0)

	if all && len(filter) < 1{
		companies, total, err = account.GetListEntity(&models.Company{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		companies, total, err = account.GetPaginationListEntity(&models.Company{}, offset, limit, sortBy, search, filter,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	}
	
	resp := u.Message(true, "GET Companies pagination list")
	resp["companies"] = companies
	resp["total"] = total
	u.Respond(w, resp)
}

func CompanyUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	companyId, err := utilsCr.GetUINTVarFromRequest(r, "companyId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")
	var productCard models.Company

	err = account.LoadEntity(&productCard, companyId,nil)

	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&productCard, input,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Company Update")
	resp["company"] = productCard
	u.Respond(w, resp)
}

func CompanyDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	companyId, err := utilsCr.GetUINTVarFromRequest(r, "companyId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var productCard models.Company
	err = account.LoadEntity(&productCard, companyId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&productCard); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении карточки товара"))
		return
	}

	resp := u.Message(true, "DELETE Company Successful")
	u.Respond(w, resp)
}

