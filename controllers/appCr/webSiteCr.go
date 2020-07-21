package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func WebSiteCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.WebSite
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	webSite, err := account.CreateEntity(&input.WebSite)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания ключа"}))
		return
	}

	resp := u.Message(true, "POST WebSite Created")
	resp["webSite"] = webSite
	u.Respond(w, resp)
}

func WebSiteGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке webSite Id"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить магазин"))
		return
	}

	resp := u.Message(true, "GET WebSite ")
	resp["webSite"] = webSite
	u.Respond(w, resp)
}

func WebSiteListGet(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	sortDesc := utilsCr.GetQueryBoolVarFromGET(r, "sortDesc") // обратный или нет порядок
	sortBy, ok := utilsCr.GetQuerySTRVarFromGET(r, "sortBy")
	if !ok {
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}

	webSites, total, err := account.GetListEntity(&models.WebSite{}, sortBy)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	resp := u.Message(true, "GET WebSite List")
	resp["webSites"] = webSites
	resp["total"] = total
	u.Respond(w, resp)
}

func WebSiteListPaginationGet(w http.ResponseWriter, r *http.Request) {

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
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}
	// 2. Узнаем, какой список нужен
	all, allOk := utilsCr.GetQuerySTRVarFromGET(r, "all")


	var total uint = 0
	webSites := make([]models.Entity,0)

	if all == "true" && allOk {
		webSites, total, err = account.GetListEntity(&models.WebSite{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список ВебХуков"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		webSites, total, err = account.GetPaginationListEntity(&models.WebSite{}, offset, limit, sortBy, search)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список ВебХуков"))
			return
		}
	}


	resp := u.Message(true, "GET Web Sites PaginationList")
	resp["webSites"] = webSites
	resp["total"] = total
	u.Respond(w, resp)
}

func WebSiteUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// webSite, err := account.UpdateWebSite(webSiteId, &input.WebSite)
	err = account.UpdateEntity(&webSite, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH WebSite Update")
	resp["webSite"] = webSite
	u.Respond(w, resp)
}

func WebSiteDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID шаблона"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&webSite); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении магазина"))
		return
	}

	resp := u.Message(true, "DELETE WebSite Successful")
	u.Respond(w, resp)
}

////////////////////////////////////
// ########### Email Box ###########

func EmailBoxCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	// Get JSON-request
	var input struct{
		models.EmailBox
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	emailBox, err := webSite.CreateEmailBox(input.EmailBox)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST Email Box Created")
	resp["emailBox"] = emailBox
	u.Respond(w, resp)
}

func EmailBoxGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	emailBoxId, err := utilsCr.GetUINTVarFromRequest(r, "emailBoxId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке webSite Id"))
		return
	}

	var emailBox models.EmailBox
	err = account.LoadEntity(&emailBox, emailBoxId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить почтовый ящик"))
		return
	}

	resp := u.Message(true, "GET EmailBox")
	resp["emailBox"] = emailBox
	u.Respond(w, resp)
}

// без учета сайта
func EmailBoxFullListGet(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	sortDesc := utilsCr.GetQueryBoolVarFromGET(r, "sortDesc") // обратный или нет порядок
	sortBy, ok := utilsCr.GetQuerySTRVarFromGET(r, "sortBy")
	if !ok {
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}

	emailBoxes, total, err := account.GetListEntity(&models.EmailBox{},sortBy)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список почтовых ящиков"))
		return
	}

	resp := u.Message(true, "GET Email Box Full List")
	resp["emailBoxes"] = emailBoxes
	resp["total"] = total
	u.Respond(w, resp)
}

// Загружает весь список почтовых ящиков. Пагинации нет.
func EmailBoxListGet(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	sortDesc := utilsCr.GetQueryBoolVarFromGET(r, "sortDesc") // обратный или нет порядок
	sortBy, ok := utilsCr.GetQuerySTRVarFromGET(r, "sortBy")
	if !ok {
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}

	emailBoxes, err := webSite.GetEmailBoxList(sortBy)	                                  //GetEmailBoxList
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список почтовых ящиков"))
		return
	}

	resp := u.Message(true, "GET Email Box List")
	resp["emailBoxes"] = emailBoxes
	resp["total"] = len(emailBoxes)
	u.Respond(w, resp)
}

func EmailBoxUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailBoxId, err := utilsCr.GetUINTVarFromRequest(r, "emailBoxId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке emailBoxId Id"))
		return
	}
	var emailBox models.EmailBox
	err = account.LoadEntity(&emailBox, emailBoxId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить почтовый ящик"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&emailBox, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH EmailBox Update")
	resp["emailBox"] = emailBox
	u.Respond(w, resp)
}

func EmailBoxDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	emailBoxId, err := utilsCr.GetUINTVarFromRequest(r, "emailBoxId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке emailBoxId Id"))
		return
	}
	var emailBox models.EmailBox
	err = account.LoadEntity(&emailBox, emailBoxId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить почтовый ящик"))
		return
	}

	if err = account.DeleteEntity(&emailBox); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении почтового ящика"))
		return
	}

	resp := u.Message(true, "DELETE EmailBox Successful")
	u.Respond(w, resp)
}
