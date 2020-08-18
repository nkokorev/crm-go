package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func ArticleCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Article
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	article, err := account.CreateEntity(&input.Article)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка во время создания продукта"))
		return
	}

	resp := u.Message(true, "POST Article Created")
	resp["article"] = article
	u.Respond(w, resp)
}

func ArticleGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	articleId, err := utilsCr.GetUINTVarFromRequest(r, "articleId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id статьи"))
		return
	}

	var article models.Article

	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "publicId")

	if publicOk  {
		err = account.LoadEntityByPublicId(&article, articleId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&article, articleId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	}



	resp := u.Message(true, "GET Article")
	resp["article"] = article
	u.Respond(w, resp)
}

func ArticleListPaginationGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
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
	articles := make([]models.Entity, 0)

	if all == "true" && allOk {
		articles, total, err = account.GetListEntity(&models.Article{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить данные"))
			return
		}
	} else {
		// emailNotifications, total, err = account.GetEmailNotificationsPaginationList(offset, limit, search)
		articles, total, err = account.GetPaginationListEntity(&models.Article{}, offset, limit, sortBy, search, nil)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить данные"))
			return
		}
	}
	
	resp := u.Message(true, "GET Article List Pagination")
	resp["total"] = total
	resp["articles"] = articles
	u.Respond(w, resp)
}

func ArticleUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	articleId, err := utilsCr.GetUINTVarFromRequest(r, "articleId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id товара"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	var article models.Article
	if err = account.LoadEntity(&article, articleId); err != nil {
		u.Respond(w, u.MessageError(err, "Статья не найдена"))
		return
	}

	err = account.UpdateEntity(&article, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Article Update")
	resp["article"] = article
	u.Respond(w, resp)
}

func ArticleDelete(w http.ResponseWriter, r *http.Request) {

	
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	articleId, err := utilsCr.GetUINTVarFromRequest(r, "articleId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id товара"))
		return
	}

	var article models.Article
	if err = account.LoadEntity(&article, articleId); err != nil {
		u.Respond(w, u.MessageError(err, "Статья не найдена"))
		return
	}

	if err = account.DeleteEntity(&article); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении статьи"))
		return
	}

	resp := u.Message(true, "DELETE Article Successful")
	u.Respond(w, resp)
}
