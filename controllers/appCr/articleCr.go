package appCr

import (
	"encoding/json"
	"fmt"
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

	article, err := account.CreateArticle(input.Article)
	if err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка во время создания продукта"))
		return
	}

	resp := u.Message(true, "POST Article Created")
	resp["article"] = *article
	u.Respond(w, resp)
}

func ArticleGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	articleId, err := utilsCr.GetUINTVarFromRequest(r, "articleId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID статьи"))
		return
	}

	article, err := account.GetArticle(articleId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список магазинов"))
		return
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

	// 2. Узнаем, какой список нужен
	limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 100
	}
	offset, ok := utilsCr.GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	articles, total, err := account.GetArticleListPagination(offset, limit, search)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список статей"))
		return
	}

	resp := u.Message(true, "GET Article List Pagination")
	resp["total"] = total
	resp["articles"] = articles
	u.Respond(w, resp)
}

func ArticleUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		// fmt.Println(err)
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	articleId, err := utilsCr.GetUINTVarFromRequest(r, "articleId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID товара"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	//card, err := account.UpdateProduct(productId, input.Product)
	article, err := account.UpdateArticle(articleId, input)
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
		// fmt.Println(err)
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	articleId, err := utilsCr.GetUINTVarFromRequest(r, "articleId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID товара"))
		return
	}


	if err = account.DeleteArticle(articleId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении карточки товара"))
		return
	}

	resp := u.Message(true, "DELETE Article Successful")
	u.Respond(w, resp)
}
